/**
 * Plugin Sandbox Runtime
 * 
 * Manages Web Worker execution environment for plugins with security isolation
 * and resource limits as specified in ADR-013.
 */

import type { 
  Plugin, 
  PluginManifest, 
  PluginContext, 
  PluginModule, 
  Logger 
} from '../api.js';
import { PluginError } from '../api.js';
import { HostService } from './host-services.js';

/**
 * Plugin execution configuration.
 */
export interface SandboxConfig {
  timeoutMs: number;         // CPU timeout (default: 60000ms)
  memoryLimitMB: number;     // Memory limit (default: 256MB)
  allowedOrigins: string[];  // Allowed fetch origins for net permission
}

/**
 * Default sandbox configuration.
 */
export const DEFAULT_SANDBOX_CONFIG: SandboxConfig = {
  timeoutMs: 60000,      // 60 seconds
  memoryLimitMB: 256,    // 256MB
  allowedOrigins: []     // No network access by default
};

/**
 * Plugin execution result.
 */
export interface ExecutionResult<T = any> {
  success: boolean;
  result?: T;
  error?: PluginError;
  duration: number;
  memoryUsed?: number;
}

/**
 * Worker message types for plugin communication.
 */
export type WorkerMessage = 
  | { type: 'init'; manifest: PluginManifest; code: string; config: SandboxConfig }
  | { type: 'execute'; method: string; args: any[] }
  | { type: 'terminate' };

export type WorkerResponse =
  | { type: 'ready' }
  | { type: 'result'; data: any }
  | { type: 'error'; error: { message: string; code: string; stack?: string } }
  | { type: 'host-call'; method: string; args: any[]; callId: string }
  | { type: 'log'; level: string; message: string; args: any[] };

/**
 * Plugin sandbox that executes plugins in isolated Web Workers.
 */
export class PluginSandbox {
  private worker: Worker | null = null;
  private manifest: PluginManifest;
  private config: SandboxConfig;
  private hostService: HostService;
  private logger: Logger;
  private pendingCalls = new Map<string, { resolve: Function; reject: Function }>();
  private callIdCounter = 0;

  constructor(
    manifest: PluginManifest, 
    config: SandboxConfig = DEFAULT_SANDBOX_CONFIG
  ) {
    this.manifest = manifest;
    this.config = config;
    this.hostService = new HostService(manifest.permissions, manifest.id, manifest.name);
    this.logger = new SandboxLogger(manifest.id);
  }

  /**
   * Gets the host service (for accessing permission manager).
   */
  getHostService(): HostService {
    return this.hostService;
  }

  /**
   * Initializes the sandbox with plugin code.
   */
  async initialize(pluginCode: string): Promise<void> {
    if (this.worker) {
      await this.terminate();
    }

    return new Promise((resolve, reject) => {
      // Create worker from inline code
      const workerCode = this.createWorkerCode();
      const blob = new Blob([workerCode], { type: 'application/javascript' });
      const workerUrl = URL.createObjectURL(blob);

      this.worker = new Worker(workerUrl);
      
      this.worker.onmessage = (event) => {
        this.handleWorkerMessage(event.data);
      };

      this.worker.onerror = (error) => {
        this.logger.error('Worker error:', error);
        reject(new PluginError(
          `Worker initialization failed: ${error.message}`,
          'WORKER_INIT_FAILED',
          this.manifest.id
        ));
      };

      // Set up message handler for ready signal
      const handleReady = (event: MessageEvent) => {
        if (event.data.type === 'ready') {
          this.worker!.removeEventListener('message', handleReady);
          resolve();
        } else if (event.data.type === 'error') {
          reject(new PluginError(
            event.data.error.message,
            event.data.error.code,
            this.manifest.id
          ));
        }
      };

      this.worker.addEventListener('message', handleReady);

      // Initialize the worker
      this.worker.postMessage({
        type: 'init',
        manifest: this.manifest,
        code: pluginCode,
        config: this.config
      });

      // Cleanup blob URL
      URL.revokeObjectURL(workerUrl);

      // Set timeout for initialization
      setTimeout(() => {
        if (this.worker) {
          reject(new PluginError(
            'Plugin initialization timeout',
            'INIT_TIMEOUT',
            this.manifest.id
          ));
        }
      }, this.config.timeoutMs);
    });
  }

  /**
   * Executes a plugin method with the given arguments.
   */
  async execute<T = any>(method: string, ...args: any[]): Promise<ExecutionResult<T>> {
    if (!this.worker) {
      throw new PluginError('Sandbox not initialized', 'NOT_INITIALIZED', this.manifest.id);
    }

    const startTime = Date.now();

    return new Promise((resolve) => {
      const timeoutId = setTimeout(() => {
        this.terminate();
        resolve({
          success: false,
          error: new PluginError(
            'Plugin execution timeout',
            'EXECUTION_TIMEOUT',
            this.manifest.id
          ),
          duration: Date.now() - startTime
        });
      }, this.config.timeoutMs);

      const handleResult = (event: MessageEvent) => {
        const data = event.data;
        
        if (data.type === 'result') {
          clearTimeout(timeoutId);
          this.worker!.removeEventListener('message', handleResult);
          resolve({
            success: true,
            result: data.data,
            duration: Date.now() - startTime
          });
        } else if (data.type === 'error') {
          clearTimeout(timeoutId);
          this.worker!.removeEventListener('message', handleResult);
          resolve({
            success: false,
            error: new PluginError(
              data.error.message,
              data.error.code,
              this.manifest.id,
              { stack: data.error.stack }
            ),
            duration: Date.now() - startTime
          });
        }
      };

      this.worker!.addEventListener('message', handleResult);
      this.worker!.postMessage({
        type: 'execute',
        method,
        args
      });
    });
  }

  /**
   * Terminates the sandbox and cleans up resources.
   */
  async terminate(): Promise<void> {
    if (this.worker) {
      this.worker.postMessage({ type: 'terminate' });
      this.worker.terminate();
      this.worker = null;
    }

    // Reject any pending calls
    for (const [callId, { reject }] of this.pendingCalls) {
      reject(new PluginError('Sandbox terminated', 'SANDBOX_TERMINATED', this.manifest.id));
    }
    this.pendingCalls.clear();
  }

  /**
   * Handles messages from the worker.
   */
  private handleWorkerMessage(data: WorkerResponse): void {
    switch (data.type) {
      case 'host-call':
        this.handleHostCall(data.method, data.args, data.callId);
        break;
      case 'log':
        this.handleLogMessage(data.level, data.message, data.args);
        break;
    }
  }

  /**
   * Handles host service calls from the plugin.
   */
  private async handleHostCall(method: string, args: any[], callId: string): Promise<void> {
    try {
      const result = await this.hostService.call(method, args);
      this.worker?.postMessage({
        type: 'host-response',
        callId,
        result
      });
    } catch (error) {
      this.worker?.postMessage({
        type: 'host-error',
        callId,
        error: {
          message: error instanceof Error ? error.message : 'Unknown error',
          code: error instanceof PluginError ? error.code : 'HOST_CALL_FAILED'
        }
      });
    }
  }

  /**
   * Handles log messages from the plugin.
   */
  private handleLogMessage(level: string, message: string, args: any[]): void {
    switch (level) {
      case 'debug':
        this.logger.debug(message, ...args);
        break;
      case 'info':
        this.logger.info(message, ...args);
        break;
      case 'warn':
        this.logger.warn(message, ...args);
        break;
      case 'error':
        this.logger.error(message, ...args);
        break;
    }
  }

  /**
   * Creates the worker code that runs inside the Web Worker.
   */
  private createWorkerCode(): string {
    return `
// Plugin Worker Runtime
// This code runs inside a Web Worker to provide isolated execution

let manifest;
let config;
let pluginModule;
let hostCallId = 0;
const pendingHostCalls = new Map();

// Host proxy that forwards calls to the main thread
const hostProxy = new Proxy({}, {
  get(target, prop) {
    return (...args) => {
      const callId = (hostCallId++).toString();
      
      return new Promise((resolve, reject) => {
        pendingHostCalls.set(callId, { resolve, reject });
        
        self.postMessage({
          type: 'host-call',
          method: prop,
          args,
          callId
        });
        
        // Set timeout for host calls
        setTimeout(() => {
          if (pendingHostCalls.has(callId)) {
            pendingHostCalls.delete(callId);
            reject(new Error('Host call timeout'));
          }
        }, 30000); // 30 second timeout for host calls
      });
    };
  }
});

// Logger that forwards to main thread
const logger = {
  debug: (message, ...args) => self.postMessage({ type: 'log', level: 'debug', message, args }),
  info: (message, ...args) => self.postMessage({ type: 'log', level: 'info', message, args }),
  warn: (message, ...args) => self.postMessage({ type: 'log', level: 'warn', message, args }),
  error: (message, ...args) => self.postMessage({ type: 'log', level: 'error', message, args })
};

// Message handler
self.addEventListener('message', async (event) => {
  const { type, ...data } = event.data;
  
  try {
    switch (type) {
      case 'init':
        manifest = data.manifest;
        config = data.config;
        
        // Load plugin code
        try {
          // Create a function that provides the plugin context
          const pluginFunction = new Function(
            'host', 'logger', 'manifest',
            \`
            const context = { manifest, host, logger };
            \${data.code}
            
            // Plugin must export a register function
            if (typeof register !== 'function') {
              throw new Error('Plugin must export a register function');
            }
            
            return { register, events: typeof events !== 'undefined' ? events : undefined };
            \`
          );
          
          pluginModule = pluginFunction(hostProxy, logger, manifest);
          self.postMessage({ type: 'ready' });
        } catch (error) {
          self.postMessage({ 
            type: 'error', 
            error: { 
              message: error.message, 
              code: 'PLUGIN_LOAD_FAILED',
              stack: error.stack 
            } 
          });
        }
        break;
        
      case 'execute':
        if (!pluginModule) {
          throw new Error('Plugin not initialized');
        }
        
        const result = await executePluginMethod(data.method, data.args);
        self.postMessage({ type: 'result', data: result });
        break;
        
      case 'host-response':
        if (pendingHostCalls.has(data.callId)) {
          const { resolve } = pendingHostCalls.get(data.callId);
          pendingHostCalls.delete(data.callId);
          resolve(data.result);
        }
        break;
        
      case 'host-error':
        if (pendingHostCalls.has(data.callId)) {
          const { reject } = pendingHostCalls.get(data.callId);
          pendingHostCalls.delete(data.callId);
          const error = new Error(data.error.message);
          error.code = data.error.code;
          reject(error);
        }
        break;
        
      case 'terminate':
        self.close();
        break;
    }
  } catch (error) {
    self.postMessage({ 
      type: 'error', 
      error: { 
        message: error.message, 
        code: 'WORKER_ERROR',
        stack: error.stack 
      } 
    });
  }
});

async function executePluginMethod(method, args) {
  const context = { manifest, host: hostProxy, logger };
  
  if (method === 'register') {
    return await pluginModule.register(context);
  }
  
  throw new Error(\`Unknown method: \${method}\`);
}

// Error handler
self.addEventListener('error', (error) => {
  self.postMessage({ 
    type: 'error', 
    error: { 
      message: error.message, 
      code: 'WORKER_RUNTIME_ERROR',
      stack: error.error?.stack 
    } 
  });
});
`;
  }
}

/**
 * Logger implementation that forwards to the main thread.
 */
class SandboxLogger implements Logger {
  constructor(private pluginId: string) {}

  debug(message: string, ...args: any[]): void {
    console.debug(`[Plugin:${this.pluginId}]`, message, ...args);
  }

  info(message: string, ...args: any[]): void {
    console.info(`[Plugin:${this.pluginId}]`, message, ...args);
  }

  warn(message: string, ...args: any[]): void {
    console.warn(`[Plugin:${this.pluginId}]`, message, ...args);
  }

  error(message: string, ...args: any[]): void {
    console.error(`[Plugin:${this.pluginId}]`, message, ...args);
  }
}