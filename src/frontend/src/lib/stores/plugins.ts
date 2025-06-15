export interface Plugin {
  id: string;
  name: string;
  version: string;
  description: string;
  author: string;
  enabled: boolean;
  config: Record<string, any>;
}

export class PluginStore {
  $state = {
    plugins: [] as Plugin[],
    loading: false,
    error: null as string | null,
  };

  // Load plugins from backend
  async loadPlugins() {
    this.$state.loading = true;
    this.$state.error = null;
    try {
      const response = await fetch("/api/plugins");
      if (!response.ok) throw new Error("Failed to load plugins");
      const plugins = await response.json();
      this.$state.plugins = plugins;
    } catch (error) {
      this.$state.error =
        error instanceof Error ? error.message : "Failed to load plugins";
    } finally {
      this.$state.loading = false;
    }
  }

  // Install a new plugin
  async installPlugin(pluginUrl: string) {
    this.$state.loading = true;
    this.$state.error = null;
    try {
      const response = await fetch("/api/plugins/install", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url: pluginUrl }),
      });
      if (!response.ok) throw new Error("Failed to install plugin");
      const newPlugin = await response.json();
      this.$state.plugins = [...this.$state.plugins, newPlugin];
      return newPlugin;
    } catch (error) {
      this.$state.error =
        error instanceof Error ? error.message : "Failed to install plugin";
      throw error;
    } finally {
      this.$state.loading = false;
    }
  }

  // Update plugin configuration
  async updatePluginConfig(id: string, config: Record<string, any>) {
    this.$state.loading = true;
    this.$state.error = null;
    try {
      const response = await fetch(`/api/plugins/${id}/config`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ config }),
      });
      if (!response.ok) throw new Error("Failed to update plugin config");
      const updatedPlugin = await response.json();
      this.$state.plugins = this.$state.plugins.map((p) =>
        p.id === id ? updatedPlugin : p
      );
      return updatedPlugin;
    } catch (error) {
      this.$state.error =
        error instanceof Error
          ? error.message
          : "Failed to update plugin config";
      throw error;
    } finally {
      this.$state.loading = false;
    }
  }

  // Toggle plugin enabled state
  async togglePlugin(id: string, enabled: boolean) {
    this.$state.loading = true;
    this.$state.error = null;
    try {
      const response = await fetch(`/api/plugins/${id}/toggle`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ enabled }),
      });
      if (!response.ok) throw new Error("Failed to toggle plugin");
      const updatedPlugin = await response.json();
      this.$state.plugins = this.$state.plugins.map((p) =>
        p.id === id ? updatedPlugin : p
      );
      return updatedPlugin;
    } catch (error) {
      this.$state.error =
        error instanceof Error ? error.message : "Failed to toggle plugin";
      throw error;
    } finally {
      this.$state.loading = false;
    }
  }

  // Uninstall a plugin
  async uninstallPlugin(id: string) {
    this.$state.loading = true;
    this.$state.error = null;
    try {
      const response = await fetch(`/api/plugins/${id}`, {
        method: "DELETE",
      });
      if (!response.ok) throw new Error("Failed to uninstall plugin");
      this.$state.plugins = this.$state.plugins.filter((p) => p.id !== id);
    } catch (error) {
      this.$state.error =
        error instanceof Error ? error.message : "Failed to uninstall plugin";
      throw error;
    } finally {
      this.$state.loading = false;
    }
  }
}

export const plugins = new PluginStore();
