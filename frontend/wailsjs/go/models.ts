export namespace model {
	
	export class Component {
	    id: string;
	    name: string;
	    type: string;
	    description?: string;
	    parentId?: string;
	    properties?: Record<string, any>;
	    attachments?: string[];
	    metadata?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Component(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.parentId = source["parentId"];
	        this.properties = source["properties"];
	        this.attachments = source["attachments"];
	        this.metadata = source["metadata"];
	    }
	}
	export class ComponentTree {
	
	
	    static createFrom(source: any = {}) {
	        return new ComponentTree(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

export namespace snapshot {
	
	export class Snapshot {
	    id: string;
	    message: string;
	    // Go type: time
	    timestamp: any;
	    author: string;
	    tree: number[];
	
	    static createFrom(source: any = {}) {
	        return new Snapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.message = source["message"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.author = source["author"];
	        this.tree = source["tree"];
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

