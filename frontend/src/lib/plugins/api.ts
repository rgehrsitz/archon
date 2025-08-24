/**
 * Archon Plugin System API
 * 
 * This file defines the comprehensive plugin API as specified in ADR-013.
 * It provides type-safe interfaces for 10 plugin types and host services.
 */

// ============================================================================
// Core Types
// ============================================================================

export type NodeId = string;
export type ArchonScalar = string | number | boolean | null;

/**
 * Represents a node in the Archon hierarchy.
 * This is the canonical node representation exposed to plugins.
 */
export interface ArchonNode {
  id: NodeId;
  name: string;
  description?: string;
  properties: Record<string, ArchonScalar>;
  children: NodeId[]; // order is meaningful
}

/**
 * A mutation operation that can be applied to the repository.
 */
export interface Mutation {
  type: 'create' | 'update' | 'delete' | 'move' | 'reorder';
  nodeId: NodeId;
  parentId?: NodeId;
  data?: Partial<ArchonNode>;
  position?: number;
}

/**
 * Plugin permissions that gate access to host services.
 */
export type Permission = 
  | 'readRepo'          // getNode, listChildren, query
  | 'writeRepo'         // apply, commit, snapshot
  | 'attachments'       // readAttachment, writeAttachment
  | 'net'               // host.fetch (scoped, rate-limited)
  | 'indexWrite'        // indexPut
  | 'ui'                // ui.* (commands, panels, modals, toasts)
  | `secrets:${string}` // scoped secret access (e.g., 'secrets:jira*')
  ;

/**
 * Plugin manifest metadata.
 */
/**
 * Plugin type names for categorization and validation.
 */
export type PluginType = 
  | 'Importer' 
  | 'Exporter' 
  | 'Transformer' 
  | 'Validator' 
  | 'Panel' 
  | 'Provider' 
  | 'AttachmentProcessor' 
  | 'ConflictResolver' 
  | 'SearchIndexer' 
  | 'UIContrib';

export interface PluginManifest {
  id: string;
  name: string;
  version: string;
  type: PluginType;
  description?: string;
  author?: string;
  license?: string;
  permissions: Permission[];
  entryPoint: string;
  archonVersion?: string; // Semantic version requirement
  integrity?: string; // SHA-256 hash for verification
  metadata?: {
    category?: string;
    tags?: string[];
    website?: string;
    repository?: string;
  };
}

// ============================================================================
// Host Services
// ============================================================================

/**
 * Main host interface that provides services to plugins.
 * All methods are gated by permissions declared in the plugin manifest.
 */
export interface Host {
  // Repository access
  getNode(id: NodeId): Promise<ArchonNode | null>;
  listChildren(id: NodeId): Promise<NodeId[]>;
  query(selector: string): Promise<NodeId[]>; // future: simple selectors
  apply(edits: Mutation[]): Promise<void>;    // gated by "writeRepo"
  
  // Git & snapshots
  commit(message: string): Promise<string>;   // returns commit sha
  snapshot(tag: string, notes?: Record<string, any>): Promise<string>;
  
  // Attachments
  readAttachment(hash: string): Promise<ArrayBuffer>;
  writeAttachment(bytes: ArrayBuffer, filename?: string): Promise<{ hash: string, path: string }>;
  
  // Network (only if permission granted)
  fetch(input: RequestInfo, init?: RequestInit): Promise<Response>;
  
  // Index (if granted)
  indexPut(docs: IndexDoc[]): Promise<void>;
  
  // UI services
  ui: UI;
  
  // Secrets (scoped; e.g., Jira token)
  secrets: Secrets;
}

/**
 * UI services for plugin interaction with the host interface.
 */
export interface UI {
  registerCommand(cmd: Command): void;
  showPanel(panel: PanelDescriptor): void;
  showModal(modal: ModalDescriptor): Promise<void>;
  notify(opts: { level: 'info' | 'warn' | 'error', message: string }): void;
}

/**
 * Secret management for external service authentication.
 */
export interface Secrets {
  get(name: string): Promise<string | null>;
  set(name: string, value: string, opts?: { description?: string }): Promise<void>;
  delete(name: string): Promise<void>;
}

/**
 * Command registration for UI contributions.
 */
export interface Command {
  id: string;
  title: string;
  category?: string;
  keybinding?: string;
  execute: () => Promise<void>;
}

/**
 * Panel descriptor for custom UI panels.
 */
export interface PanelDescriptor {
  id: string;
  title: string;
  component: any; // Svelte component
  props?: Record<string, any>;
}

/**
 * Modal descriptor for dialog interactions.
 */
export interface ModalDescriptor {
  id: string;
  title: string;
  component: any; // Svelte component
  props?: Record<string, any>;
  size?: 'sm' | 'md' | 'lg' | 'xl';
}

/**
 * Index document for search indexing.
 */
export interface IndexDoc {
  id: string;
  content: string;
  metadata?: Record<string, any>;
}

// ============================================================================
// Plugin Types (10 categories)
// ============================================================================

/**
 * Base interface for all plugin types.
 */
export interface BasePlugin {
  kind: string;
  meta: {
    id: string;
    name: string;
    version: string;
    description?: string;
  };
}

/**
 * Import plugins create nodes from external data sources.
 */
export interface Importer extends BasePlugin {
  kind: 'importer';
  meta: BasePlugin['meta'] & {
    formats: string[]; // supported file formats/MIME types
  };
  run(input: Uint8Array | string, options?: Record<string, any>): Promise<{ root: ArchonNode }>;
}

/**
 * Export plugins serialize nodes/subtrees to external formats.
 */
export interface Exporter extends BasePlugin {
  kind: 'exporter';
  meta: BasePlugin['meta'] & {
    formats: string[]; // output formats
  };
  run(root: NodeId, options?: Record<string, any>): Promise<Uint8Array | string>;
}

/**
 * Transformer plugins perform bulk edits and data transformations.
 */
export interface Transformer extends BasePlugin {
  kind: 'transformer';
  run(scope: NodeId | NodeId[], options?: Record<string, any>): Promise<Mutation[]>;
}

/**
 * Validator plugins perform read-only checks and can block operations.
 */
export interface Validator extends BasePlugin {
  kind: 'validator';
  run(scope: NodeId | NodeId[], options?: Record<string, any>): Promise<ValidationReport>;
}

/**
 * Panel plugins provide custom UI components.
 */
export interface Panel extends BasePlugin {
  kind: 'panel';
  mount(host: Host, props: { root: NodeId }): PanelHandle;
}

/**
 * Provider plugins connect to external systems (CMDB, ERP, Jira).
 */
export interface Provider extends BasePlugin {
  kind: 'provider';
  meta: BasePlugin['meta'] & {
    services: string[]; // e.g., ["jira", "confluence"]
  };
  configure(config: Record<string, any>): Promise<void>;
  pull(options?: Record<string, any>): Promise<ProviderPullResult>;
  push(options?: Record<string, any>): Promise<ProviderPushResult>;
}

/**
 * AttachmentProcessor plugins analyze and transform file attachments.
 */
export interface AttachmentProcessor extends BasePlugin {
  kind: 'attachment-processor';
  meta: BasePlugin['meta'] & {
    types: string[]; // supported MIME types
  };
  run(file: ArrayBuffer, filename?: string): Promise<AttachmentAnalysis>;
}

/**
 * ConflictResolver plugins provide custom semantic merge strategies.
 */
export interface ConflictResolver extends BasePlugin {
  kind: 'conflict-resolver';
  meta: BasePlugin['meta'] & {
    fields?: string[]; // specific field types this resolver handles
  };
  resolve(base: ArchonNode, ours: ArchonNode, theirs: ArchonNode): Promise<ArchonNode | 'conflict'>;
}

/**
 * SearchIndexer plugins add domain-specific tokenization to the search index.
 */
export interface SearchIndexer extends BasePlugin {
  kind: 'search-indexer';
  index(scope: NodeId | NodeId[], emit: (doc: IndexDoc) => void): Promise<void>;
}

/**
 * UIContrib plugins contribute commands, menus, and contextual UI elements.
 */
export interface UIContrib extends BasePlugin {
  kind: 'ui-contrib';
  contribute(ui: UI): Promise<void>;
}

/**
 * Union type of all plugin types.
 */
export type Plugin = 
  | Importer 
  | Exporter 
  | Transformer 
  | Validator 
  | Panel 
  | Provider 
  | AttachmentProcessor 
  | ConflictResolver 
  | SearchIndexer 
  | UIContrib;

// ============================================================================
// Supporting Types
// ============================================================================

/**
 * Validation report with errors and warnings.
 */
export interface ValidationReport {
  valid: boolean;
  errors: ValidationError[];
  warnings: ValidationWarning[];
}

export interface ValidationError {
  code: string;
  message: string;
  nodeId?: NodeId;
  field?: string;
}

export interface ValidationWarning {
  code: string;
  message: string;
  nodeId?: NodeId;
  field?: string;
}

/**
 * Provider pull result.
 */
export interface ProviderPullResult {
  nodes: ArchonNode[];
  metadata?: Record<string, any>;
  warnings?: string[];
}

/**
 * Provider push result.
 */
export interface ProviderPushResult {
  success: boolean;
  created?: string[];
  updated?: string[];
  errors?: string[];
}

/**
 * Attachment analysis result.
 */
export interface AttachmentAnalysis {
  metadata: Record<string, any>;
  extractedText?: string;
  thumbnail?: ArrayBuffer;
  warnings?: string[];
}

/**
 * Panel handle for managing mounted panels.
 */
export interface PanelHandle {
  unmount(): void;
  update(props: Record<string, any>): void;
}

// ============================================================================
// Event System
// ============================================================================

/**
 * Plugin lifecycle events that plugins can subscribe to.
 */
export interface Events {
  onBeforeCommit?(ctx: CommitContext): Promise<void | 'block'>;
  onAfterCommit?(ctx: CommitContext): Promise<void>;
  onBeforeSnapshot?(tag: string): Promise<void | 'block'>;
  onAfterSnapshot?(tag: string): Promise<void>;
  onPull?(result: PullResult): Promise<void>;
  onMergeStart?(ctx: MergeContext): Promise<void>;
  onMergeEnd?(ctx: MergeResult): Promise<void>;
}

export interface CommitContext {
  message: string;
  changes: Mutation[];
  author?: string;
}

export interface PullResult {
  from: string;
  to: string;
  changes: string[];
}

export interface MergeContext {
  base: string;
  ours: string;
  theirs: string;
}

export interface MergeResult {
  success: boolean;
  conflicts?: string[];
  applied?: Mutation[];
}

// ============================================================================
// Plugin Runtime
// ============================================================================

/**
 * Plugin execution context provided to plugins at runtime.
 */
export interface PluginContext {
  manifest: PluginManifest;
  host: Host;
  logger: Logger;
}

/**
 * Logger interface for plugin debugging and monitoring.
 */
export interface Logger {
  debug(message: string, ...args: any[]): void;
  info(message: string, ...args: any[]): void;
  warn(message: string, ...args: any[]): void;
  error(message: string, ...args: any[]): void;
}

/**
 * Plugin execution error.
 */
export class PluginError extends Error {
  constructor(
    message: string,
    public code: string,
    public pluginId: string,
    public details?: Record<string, any>
  ) {
    super(message);
    this.name = 'PluginError';
  }
}

// ============================================================================
// Runtime Registration
// ============================================================================

/**
 * Plugin registration function that plugins must export.
 */
export type PluginRegistration = (context: PluginContext) => Promise<Plugin | Plugin[]>;

/**
 * Plugin module interface that all plugins must implement.
 */
export interface PluginModule {
  register: PluginRegistration;
  events?: Events;
}