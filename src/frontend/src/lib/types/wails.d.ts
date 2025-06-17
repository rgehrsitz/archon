// Wails bindings type definitions
export interface Component {
    id: string;
    name: string;
    type: string;
    description?: string;
    parentId?: string;
    properties?: Record<string, any>;
    attachments?: string[];
    metadata?: Record<string, string>;
}

export interface ComponentTree {
    components: Component[];
    rootIds: string[];
}

export interface Snapshot {
    id: string;
    name: string;
    description?: string;
    timestamp: string;
    author: string;
}

export interface Plugin {
    id: string;
    name: string;
    version: string;
    category: string;
    description?: string;
    enabled: boolean;
}

declare global {
    interface Window {
        go?: {
            main?: {
                App?: {
                    LoadProject: (path: string) => Promise<void>;
                    GetComponentTree: () => Promise<ComponentTree>;
                    CreateComponent: (component: Partial<Component>) => Promise<Component>;
                    UpdateComponent: (id: string, component: Partial<Component>) => Promise<Component>;
                    DeleteComponent: (id: string) => Promise<void>;
                    CreateSnapshot: (message: string) => Promise<Snapshot>;
                    GetSnapshots: () => Promise<Snapshot[]>;
                    LoadPlugin: (path: string) => Promise<void>;
                    ExecutePlugin: (pluginId: string, params: Record<string, any>) => Promise<any>;
                };
            };
        };
    }
}

export { };
