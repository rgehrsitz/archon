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

  constructor(permissions: Permission[], pluginId: string, pluginName: string) {
    this.permissionManager = new UIPermissionManager(permissions, pluginId, pluginName);
    this.uiService = new UIService();
    this.secretService = new SecretService();
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
      const node = await NodeAPI.getNode(id);
      return node ? this.convertToArchonNode(node) : null;
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
      const children = await NodeAPI.listChildren(id);
      return children.map(child => child.id);
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
      // Basic query implementation using search API
      // Treat selectors as search queries for now
      const results = await SearchAPI.searchNodes(selector, 100);
      return results.results.map(result => result.nodeId);
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
      // Apply each mutation sequentially
      for (const edit of edits) {
        await this.applyMutation(edit);
      }
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
      // Note: Git commit API may need to be implemented in Wails backend
      // For now, return a placeholder
      throw new PluginError('Git commit not yet implemented in backend', 'NOT_IMPLEMENTED', 'host');
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
      const result = await SnapshotAPI.createSnapshot(tag, notes?.description || '', notes || {});
      return result.hash || tag;
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
    
    try {
      // TODO: Implement attachment reading
      // This would call to the Go backend to read attachment by hash
      throw new PluginError('Attachment reading not yet implemented', 'NOT_IMPLEMENTED', 'host');
    } catch (error) {
      throw new PluginError(
        `Failed to read attachment: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'ATTACHMENT_READ_FAILED',
        'host'
      );
    }
  }

  async writeAttachment(bytes: ArrayBuffer, filename?: string): Promise<{ hash: string, path: string }> {
    this.requirePermission('attachments');
    
    try {
      // TODO: Implement attachment writing
      // This would call to the Go backend to store the attachment
      throw new PluginError('Attachment writing not yet implemented', 'NOT_IMPLEMENTED', 'host');
    } catch (error) {
      throw new PluginError(
        `Failed to write attachment: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'ATTACHMENT_WRITE_FAILED',
        'host'
      );
    }
  }

  // ============================================================================
  // Network
  // ============================================================================

  async fetch(input: RequestInfo, init?: RequestInit): Promise<Response> {
    this.requirePermission('net');
    
    // TODO: Implement network access with origin validation and rate limiting
    // For now, this is a placeholder
    throw new PluginError('Network access not yet implemented', 'NOT_IMPLEMENTED', 'host');
  }

  // ============================================================================
  // Search Index
  // ============================================================================

  async indexPut(docs: IndexDoc[]): Promise<void> {
    this.requirePermission('indexWrite');
    
    try {
      // TODO: Implement index writing
      // This would call to the Go backend to add documents to the search index
      throw new PluginError('Index writing not yet implemented', 'NOT_IMPLEMENTED', 'host');
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
        return this.secrets.get(args[0]);
      case 'secrets.set':
        return this.secrets.set(args[0], args[1], args[2]);
      case 'secrets.delete':
        return this.secrets.delete(args[0]);
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
   * Converts a Wails node model to an ArchonNode.
   */
  private convertToArchonNode(node: NodeAPI.Node): ArchonNode {
    // Convert properties from Wails format to plugin format
    const properties: Record<string, any> = {};
    
    if (node.properties) {
      for (const [key, prop] of Object.entries(node.properties)) {
        properties[key] = prop.value;
      }
    }

    return {
      id: node.id,
      name: node.name,
      description: node.description || '',
      properties,
      children: node.children || []
    };
  }

  /**
   * Applies a single mutation to the repository.
   */
  private async applyMutation(mutation: Mutation): Promise<void> {
    switch (mutation.type) {
      case 'create':
        if (!mutation.parentId || !mutation.data) {
          throw new PluginError('Create mutation requires parentId and data', 'INVALID_MUTATION', 'host');
        }
        await NodeAPI.createNode({
          parentId: mutation.parentId,
          name: mutation.data.name || 'Untitled',
          description: mutation.data.description || '',
          properties: this.convertPropertiesToWailsFormat(mutation.data.properties || {})
        });
        break;

      case 'update':
        if (!mutation.data) {
          throw new PluginError('Update mutation requires data', 'INVALID_MUTATION', 'host');
        }
        await NodeAPI.updateNode({
          id: mutation.nodeId,
          name: mutation.data.name,
          description: mutation.data.description,
          properties: mutation.data.properties 
            ? this.convertPropertiesToWailsFormat(mutation.data.properties)
            : undefined
        });
        break;

      case 'delete':
        await NodeAPI.deleteNode(mutation.nodeId);
        break;

      case 'move':
        if (!mutation.parentId) {
          throw new PluginError('Move mutation requires parentId', 'INVALID_MUTATION', 'host');
        }
        await NodeAPI.moveNode({
          nodeId: mutation.nodeId,
          newParentId: mutation.parentId,
          position: mutation.position
        });
        break;

      case 'reorder':
        if (!mutation.parentId || !mutation.data?.children) {
          throw new PluginError('Reorder mutation requires parentId and children array', 'INVALID_MUTATION', 'host');
        }
        await NodeAPI.reorderChildren({
          parentId: mutation.parentId,
          orderedChildIds: mutation.data.children
        });
        break;

      default:
        throw new PluginError(`Unknown mutation type: ${(mutation as any).type}`, 'INVALID_MUTATION', 'host');
    }
  }

  /**
   * Converts plugin properties format to Wails properties format.
   */
  private convertPropertiesToWailsFormat(properties: Record<string, any>): Record<string, any> {
    const converted: Record<string, any> = {};
    
    for (const [key, value] of Object.entries(properties)) {
      converted[key] = {
        value,
        typeHint: this.inferTypeHint(value)
      };
    }
    
    return converted;
  }

  /**
   * Infers type hint from a value.
   */
  private inferTypeHint(value: any): string {
    if (typeof value === 'string') return 'string';
    if (typeof value === 'number') return 'number';
    if (typeof value === 'boolean') return 'boolean';
    if (value === null) return 'string';
    if (value instanceof Date) return 'date';
    return 'string'; // default fallback
  }
}