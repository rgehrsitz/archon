export async function threeWay(base: string, ours: string, theirs: string) {
  return (window as any).go?.api?.MergeService?.ThreeWay(base, ours, theirs);
}
