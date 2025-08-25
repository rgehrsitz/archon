/**
 * Host Services Implementation
 * 
 * Provides secure, permission-gated access to Archon's APIs for plugins.
 * Each service method checks permissions before executing.
 */

import type { 
  Host, 
  Permission, 
  NodeId, 
  ArchonNode, 
  Mutation, 
  IndexDoc,
  UI,
  Secrets 
} from '../api.js';
import { PluginError } from '../api.js';

// Import Wails services
import * as PluginAPI from '../../../wailsjs/go/api/PluginService.js';
import * as NodeAPI from '../../api/nodes.js';
import * as GitAPI from '../../api/git.js';
import * as SnapshotAPI from '../../api/snapshots.js';
import { SearchAPI } from '../../api/search.js';
import { UIPermissionManager } from './ui-permission-manager.js';
import { UIService } from './ui-service.js';
import { SecretService } from './secret-service.js';

/**
 * Host service implementation that provides secure API access to plugins.
 */
export class HostService implements Host {
  private permissionManager: UIPermissionManager;
  private uiService: UIService;
  private secretService: SecretService;
  private pluginId: string;

  constructor(permissions: Permission[], pluginId: string, pluginName: string) {
    this.permissionManager = new UIPermissionManager(permissions, pluginId, pluginName);
    this.uiService = new UIService();
    this.secretService = new SecretService();
    this.pluginId = pluginId;
  }

  /**
   * Gets the UI permission manager for consent dialogs.
   */
  getPermissionManager(): UIPermissionManager {
    return this.permissionManager;
  }

  // ============================================================================
  // Repository Access
  // ============================================================================

  async getNode(id: NodeId): Promise<ArchonNode | null> {
    this.requirePermission('readRepo');
    
    try {
      const nodeData = await PluginAPI.PluginGetNode('', this.pluginId, id);
      return nodeData ? this.convertPluginNodeToArchonNode(nodeData) : null;
    } catch (error) {
      throw new PluginError(
        `Failed to get node: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'NODE_GET_FAILED',
        'host'
      );
    }
  }

  async listChildren(id: NodeId): Promise<NodeId[]> {
    this.requirePermission('readRepo');
    
    try {
      return await PluginAPI.PluginListChildren('', this.pluginId, id);
    } catch (error) {
      throw new PluginError(
        `Failed to list children: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'NODE_LIST_CHILDREN_FAILED',
        'host'
      );
    }
  }

  async query(selector: string): Promise<NodeId[]> {
    this.requirePermission('readRepo');
    
    try {
      const results = await PluginAPI.PluginQuery('', this.pluginId, selector, 100);
      return results.map(nodeData => nodeData.id || '');
    } catch (error) {
      throw new PluginError(
        `Failed to query nodes: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'NODE_QUERY_FAILED',
        'host'
      );
    }
  }

  async apply(edits: Mutation[]): Promise<void> {
    this.requirePermission('writeRepo');
    
    try {
      const pluginMutations = edits.map(edit => this.convertToPluginMutation(edit));
      await PluginAPI.PluginApplyMutations('', this.pluginId, pluginMutations);
    } catch (error) {
      throw new PluginError(
        `Failed to apply mutations: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'MUTATION_APPLY_FAILED',
        'host'
      );
    }
  }

  // ============================================================================
  // Git & Snapshots
  // ============================================================================

  async commit(message: string): Promise<string> {
    this.requirePermission('writeRepo');
    
    try {
      return await PluginAPI.PluginCommit('', this.pluginId, message);
    } catch (error) {
      throw new PluginError(
        `Failed to commit: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'COMMIT_FAILED',
        'host'
      );
    }
  }

  async snapshot(tag: string, notes?: Record<string, any>): Promise<string> {
    this.requirePermission('writeRepo');
    
    try {
      const message = notes?.description || `Snapshot: ${tag}`;
      return await PluginAPI.PluginSnapshot('', this.pluginId, message);
    } catch (error) {
      throw new PluginError(
        `Failed to create snapshot: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'SNAPSHOT_FAILED',
        'host'
      );
    }
  }

  // ============================================================================
  // Attachments
  // ============================================================================

  async readAttachment(hash: string): Promise<ArrayBuffer> {
    this.requirePermission('attachments');
    
    // Note: Attachment reading will be implemented when AttachmentProcessor plugins are added
    throw new PluginError('Attachment reading not yet implemented', 'NOT_IMPLEMENTED', 'host');
  }

  async writeAttachment(bytes: ArrayBuffer, filename?: string): Promise<{ hash: string, path: string }> {
    this.requirePermission('attachments');
    
    // Note: Attachment writing will be implemented when AttachmentProcessor plugins are added
    throw new PluginError('Attachment writing not yet implemented', 'NOT_IMPLEMENTED', 'host');
  }

  // ============================================================================
  // Network
  // ============================================================================

  async fetch(input: RequestInfo, init?: RequestInit): Promise<Response> {
    this.requirePermission('net');
    
    try {
      const url = typeof input === 'string' ? input : input.url;
      const method = init?.method || 'GET';
      const headers = init?.headers || {};
      const body = init?.body ? new Uint8Array(await new Response(init.body).arrayBuffer()) : undefined;
      
      const proxyReq = {
        method,
        url,
        headers: typeof headers === 'object' ? headers as Record<string, string> : {},
        body: body ? Array.from(body) : undefined,
        timeoutMs: 30000
      };
      
      const proxyResp = await PluginAPI.PluginNetRequest('', this.pluginId, proxyReq);
      
      return new Response(
        proxyResp.body ? new Uint8Array(proxyResp.body) : null,
        {
          status: proxyResp.status,
          headers: proxyResp.headers || {}
        }
      );
    } catch (error) {
      throw new PluginError(
        `Failed to fetch: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'FETCH_FAILED',
        'host'
      );
    }
  }

  // ============================================================================
  // Search Index
  // ============================================================================

  async indexPut(docs: IndexDoc[]): Promise<void> {
    this.requirePermission('indexWrite');
    
    try {
      // Index each document individually
      for (const doc of docs) {
        await PluginAPI.PluginIndexPut('', this.pluginId, doc.id, doc.content);
      }
    } catch (error) {
      throw new PluginError(
        `Failed to write to index: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'INDEX_WRITE_FAILED',
        'host'
      );
    }
  }

  // ============================================================================
  // UI Services
  // ============================================================================

  get ui(): UI {
    this.requirePermission('ui');
    return this.uiService;
  }

  // ============================================================================
  // Secrets
  // ============================================================================

  get secrets(): Secrets {
    // Permission check is done in SecretService based on secret name
    return this.secretService;
  }

  // ============================================================================
  // Internal Methods
  // ============================================================================

  /**
   * Calls a host service method by name (used by sandbox).
   */
  async call(method: string, args: any[]): Promise<any> {
    switch (method) {
      case 'getNode':
        return this.getNode(args[0]);
      case 'listChildren':
        return this.listChildren(args[0]);
      case 'query':
        return this.query(args[0]);
      case 'apply':
        return this.apply(args[0]);
      case 'commit':
        return this.commit(args[0]);
      case 'snapshot':
        return this.snapshot(args[0], args[1]);
      case 'readAttachment':
        return this.readAttachment(args[0]);
      case 'writeAttachment':
        return this.writeAttachment(args[0], args[1]);
      case 'fetch':
        return this.fetch(args[0], args[1]);
      case 'indexPut':
        return this.indexPut(args[0]);
      // UI methods
      case 'ui.registerCommand':
        return this.ui.registerCommand(args[0]);
      case 'ui.showPanel':
        return this.ui.showPanel(args[0]);
      case 'ui.showModal':
        return this.ui.showModal(args[0]);
      case 'ui.notify':
        return this.ui.notify(args[0]);
      // Secret methods
      case 'secrets.get':
        return this.secretsGet(args[0]);
      case 'secrets.set':
        return this.secretsSet(args[0], args[1], args[2]);
      case 'secrets.delete':
        return this.secretsDelete(args[0]);
      default:
        throw new PluginError(`Unknown host method: ${method}`, 'UNKNOWN_METHOD', 'host');
    }
  }

  /**
   * Checks if the required permission is available.
   */
  private requirePermission(permission: Permission): void {
    if (!this.permissionManager.hasPermission(permission)) {
      throw new PluginError(
        `Permission denied: ${permission}`,
        'PERMISSION_DENIED',
        'host'
      );
    }
  }

  /**
   * Converts a plugin NodeData to an ArchonNode.
   */
  private convertPluginNodeToArchonNode(nodeData: any): ArchonNode {
    return {
      id: nodeData.id || '',
      name: nodeData.name || '',
      description: nodeData.description || '',
      properties: nodeData.properties || {},
      children: nodeData.children || []
    };
  }

  /**
   * Converts an ArchonNode mutation to a plugin Mutation.
   */
  private convertToPluginMutation(mutation: Mutation): any {
    return {
      type: mutation.type,
      nodeId: mutation.nodeId,
      parentId: mutation.parentId,
      data: mutation.data ? {
        id: mutation.data.id,
        name: mutation.data.name,
        description: mutation.data.description,
        properties: mutation.data.properties,
        children: mutation.data.children
      } : null,
      position: mutation.position
    };
  }

  /**
   * Secrets methods using plugin API.
   */
  private async secretsGet(name: string): Promise<string | null> {
    try {
      const secret = await PluginAPI.PluginSecretsGet('', this.pluginId, name);
      return secret && !secret.redacted ? secret.value : null;
    } catch (error) {
      throw new PluginError(
        `Failed to get secret: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'SECRETS_GET_FAILED',
        'host'
      );
    }
  }

  private async secretsSet(name: string, value: string, opts?: { description?: string }): Promise<void> {
    // Note: secrets set functionality not yet implemented in backend
    throw new PluginError('Setting secrets not yet implemented', 'NOT_IMPLEMENTED', 'host');
  }

  private async secretsDelete(name: string): Promise<void> {
    // Note: secrets delete functionality not yet implemented in backend
    throw new PluginError('Deleting secrets not yet implemented', 'NOT_IMPLEMENTED', 'host');
  }

}