// Simple placeholder for a plugin worker loader. Real impl will spawn a Worker per run.
export async function runPlugin(plugin: { run: (input: Uint8Array, opts?: Record<string, any>) => Promise<any> }, input: Uint8Array, opts?: Record<string, any>) {
  return plugin.run(input, opts);
}
