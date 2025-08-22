export type ArchonNode = {
  id: string;
  name: string;
  description?: string;
  properties?: Record<string, { key: string; typeHint?: string; value: any }>;
  children: string[];
};

export type ImportResult = {
  nodes: ArchonNode[];
  warnings?: string[];
};

export type ImportPlugin = {
  id: string;
  name: string;
  run(input: Uint8Array, opts?: Record<string, any>): Promise<ImportResult>;
};
