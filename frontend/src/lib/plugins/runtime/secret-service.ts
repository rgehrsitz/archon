/**
 * Secret Service Implementation
 * 
 * Provides secure secret storage and retrieval for plugins.
 */

import type { Secrets } from '../api.js';
import { PluginError } from '../api.js';

/**
 * Secret metadata.
 */
export interface SecretMetadata {
  name: string;
  description?: string;
  createdAt: Date;
  updatedAt: Date;
  scope: string; // e.g., 'jira', 'general'
}

/**
 * Secret storage entry.
 */
interface SecretEntry {
  value: string;
  metadata: SecretMetadata;
}

/**
 * Secret service implementation with scoped access control.
 */
export class SecretService implements Secrets {
  private secrets = new Map<string, SecretEntry>();
  
  // In a real implementation, this would use secure storage
  // like the browser's secure storage API or delegate to the Go backend
  
  /**
   * Retrieves a secret value by name.
   */
  async get(name: string): Promise<string | null> {
    this.validateSecretAccess(name);
    
    const entry = this.secrets.get(name);
    return entry ? entry.value : null;
  }

  /**
   * Stores a secret value.
   */
  async set(name: string, value: string, opts?: { description?: string }): Promise<void> {
    this.validateSecretAccess(name);
    
    const existing = this.secrets.get(name);
    const now = new Date();
    
    const metadata: SecretMetadata = {
      name,
      description: opts?.description,
      createdAt: existing?.metadata.createdAt || now,
      updatedAt: now,
      scope: this.extractScope(name)
    };

    this.secrets.set(name, {
      value,
      metadata
    });
  }

  /**
   * Deletes a secret.
   */
  async delete(name: string): Promise<void> {
    this.validateSecretAccess(name);
    
    this.secrets.delete(name);
  }

  /**
   * Lists secret metadata (values are not included for security).
   */
  async listSecrets(scope?: string): Promise<SecretMetadata[]> {
    const results: SecretMetadata[] = [];
    
    for (const [name, entry] of this.secrets) {
      // Check scope filter
      if (scope && entry.metadata.scope !== scope) {
        continue;
      }
      
      // Check access permissions
      try {
        this.validateSecretAccess(name);
        results.push(entry.metadata);
      } catch {
        // Skip secrets we don't have access to
        continue;
      }
    }
    
    return results;
  }

  /**
   * Validates that the current context can access a secret.
   */
  private validateSecretAccess(name: string): void {
    const scope = this.extractScope(name);
    
    // TODO: Check against plugin permissions
    // For now, we'll implement basic scope validation
    
    // This would normally check against the plugin's declared permissions
    // e.g., if plugin has 'secrets:jira*', they can access jira.* secrets
    console.log(`Secret access requested for: ${name} (scope: ${scope})`);
  }

  /**
   * Extracts the scope from a secret name.
   */
  private extractScope(name: string): string {
    const parts = name.split('.');
    return parts.length > 1 ? parts[0] : 'general';
  }
}

/**
 * Predefined secret name patterns for common services.
 */
export const COMMON_SECRET_PATTERNS = {
  // Jira secrets
  JIRA_TOKEN: 'jira.token',
  JIRA_OAUTH_TOKEN: 'jira.oauth.token',
  JIRA_OAUTH_REFRESH: 'jira.oauth.refresh',
  JIRA_BASE_URL: 'jira.baseUrl',
  
  // Generic OAuth patterns
  OAUTH_CLIENT_ID: (service: string) => `${service}.oauth.clientId`,
  OAUTH_CLIENT_SECRET: (service: string) => `${service}.oauth.clientSecret`,
  OAUTH_ACCESS_TOKEN: (service: string) => `${service}.oauth.accessToken`,
  OAUTH_REFRESH_TOKEN: (service: string) => `${service}.oauth.refreshToken`,
  
  // API keys
  API_KEY: (service: string) => `${service}.apiKey`,
  API_SECRET: (service: string) => `${service}.apiSecret`,
};

/**
 * Creates a scoped secret name.
 */
export function createSecretName(scope: string, key: string): string {
  return `${scope}.${key}`;
}

/**
 * Validates a secret name format.
 */
export function isValidSecretName(name: string): boolean {
  // Allow alphanumeric characters, dots, dashes, and underscores
  // Must contain at least one dot for scoping
  return /^[a-zA-Z0-9][a-zA-Z0-9._-]*\.[a-zA-Z0-9._-]+$/.test(name);
}

/**
 * Secret configuration helper for common services.
 */
export class SecretConfigHelper {
  constructor(private secretService: SecretService) {}

  /**
   * Configures Jira authentication secrets.
   */
  async configureJira(config: {
    baseUrl: string;
    authType: 'token' | 'oauth';
    token?: string;
    oauthConfig?: {
      clientId: string;
      clientSecret: string;
      accessToken: string;
      refreshToken?: string;
    };
  }): Promise<void> {
    await this.secretService.set(COMMON_SECRET_PATTERNS.JIRA_BASE_URL, config.baseUrl, {
      description: 'Jira base URL'
    });

    if (config.authType === 'token' && config.token) {
      await this.secretService.set(COMMON_SECRET_PATTERNS.JIRA_TOKEN, config.token, {
        description: 'Jira API token'
      });
    } else if (config.authType === 'oauth' && config.oauthConfig) {
      const oauth = config.oauthConfig;
      
      await this.secretService.set(COMMON_SECRET_PATTERNS.OAUTH_CLIENT_ID('jira'), oauth.clientId, {
        description: 'Jira OAuth client ID'
      });
      
      await this.secretService.set(COMMON_SECRET_PATTERNS.OAUTH_CLIENT_SECRET('jira'), oauth.clientSecret, {
        description: 'Jira OAuth client secret'
      });
      
      await this.secretService.set(COMMON_SECRET_PATTERNS.OAUTH_ACCESS_TOKEN('jira'), oauth.accessToken, {
        description: 'Jira OAuth access token'
      });
      
      if (oauth.refreshToken) {
        await this.secretService.set(COMMON_SECRET_PATTERNS.OAUTH_REFRESH_TOKEN('jira'), oauth.refreshToken, {
          description: 'Jira OAuth refresh token'
        });
      }
    }
  }

  /**
   * Gets Jira configuration from secrets.
   */
  async getJiraConfig(): Promise<{
    baseUrl?: string;
    token?: string;
    oauthConfig?: {
      clientId: string;
      clientSecret: string;
      accessToken: string;
      refreshToken?: string;
    };
  }> {
    const baseUrl = await this.secretService.get(COMMON_SECRET_PATTERNS.JIRA_BASE_URL);
    const token = await this.secretService.get(COMMON_SECRET_PATTERNS.JIRA_TOKEN);
    
    const oauthClientId = await this.secretService.get(COMMON_SECRET_PATTERNS.OAUTH_CLIENT_ID('jira'));
    const oauthClientSecret = await this.secretService.get(COMMON_SECRET_PATTERNS.OAUTH_CLIENT_SECRET('jira'));
    const oauthAccessToken = await this.secretService.get(COMMON_SECRET_PATTERNS.OAUTH_ACCESS_TOKEN('jira'));
    const oauthRefreshToken = await this.secretService.get(COMMON_SECRET_PATTERNS.OAUTH_REFRESH_TOKEN('jira'));

    const oauthConfig = (oauthClientId && oauthClientSecret && oauthAccessToken) ? {
      clientId: oauthClientId,
      clientSecret: oauthClientSecret,
      accessToken: oauthAccessToken,
      refreshToken: oauthRefreshToken || undefined
    } : undefined;

    return {
      baseUrl: baseUrl || undefined,
      token: token || undefined,
      oauthConfig
    };
  }
}