export interface Component {
  id: string;
  name: string;
  type: string;
  config: Record<string, any>;
  parentId: string | null;
}

export class ComponentStore {
  $state = {
    components: [] as Component[],
    loading: false,
    error: null as string | null,
  };

  // Load components from backend
  async loadComponents() {
    this.$state.loading = true;
    this.$state.error = null;
    try {
      const response = await fetch("/api/components");
      if (!response.ok) throw new Error("Failed to load components");
      const components = await response.json();
      this.$state.components = components;
    } catch (error) {
      this.$state.error =
        error instanceof Error ? error.message : "Failed to load components";
    } finally {
      this.$state.loading = false;
    }
  }

  // Add a new component
  async addComponent(component: Omit<Component, "id">) {
    this.$state.loading = true;
    this.$state.error = null;
    try {
      const response = await fetch("/api/components", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(component),
      });
      if (!response.ok) throw new Error("Failed to add component");
      const newComponent = await response.json();
      this.$state.components = [...this.$state.components, newComponent];
      return newComponent;
    } catch (error) {
      this.$state.error =
        error instanceof Error ? error.message : "Failed to add component";
      throw error;
    } finally {
      this.$state.loading = false;
    }
  }

  // Update an existing component
  async updateComponent(id: string, updates: Partial<Component>) {
    this.$state.loading = true;
    this.$state.error = null;
    try {
      const response = await fetch(`/api/components/${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(updates),
      });
      if (!response.ok) throw new Error("Failed to update component");
      const updatedComponent = await response.json();
      this.$state.components = this.$state.components.map((c) =>
        c.id === id ? updatedComponent : c
      );
      return updatedComponent;
    } catch (error) {
      this.$state.error =
        error instanceof Error ? error.message : "Failed to update component";
      throw error;
    } finally {
      this.$state.loading = false;
    }
  }

  // Delete a component
  async deleteComponent(id: string) {
    this.$state.loading = true;
    this.$state.error = null;
    try {
      const response = await fetch(`/api/components/${id}`, {
        method: "DELETE",
      });
      if (!response.ok) throw new Error("Failed to delete component");
      this.$state.components = this.$state.components.filter(
        (c) => c.id !== id
      );
    } catch (error) {
      this.$state.error =
        error instanceof Error ? error.message : "Failed to delete component";
      throw error;
    } finally {
      this.$state.loading = false;
    }
  }
}

export const components = new ComponentStore();
