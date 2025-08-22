export async function createSnapshot(tag: string, message = "", notes: Record<string, any> = {}) {
  return (window as any).go?.api?.SnapshotService?.Create(tag, message, notes);
}

export async function listSnapshots() {
  return (window as any).go?.api?.SnapshotService?.List();
}
