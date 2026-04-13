export namespace hosts {
	
	export class HostEntry {
	    id: string;
	    ip: string;
	    domain: string;
	    comment: string;
	    enabled: boolean;
	    groupId: string;
	    source: string;
	    raw: string;
	
	    static createFrom(source: any = {}) {
	        return new HostEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.ip = source["ip"];
	        this.domain = source["domain"];
	        this.comment = source["comment"];
	        this.enabled = source["enabled"];
	        this.groupId = source["groupId"];
	        this.source = source["source"];
	        this.raw = source["raw"];
	    }
	}
	export class HostGroup {
	    id: string;
	    name: string;
	    enabled: boolean;
	    order: number;
	
	    static createFrom(source: any = {}) {
	        return new HostGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.enabled = source["enabled"];
	        this.order = source["order"];
	    }
	}
	export class HostWorkspace {
	    composeEnabled: boolean;
	    theme: string;
	    groups: HostGroup[];
	    entries: HostEntry[];
	    readonlySystemLines: string[];
	    systemHostsLines: string[];
	
	    static createFrom(source: any = {}) {
	        return new HostWorkspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.composeEnabled = source["composeEnabled"];
	        this.theme = source["theme"];
	        this.groups = this.convertValues(source["groups"], HostGroup);
	        this.entries = this.convertValues(source["entries"], HostEntry);
	        this.readonlySystemLines = source["readonlySystemLines"];
	        this.systemHostsLines = source["systemHostsLines"];
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

export namespace update {
	
	export class Info {
	    currentVersion: string;
	    latestVersion: string;
	    hasUpdate: boolean;
	    releaseNotes: string;
	    publishedAt: string;
	    assetName: string;
	    assetURL: string;
	    assetSize: number;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.hasUpdate = source["hasUpdate"];
	        this.releaseNotes = source["releaseNotes"];
	        this.publishedAt = source["publishedAt"];
	        this.assetName = source["assetName"];
	        this.assetURL = source["assetURL"];
	        this.assetSize = source["assetSize"];
	    }
	}
	export class Settings {
	    autoCheckOnStartup: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.autoCheckOnStartup = source["autoCheckOnStartup"];
	    }
	}

}

