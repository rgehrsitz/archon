// TypeScript definitions for project-related API calls
export interface Project {
  rootId: string;
  schemaVersion: number;
  settings?: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export interface ProjectInfo {
  rootId: string;
  schemaVersion: number;
  createdAt: string;
  updatedAt: string;
  settings?: Record<string, any>;
  currentPath?: string;
}

export interface ErrorEnvelope {
  code: string;
  message: string;
  details?: any;
}

// Project API functions
export async function createProject(path: string, settings: Record<string, any>): Promise<Project> {
  const result = await (window as any).go.api.ProjectService.CreateProject(path, settings);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
  return result;
}

export async function openProject(path: string): Promise<Project> {
  const result = await (window as any).go.api.ProjectService.OpenProject(path);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
  return result;
}

export async function closeProject(): Promise<void> {
  const result = await (window as any).go.api.ProjectService.CloseProject();
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
}

export async function getProjectInfo(): Promise<ProjectInfo> {
  const result = await (window as any).go.api.ProjectService.GetProjectInfo();
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
  return result;
}

export async function updateProjectSettings(settings: Record<string, any>): Promise<void> {
  const result = await (window as any).go.api.ProjectService.UpdateProjectSettings(settings);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
}

export async function projectExists(path: string): Promise<boolean> {
  const result = await (window as any).go.api.ProjectService.ProjectExists(path);
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
  return result;
}

export async function getCurrentProjectPath(): Promise<string> {
  const result = await (window as any).go.api.ProjectService.GetCurrentProjectPath();
  if (result.Code) {
    throw new Error(`${result.Code}: ${result.Message}`);
  }
  return result;
}

export async function isProjectOpen(): Promise<boolean> {
  return await (window as any).go.api.ProjectService.IsProjectOpen();
}
