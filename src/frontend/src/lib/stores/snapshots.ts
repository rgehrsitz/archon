export interface Snapshot {
  id: string;
  tag: string;
  message: string;
  timestamp: string;
  author: string;
  components: string[]; // Array of component IDs in this snapshot
}

export class SnapshotStore {
  $state = {
    snapshots: [] as Snapshot[],
    loading: false,
    error: null as string | null,
  };

  // Load snapshots from backend
  async loadSnapshots() {
    this.$state.loading = true;
    this.$state.error = null;
    try {
      const response = await fetch("/api/snapshots");
      if (!response.ok) throw new Error("Failed to load snapshots");
      const snapshots = await response.json();
      this.$state.snapshots = snapshots;
    } catch (error) {
      this.$state.error =
        error instanceof Error ? error.message : "Failed to load snapshots";
    } finally {
      this.$state.loading = false;
    }
  }

  // Create a new snapshot
  async createSnapshot(tag: string, message: string) {
    this.$state.loading = true;
    this.$state.error = null;
    try {
      const response = await fetch("/api/snapshots", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ tag, message }),
      });
      if (!response.ok) throw new Error("Failed to create snapshot");
      const newSnapshot = await response.json();
      this.$state.snapshots = [...this.$state.snapshots, newSnapshot];
      return newSnapshot;
    } catch (error) {
      this.$state.error =
        error instanceof Error ? error.message : "Failed to create snapshot";
      throw error;
    } finally {
      this.$state.loading = false;
    }
  }

  // Delete a snapshot
  async deleteSnapshot(id: string) {
    this.$state.loading = true;
    this.$state.error = null;
    try {
      const response = await fetch(`/api/snapshots/${id}`, {
        method: "DELETE",
      });
      if (!response.ok) throw new Error("Failed to delete snapshot");
      this.$state.snapshots = this.$state.snapshots.filter((s) => s.id !== id);
    } catch (error) {
      this.$state.error =
        error instanceof Error ? error.message : "Failed to delete snapshot";
      throw error;
    } finally {
      this.$state.loading = false;
    }
  }

  // Restore a snapshot
  async restoreSnapshot(id: string) {
    this.$state.loading = true;
    this.$state.error = null;
    try {
      const response = await fetch(`/api/snapshots/${id}/restore`, {
        method: "POST",
      });
      if (!response.ok) throw new Error("Failed to restore snapshot");
    } catch (error) {
      this.$state.error =
        error instanceof Error ? error.message : "Failed to restore snapshot";
      throw error;
    } finally {
      this.$state.loading = false;
    }
  }
}

export const snapshots = new SnapshotStore();
