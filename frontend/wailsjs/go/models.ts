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

