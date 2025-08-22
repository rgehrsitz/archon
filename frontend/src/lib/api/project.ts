export async function newProject(path: string, settings: Record<string, any>) {
  // Wails exposes bound services under window.go
  return (window as any).go?.api?.ProjectService?.New(path, settings);
}

export async function openProject(path: string) {
  return (window as any).go?.api?.ProjectService?.Open(path);
}
