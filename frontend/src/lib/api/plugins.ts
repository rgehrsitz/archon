// Plugin API wrapper for frontend
export interface PluginManifest {
  id: string;
  name: string;
  version: string;
  type: string;
  description?: string;
  author?: string;
  license?: string;
  permissions: string[];
  entryPoint: string;
  archonVersion?: string;
  integrity?: string;
}

export interface PluginInstallation {
  manifest: PluginManifest;
  path: string;
  installedAt: string; // ISO string instead of time.Time
  enabled: boolean;
  source: string;
}

export interface PluginPermissionGrant {
  pluginId: string;
  permission: string;
  granted: boolean;
  temporary: boolean;
  expiresAt?: string; // ISO string instead of time.Time
  grantedAt: string; // ISO string instead of time.Time
}

// Plugin API functions using direct Wails calls
export async function getPlugins(): Promise<PluginInstallation[]> {
  try {
    const result = await (window as any).go.api.PluginService.GetPlugins();
    if (result.Code) {
      throw new Error(`${result.Code}: ${result.Message}`);
    }
    return result;
  } catch (error) {
    console.error('Failed to get plugins:', error);
    return [];
  }
}

export async function getEnabledPlugins(): Promise<PluginInstallation[]> {
  try {
    const result = await (window as any).go.api.PluginService.GetEnabledPlugins();
    if (result.Code) {
      throw new Error(`${result.Code}: ${result.Message}`);
    }
    return result;
  } catch (error) {
    console.error('Failed to get enabled plugins:', error);
    return [];
  }
}

export async function enablePlugin(pluginId: string): Promise<void> {
  const result = await (window as any).go.api.PluginService.EnablePlugin(pluginId);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
}

export async function disablePlugin(pluginId: string): Promise<void> {
  const result = await (window as any).go.api.PluginService.DisablePlugin(pluginId);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
}

export async function getPluginPermissions(pluginId: string): Promise<PluginPermissionGrant[]> {
  try {
    const result = await (window as any).go.api.PluginService.GetPluginPermissions(pluginId);
    if (result.Code) {
      throw new Error(`${result.Code}: ${result.Message}`);
    }
    return result;
  } catch (error) {
    console.error('Failed to get plugin permissions:', error);
    return [];
  }
}

export async function grantPermission(pluginId: string, permission: string, temporary: boolean = false, durationMs: number = 0): Promise<void> {
  const result = await (window as any).go.api.PluginService.GrantPermission(pluginId, permission, temporary, durationMs);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
}

export async function revokePermission(pluginId: string, permission: string): Promise<void> {
  const result = await (window as any).go.api.PluginService.RevokePermission(pluginId, permission);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
}

export async function installPlugin(sourcePath: string): Promise<PluginInstallation> {
  const result = await (window as any).go.api.PluginService.InstallPlugin(sourcePath);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
  return result;
}

export async function uninstallPlugin(pluginId: string): Promise<void> {
  const result = await (window as any).go.api.PluginService.UninstallPlugin(pluginId);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
}
