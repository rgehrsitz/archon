/**
 * UI Permission Manager
 * 
 * Manages permission consent dialogs and user interaction for the plugin system.
 * Integrates with the PermissionManager to provide UI-based consent flows.
 */

import { writable, type Writable } from 'svelte/store';
import { 
  PermissionManager, 
  type PermissionRequest, 
  type PermissionRequestCallback 
} from './permissions.js';
import type { Permission } from '../api.js';

export interface PermissionConsentState {
  isOpen: boolean;
  request: PermissionRequest | null;
  pluginName: string;
  pluginId: string;
  resolve: ((granted: boolean) => void) | null;
}

/**
 * UI-integrated permission manager that handles consent dialogs.
 */
export class UIPermissionManager extends PermissionManager {
  private consentState: Writable<PermissionConsentState> = writable({
    isOpen: false,
    request: null,
    pluginName: '',
    pluginId: '',
    resolve: null
  });

  private pluginName: string;
  private pluginId: string;

  constructor(
    declaredPermissions: Permission[],
    pluginId: string,
    pluginName: string
  ) {
    super(declaredPermissions);
    this.pluginId = pluginId;
    this.pluginName = pluginName;
    
    // Register the UI consent callback
    this.onPermissionRequest(this.handleUIPermissionRequest.bind(this));
  }

  /**
   * Gets the consent dialog state store.
   */
  getConsentState(): Writable<PermissionConsentState> {
    return this.consentState;
  }

  /**
   * Handles consent dialog result from UI.
   */
  handleConsentResult(granted: boolean, options: { temporary?: boolean; duration?: number } = {}): void {
    this.consentState.update(state => {
      if (state.resolve && state.request) {
        if (granted) {
          this.grantPermission(state.request.permission, {
            temporary: options.temporary,
            duration: options.duration
          });
        }
        state.resolve(granted);
      }
      
      return {
        isOpen: false,
        request: null,
        pluginName: '',
        pluginId: '',
        resolve: null
      };
    });
  }

  /**
   * Shows permission consent dialog and waits for user response.
   */
  private handleUIPermissionRequest(request: PermissionRequest): Promise<boolean> {
    return new Promise((resolve) => {
      this.consentState.set({
        isOpen: true,
        request,
        pluginName: this.pluginName,
        pluginId: this.pluginId,
        resolve
      });
    });
  }

  /**
   * Closes any open consent dialog.
   */
  closeConsentDialog(): void {
    this.consentState.update(state => {
      if (state.resolve) {
        state.resolve(false);
      }
      return {
        isOpen: false,
        request: null,
        pluginName: '',
        pluginId: '',
        resolve: null
      };
    });
  }

  /**
   * Pre-grants permissions without showing dialog (for development/testing).
   */
  preGrantPermissions(permissions: Permission[]): void {
    for (const permission of permissions) {
      if (this.getDeclaredPermissions().includes(permission)) {
        this.grantPermission(permission);
      }
    }
  }

  /**
   * Revokes all granted permissions and shows dialog for next request.
   */
  resetAllPermissions(): void {
    for (const permission of this.getDeclaredPermissions()) {
      this.revokePermission(permission);
    }
    this.closeConsentDialog();
  }

  /**
   * Gets a summary of permission status for UI display.
   */
  getPermissionSummary(): {
    total: number;
    granted: number;
    pending: number;
    denied: number;
    details: Array<{
      permission: Permission;
      granted: boolean;
      temporary: boolean;
      expiresAt?: Date;
    }>;
  } {
    const declared = this.getDeclaredPermissions();
    const granted = this.getGrantedPermissions();
    
    const details = declared.map(permission => {
      const grant = granted.find(g => g.permission === permission);
      return {
        permission,
        granted: !!grant?.granted,
        temporary: !!grant?.temporary,
        expiresAt: grant?.expiresAt
      };
    });

    return {
      total: declared.length,
      granted: granted.length,
      pending: 0, // We don't track pending state in this simple implementation
      denied: declared.length - granted.length,
      details
    };
  }
}