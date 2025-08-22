export async function getNode(id: string) {
  return (window as any).go?.api?.NodeService?.Get(id);
}

export async function createNode(parentId: string, name: string) {
  return (window as any).go?.api?.NodeService?.Create(parentId, name);
}
