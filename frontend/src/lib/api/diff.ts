export async function diff(refA: string, refB: string) {
  return (window as any).go?.api?.DiffService?.Diff(refA, refB);
}
