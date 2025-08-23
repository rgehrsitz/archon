// TypeScript definitions for node-related API calls
export interface Property {
  typeHint?: string;
  value: any;
}

export interface Node {
  id: string;
  name: string;
  description?: string;
  properties?: Record<string, Property>;
  children: string[];
  createdAt: string;
  updatedAt: string;
}

export interface CreateNodeRequest {
  parentId: string;
  name: string;
  description?: string;
  properties?: Record<string, Property>;
}

export interface UpdateNodeRequest {
  id: string;
  name?: string;
  description?: string;
  properties?: Record<string, Property>;
}

export interface MoveNodeRequest {
  nodeId: string;
  newParentId: string;
  position?: number;
}

export interface ReorderChildrenRequest {
  parentId: string;
  orderedChildIds: string[];
}

// Node API functions
export async function createNode(request: CreateNodeRequest): Promise<Node> {
  const result = await (window as any).go.api.NodeService.CreateNode(request);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
  return result;
}

export async function getNode(nodeId: string): Promise<Node> {
  const result = await (window as any).go.api.NodeService.GetNode(nodeId);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
  return result;
}

export async function updateNode(request: UpdateNodeRequest): Promise<Node> {
  const result = await (window as any).go.api.NodeService.UpdateNode(request);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
  return result;
}

export async function deleteNode(nodeId: string): Promise<void> {
  const result = await (window as any).go.api.NodeService.DeleteNode(nodeId);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
}

export async function moveNode(request: MoveNodeRequest): Promise<void> {
  const result = await (window as any).go.api.NodeService.MoveNode(request);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
}

export async function reorderChildren(request: ReorderChildrenRequest): Promise<void> {
  const result = await (window as any).go.api.NodeService.ReorderChildren(request);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
}

export async function listChildren(nodeId: string): Promise<Node[]> {
  const result = await (window as any).go.api.NodeService.ListChildren(nodeId);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
  return result;
}

export async function getNodePath(nodeId: string): Promise<Node[]> {
  const result = await (window as any).go.api.NodeService.GetNodePath(nodeId);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
  return result;
}

export async function getRootNode(): Promise<Node> {
  const result = await (window as any).go.api.NodeService.GetRootNode();
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
  return result;
}

export async function setProperty(nodeId: string, key: string, value: any, typeHint?: string): Promise<void> {
  const result = await (window as any).go.api.NodeService.SetProperty(nodeId, key, value, typeHint || '');
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
}

export async function deleteProperty(nodeId: string, key: string): Promise<void> {
  const result = await (window as any).go.api.NodeService.DeleteProperty(nodeId, key);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
}

// Convenience functions
export async function createSimpleNode(parentId: string, name: string, description?: string): Promise<Node> {
  return createNode({ parentId, name, description });
}

export async function updateNodeName(nodeId: string, name: string): Promise<Node> {
  return updateNode({ id: nodeId, name });
}

export async function updateNodeDescription(nodeId: string, description: string): Promise<Node> {
  return updateNode({ id: nodeId, description });
}
