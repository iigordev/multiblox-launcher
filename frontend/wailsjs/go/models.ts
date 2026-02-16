export namespace main {
	
	export class Instance {
	    name: string;
	    path: string;
	    status: string;
	    iconFolder: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new Instance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.status = source["status"];
	        this.iconFolder = source["iconFolder"];
	        this.version = source["version"];
	    }
	}

}

