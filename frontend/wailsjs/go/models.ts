export namespace api {
	
	export class LogEntry {
	    level: string;
	    message: string;
	    context?: Record<string, any>;
	    timestamp?: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.message = source["message"];
	        this.context = source["context"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class LoggingConfig {
	    level: string;
	    outputConsole: boolean;
	    outputFile: boolean;
	    logDirectory: string;
	    maxFileSize: number;
	    maxBackups: number;
	    maxAge: number;
	    compressBackups: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LoggingConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.outputConsole = source["outputConsole"];
	        this.outputFile = source["outputFile"];
	        this.logDirectory = source["logDirectory"];
	        this.maxFileSize = source["maxFileSize"];
	        this.maxBackups = source["maxBackups"];
	        this.maxAge = source["maxAge"];
	        this.compressBackups = source["compressBackups"];
	    }
	}

}

export namespace errors {
	
	export class Envelope {
	    code: string;
	    message: string;
	    details?: any;
	
	    static createFrom(source: any = {}) {
	        return new Envelope(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.details = source["details"];
	    }
	}

}

export namespace store {
	
	export class ProjectStore {
	    // Go type: index
	    IndexManager?: any;
	
	    static createFrom(source: any = {}) {
	        return new ProjectStore(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.IndexManager = this.convertValues(source["IndexManager"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace plugins {
	
	export class SecretValue {
	    name: string;
	    value: string;
	    redacted: boolean;
	    metadata?: Record<string, any>;

	    static createFrom(source: any = {}) {
	        return new SecretValue(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	        this.redacted = source["redacted"];
	        this.metadata = source["metadata"];
	    }
	}

	export class ProxyRequest {
	    method: string;
	    url: string;
	    headers?: Record<string, string>;
	    body?: number[];
	    timeoutMs?: number;

	    static createFrom(source: any = {}) {
	        return new ProxyRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.url = source["url"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.timeoutMs = source["timeoutMs"];
	    }
	}

	export class ProxyResponse {
	    status: number;
	    headers?: Record<string, string>;
	    body?: number[];

	    static createFrom(source: any = {}) {
	        return new ProxyResponse(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	    }
	}

	export class NodeData {
	    id?: string;
	    name?: string;
	    description?: string;
	    properties?: Record<string, any>;
	    children?: string[];

	    static createFrom(source: any = {}) {
	        return new NodeData(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.properties = source["properties"];
	        this.children = source["children"];
	    }
	}

	export class Mutation {
	    type: string;
	    nodeId?: string;
	    parentId?: string;
	    data?: NodeData;
	    position?: number;

	    static createFrom(source: any = {}) {
	        return new Mutation(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.nodeId = source["nodeId"];
	        this.parentId = source["parentId"];
	        this.data = this.convertValues(source["data"], NodeData);
	        this.position = source["position"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

	export class PluginMetadata {
	    category?: string;
	    tags?: string[];
	    website?: string;
	    repository?: string;

	    static createFrom(source: any = {}) {
	        return new PluginMetadata(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.tags = source["tags"];
	        this.website = source["website"];
	        this.repository = source["repository"];
	    }
	}

	export class PluginManifest {
	    id: string;
	    name: string;
	    version: string;
	    type: string;
	    description?: string;
	    author?: string;
	    license?: string;
	    permissions: string[];
	    entryPoint: string;
	    archonVersion?: string;
	    integrity?: string;
	    metadata?: PluginMetadata;

	    static createFrom(source: any = {}) {
	        return new PluginManifest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.author = source["author"];
	        this.license = source["license"];
	        this.permissions = source["permissions"];
	        this.entryPoint = source["entryPoint"];
	        this.archonVersion = source["archonVersion"];
	        this.integrity = source["integrity"];
	        this.metadata = this.convertValues(source["metadata"], PluginMetadata);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

	export class PluginInstallation {
	    manifest: PluginManifest;
	    path: string;
	    // Go type: time
	    installedAt: any;
	    enabled: boolean;
	    source: string;

	    static createFrom(source: any = {}) {
	        return new PluginInstallation(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.manifest = this.convertValues(source["manifest"], PluginManifest);
	        this.path = source["path"];
	        this.installedAt = this.convertValues(source["installedAt"], null);
	        this.enabled = source["enabled"];
	        this.source = source["source"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

	export class PluginPermissionGrant {
	    pluginId: string;
	    permission: string;
	    granted: boolean;
	    temporary: boolean;
	    // Go type: time
	    expiresAt?: any;
	    // Go type: time
	    grantedAt: any;

	    static createFrom(source: any = {}) {
	        return new PluginPermissionGrant(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pluginId = source["pluginId"];
	        this.permission = source["permission"];
	        this.granted = source["granted"];
	        this.temporary = source["temporary"];
	        this.expiresAt = this.convertValues(source["expiresAt"], null);
	        this.grantedAt = this.convertValues(source["grantedAt"], null);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace types {
	
	export class Property {
	    typeHint?: string;
	    value: any;
	
	    static createFrom(source: any = {}) {
	        return new Property(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.typeHint = source["typeHint"];
	        this.value = source["value"];
	    }
	}
	export class CreateNodeRequest {
	    parentId: string;
	    name: string;
	    description?: string;
	    properties?: Record<string, Property>;
	
	    static createFrom(source: any = {}) {
	        return new CreateNodeRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parentId = source["parentId"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.properties = this.convertValues(source["properties"], Property, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MoveNodeRequest {
	    nodeId: string;
	    newParentId: string;
	    position?: number;
	
	    static createFrom(source: any = {}) {
	        return new MoveNodeRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeId = source["nodeId"];
	        this.newParentId = source["newParentId"];
	        this.position = source["position"];
	    }
	}
	export class Node {
	    id: string;
	    name: string;
	    description?: string;
	    properties?: Record<string, Property>;
	    children: string[];
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Node(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.properties = this.convertValues(source["properties"], Property, true);
	        this.children = source["children"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Project {
	    rootId: string;
	    schemaVersion: number;
	    settings?: Record<string, any>;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rootId = source["rootId"];
	        this.schemaVersion = source["schemaVersion"];
	        this.settings = source["settings"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ReorderChildrenRequest {
	    parentId: string;
	    orderedChildIds: string[];
	
	    static createFrom(source: any = {}) {
	        return new ReorderChildrenRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parentId = source["parentId"];
	        this.orderedChildIds = source["orderedChildIds"];
	    }
	}
	export class UpdateNodeRequest {
	    id: string;
	    name?: string;
	    description?: string;
	    properties?: Record<string, Property>;
	
	    static createFrom(source: any = {}) {
	        return new UpdateNodeRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.properties = this.convertValues(source["properties"], Property, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

