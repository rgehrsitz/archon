export async function runImport(pluginId: string, bytes: Uint8Array, opts: Record<string, any> = {}) {
  return (window as any).go?.api?.ImportService?.Run(pluginId, Array.from(bytes), opts);
}
