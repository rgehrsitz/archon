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

// Access Wails API through global object
const wailsApi = () => (window as any).go?.api?.ProjectService;

// Note: Based on Wails v2.10.x docs, context should be handled automatically by Wails
// However, the backend still expects context.Context, so we'll provide a minimal context
// that Go can unmarshal until this is fixed in the backend
function createMinimalContext() {
  // Return null which some Wails setups accept as a background context
  return null;
}

// Helper function to handle API responses that might be errors
function handleApiResponse<T>(result: T | ErrorEnvelope): T {
  if (result && typeof result === 'object' && 'Code' in result) {
    const error = result as ErrorEnvelope;
    throw new Error(`${error.code}: ${error.message}`);
  }
  return result as T;
}

// Project API functions - following proper Wails v2 pattern (no context from frontend)
export async function createProject(path: string, settings: Record<string, any> = {}): Promise<Project> {
  const api = wailsApi();
  if (!api) throw new Error('Wails API not available');
  
  try {
    // Call without context - backend should handle context internally
    const result = await api.CreateProject(path, settings);
    return handleApiResponse<Project>(result);
  } catch (error) {
    console.error('CreateProject failed:', error);
    console.error('This indicates backend still has context.Context parameters (see BACKEND_CONTEXT_FIX.md)');
    throw new Error(`Failed to create project: ${error}`);
  }
}

export async function openProject(path: string): Promise<Project> {
  const api = wailsApi();
  if (!api) throw new Error('Wails API not available');
  
  try {
    // Call with only the path parameter - Wails handles context automatically
    const result = await api.OpenProject(path);
    return handleApiResponse<Project>(result);
  } catch (error) {
    console.error('OpenProject failed:', error);
    throw new Error(`Failed to open project: ${error}`);
  }
}

export async function closeProject(): Promise<void> {
  const api = wailsApi();
  if (!api) throw new Error('Wails API not available');
  
  try {
    // Call with no parameters - Wails handles context automatically
    const result = await api.CloseProject();
    handleApiResponse<void>(result);
  } catch (error) {
    console.error('CloseProject failed:', error);
    throw new Error(`Failed to close project: ${error}`);
  }
}

export async function getProjectInfo(): Promise<ProjectInfo> {
  const api = wailsApi();
  if (!api) throw new Error('Wails API not available');
  
  try {
    // Call with no parameters - Wails handles context automatically
    const result = await api.GetProjectInfo();
    return handleApiResponse<ProjectInfo>(result);
  } catch (error) {
    console.error('GetProjectInfo failed:', error);
    throw new Error(`Failed to get project info: ${error}`);
  }
}

export async function updateProjectSettings(settings: Record<string, any>): Promise<void> {
  const api = wailsApi();
  if (!api) throw new Error('Wails API not available');
  
  try {
    // Call with only settings parameter - Wails handles context automatically
    const result = await api.UpdateProjectSettings(settings);
    handleApiResponse<void>(result);
  } catch (error) {
    console.error('UpdateProjectSettings failed:', error);
    throw new Error(`Failed to update project settings: ${error}`);
  }
}

export async function projectExists(path: string): Promise<boolean> {
  const api = wailsApi();
  if (!api) throw new Error('Wails API not available');
  
  try {
    // Call with only path parameter - Wails handles context automatically
    const result = await api.ProjectExists(path);
    return handleApiResponse<boolean>(result);
  } catch (error) {
    console.error('ProjectExists failed:', error);
    // Return false instead of throwing to avoid breaking the UI
    return false;
  }
}

export async function getCurrentProjectPath(): Promise<string> {
  const api = wailsApi();
  if (!api) throw new Error('Wails API not available');
  
  try {
    // Call with no parameters - Wails handles context automatically
    const result = await api.GetCurrentProjectPath();
    return handleApiResponse<string>(result);
  } catch (error) {
    console.error('GetCurrentProjectPath failed:', error);
    throw new Error(`Failed to get current project path: ${error}`);
  }
}

export async function isProjectOpen(): Promise<boolean> {
  const api = wailsApi();
  if (!api) {
    console.warn('Wails API not available');
    return false;
  }
  
  try {
    // Call the backend method without context - backend should handle context internally
    const result = await api.IsProjectOpen();
    return result;
  } catch (error) {
    console.warn('IsProjectOpen failed:', error);
    console.warn('This indicates the backend methods still include context.Context parameters');
    console.warn('Backend needs to be updated to follow Wails v2 patterns');
    return false;
  }
}

export async function getCurrentProject() {
  const api = wailsApi();
  if (!api) {
    console.warn('Wails API not available');
    return null;
  }
  
  try {
    // This method should not require any parameters
    const result = await api.GetCurrentProject();
    return result;
  } catch (error) {
    console.warn('GetCurrentProject failed:', error);
    return null;
  }
}

// File dialog functions - we'll need to add these for proper project opening
export async function selectDirectory(): Promise<string | null> {
  // For now, we'll return null and implement this with a proper file dialog later
  // In a real implementation, we'd use Wails' dialog API
  return new Promise((resolve) => {
    const path = prompt('Enter project directory path:');
    resolve(path);
  });
}

export async function selectFile(): Promise<string | null> {
  // For now, we'll return null and implement this with a proper file dialog later
  // In a real implementation, we'd use Wails' dialog API
  return new Promise((resolve) => {
    const path = prompt('Enter project file path:');
    resolve(path);
  });
}
