export namespace main {

	export class LocaleOption {
	    code: string;
	    label: string;

	    static createFrom(source: any = {}) {
	        return new LocaleOption(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.label = source["label"];
	    }
	}
	export class ProductInfo {
	    appId: string;
	    appSlug: string;
	    displayName: string;
	    version: string;
	    repository: string;
	    brandParts: Record<string, string>;
	    languages: LocaleOption[];

	    static createFrom(source: any = {}) {
	        return new ProductInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appId = source["appId"];
	        this.appSlug = source["appSlug"];
	        this.displayName = source["displayName"];
	        this.version = source["version"];
	        this.repository = source["repository"];
	        this.brandParts = source["brandParts"];
	        this.languages = this.convertValues(source["languages"], LocaleOption);
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
	export class Bootstrap {
	    product: ProductInfo;
	    currentLocale: string;
	    currentTheme: string;
	    shellLocale: Record<string, string>;
	    businessLocale: Record<string, string>;

	    static createFrom(source: any = {}) {
	        return new Bootstrap(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.product = this.convertValues(source["product"], ProductInfo);
	        this.currentLocale = source["currentLocale"];
	        this.currentTheme = source["currentTheme"];
	        this.shellLocale = source["shellLocale"];
	        this.businessLocale = source["businessLocale"];
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
	export class Heading {
	    id: string;
	    level: number;
	    text: string;

	    static createFrom(source: any = {}) {
	        return new Heading(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.level = source["level"];
	        this.text = source["text"];
	    }
	}
	export class Document {
	    path: string;
	    name: string;
	    title: string;
	    html: string;
	    outline: Heading[];
	    modifiedAt: number;
	    size: number;

	    static createFrom(source: any = {}) {
	        return new Document(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.title = source["title"];
	        this.html = source["html"];
	        this.outline = this.convertValues(source["outline"], Heading);
	        this.modifiedAt = source["modifiedAt"];
	        this.size = source["size"];
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
	export class DocumentStatus {
	    path: string;
	    exists: boolean;
	    isDocument: boolean;
	    changed: boolean;
	    modifiedAt: number;
	    size: number;

	    static createFrom(source: any = {}) {
	        return new DocumentStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.exists = source["exists"];
	        this.isDocument = source["isDocument"];
	        this.changed = source["changed"];
	        this.modifiedAt = source["modifiedAt"];
	        this.size = source["size"];
	    }
	}



	export class ReadingPosition {
	    path: string;
	    relativePath: string;
	    scrollTop: number;
	    scrollRatio: number;
	    headingId: string;
	    modifiedAt: number;
	    size: number;
	    updatedAt: number;

	    static createFrom(source: any = {}) {
	        return new ReadingPosition(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.relativePath = source["relativePath"];
	        this.scrollTop = source["scrollTop"];
	        this.scrollRatio = source["scrollRatio"];
	        this.headingId = source["headingId"];
	        this.modifiedAt = source["modifiedAt"];
	        this.size = source["size"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ReadingMemorySnapshot {
	    lastDocument: string;
	    lastPosition?: ReadingPosition;

	    static createFrom(source: any = {}) {
	        return new ReadingMemorySnapshot(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lastDocument = source["lastDocument"];
	        this.lastPosition = this.convertValues(source["lastPosition"], ReadingPosition);
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

	export class ReadingTab {
	    path: string;
	    relativePath: string;
	    name: string;

	    static createFrom(source: any = {}) {
	        return new ReadingTab(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.relativePath = source["relativePath"];
	        this.name = source["name"];
	    }
	}
	export class TreeNode {
	    name: string;
	    path: string;
	    type: string;
	    children?: TreeNode[];

	    static createFrom(source: any = {}) {
	        return new TreeNode(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.type = source["type"];
	        this.children = this.convertValues(source["children"], TreeNode);
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
	export class WorkspaceTree {
	    root: TreeNode;

	    static createFrom(source: any = {}) {
	        return new WorkspaceTree(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = this.convertValues(source["root"], TreeNode);
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
	export class RenameResult {
	    oldPath: string;
	    newPath: string;
	    nodeType: string;
	    tree?: WorkspaceTree;

	    static createFrom(source: any = {}) {
	        return new RenameResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oldPath = source["oldPath"];
	        this.newPath = source["newPath"];
	        this.nodeType = source["nodeType"];
	        this.tree = this.convertValues(source["tree"], WorkspaceTree);
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
	export class SearchResult {
	    path: string;
	    name: string;
	    relativePath: string;
	    kind: string;
	    snippet: string;

	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.relativePath = source["relativePath"];
	        this.kind = source["kind"];
	        this.snippet = source["snippet"];
	    }
	}

	export class UpdateInfo {
	    currentVersion: string;
	    latestVersion: string;
	    updateAvailable: boolean;
	    ignored: boolean;
	    releaseUrl: string;
	    downloadUrl: string;
	    sha256: string;
	    releaseNotes: string;
	    downloadedPath: string;

	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.updateAvailable = source["updateAvailable"];
	        this.ignored = source["ignored"];
	        this.releaseUrl = source["releaseUrl"];
	        this.downloadUrl = source["downloadUrl"];
	        this.sha256 = source["sha256"];
	        this.releaseNotes = source["releaseNotes"];
	        this.downloadedPath = source["downloadedPath"];
	    }
	}
	export class WorkspaceReadingSession {
	    openTabs: ReadingTab[];
	    activeDocument: string;
	    activePosition?: ReadingPosition;

	    static createFrom(source: any = {}) {
	        return new WorkspaceReadingSession(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.openTabs = this.convertValues(source["openTabs"], ReadingTab);
	        this.activeDocument = source["activeDocument"];
	        this.activePosition = this.convertValues(source["activePosition"], ReadingPosition);
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
	export class WorkspaceRefresh {
	    changed: boolean;
	    tree?: WorkspaceTree;

	    static createFrom(source: any = {}) {
	        return new WorkspaceRefresh(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.changed = source["changed"];
	        this.tree = this.convertValues(source["tree"], WorkspaceTree);
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
