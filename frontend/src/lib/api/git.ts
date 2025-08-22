export async function status() {
  return (window as any).go?.api?.GitService?.Status();
}
