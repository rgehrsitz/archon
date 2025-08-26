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

