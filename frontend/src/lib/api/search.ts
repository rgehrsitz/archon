export async function search(q: string) {
  return (window as any).go?.api?.IndexService?.Search(q);
}

export async function rebuildIndex() {
  return (window as any).go?.api?.IndexService?.Rebuild();
}
