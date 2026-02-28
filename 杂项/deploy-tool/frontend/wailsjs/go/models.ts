export namespace models {
	
	export class CheckItem {
	    name: string;
	    status: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new CheckItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.message = source["message"];
	    }
	}
	export class CheckResult {
	    success: boolean;
	    checks: CheckItem[];
	    summary: string;
	
	    static createFrom(source: any = {}) {
	        return new CheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.checks = this.convertValues(source["checks"], CheckItem);
	        this.summary = source["summary"];
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
	export class DeployHistory {
	    id: string;
	    environmentId: string;
	    environmentName: string;
	    startTime: number;
	    endTime: number;
	    status: string;
	    files: string[];
	    duration: number;
	    errorMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new DeployHistory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.environmentId = source["environmentId"];
	        this.environmentName = source["environmentName"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	        this.status = source["status"];
	        this.files = source["files"];
	        this.duration = source["duration"];
	        this.errorMessage = source["errorMessage"];
	    }
	}
	export class StepProgress {
	    name: string;
	    status: string;
	    progress: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new StepProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.message = source["message"];
	    }
	}
	export class DeployProgress {
	    environmentId: string;
	    status: string;
	    currentStep: string;
	    totalProgress: number;
	    steps: StepProgress[];
	    currentFile: string;
	    fileProgress: number;
	    speed: string;
	    startTime: number;
	    endTime: number;
	    errorMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new DeployProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.environmentId = source["environmentId"];
	        this.status = source["status"];
	        this.currentStep = source["currentStep"];
	        this.totalProgress = source["totalProgress"];
	        this.steps = this.convertValues(source["steps"], StepProgress);
	        this.currentFile = source["currentFile"];
	        this.fileProgress = source["fileProgress"];
	        this.speed = source["speed"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	        this.errorMessage = source["errorMessage"];
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
	export class TargetFile {
	    id: string;
	    localPath: string;
	    remoteName: string;
	    defaultCheck: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TargetFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.localPath = source["localPath"];
	        this.remoteName = source["remoteName"];
	        this.defaultCheck = source["defaultCheck"];
	    }
	}
	export class ServerConfig {
	    id: string;
	    name: string;
	    host: string;
	    port: number;
	    username: string;
	    password: string;
	    deployDir: string;
	    restartScript: string;
	    enableRestart: boolean;
	    useSudo: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ServerConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.deployDir = source["deployDir"];
	        this.restartScript = source["restartScript"];
	        this.enableRestart = source["enableRestart"];
	        this.useSudo = source["useSudo"];
	    }
	}
	export class LocalConfig {
	    projectRoot: string;
	    jdkPath: string;
	    mavenPath: string;
	    mavenSettingsPath: string;
	    mavenRepoPath: string;
	    mavenArgs: string;
	    mavenQuiet: boolean;
	    compactMvnLog: boolean;
	    specifyPom: boolean;
	    offlineBuild: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LocalConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectRoot = source["projectRoot"];
	        this.jdkPath = source["jdkPath"];
	        this.mavenPath = source["mavenPath"];
	        this.mavenSettingsPath = source["mavenSettingsPath"];
	        this.mavenRepoPath = source["mavenRepoPath"];
	        this.mavenArgs = source["mavenArgs"];
	        this.mavenQuiet = source["mavenQuiet"];
	        this.compactMvnLog = source["compactMvnLog"];
	        this.specifyPom = source["specifyPom"];
	        this.offlineBuild = source["offlineBuild"];
	    }
	}
	export class Environment {
	    id: string;
	    name: string;
	    identifier: string;
	    description: string;
	    cloudDeploy: boolean;
	    timeout: number;
	    dryRun: boolean;
	    backupCleanup: boolean;
	    local: LocalConfig;
	    servers: ServerConfig[];
	    targetFiles: TargetFile[];
	    checkStatus: string;
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Environment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.identifier = source["identifier"];
	        this.description = source["description"];
	        this.cloudDeploy = source["cloudDeploy"];
	        this.timeout = source["timeout"];
	        this.dryRun = source["dryRun"];
	        this.backupCleanup = source["backupCleanup"];
	        this.local = this.convertValues(source["local"], LocalConfig);
	        this.servers = this.convertValues(source["servers"], ServerConfig);
	        this.targetFiles = this.convertValues(source["targetFiles"], TargetFile);
	        this.checkStatus = source["checkStatus"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
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
	export class GlobalSettings {
	    defaultTimeout: number;
	    logRetentionDays: number;
	    backupEnabled: boolean;
	    notifyOnComplete: boolean;
	    cloudDeploy: boolean;
	    theme: string;
	    language: string;
	
	    static createFrom(source: any = {}) {
	        return new GlobalSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defaultTimeout = source["defaultTimeout"];
	        this.logRetentionDays = source["logRetentionDays"];
	        this.backupEnabled = source["backupEnabled"];
	        this.notifyOnComplete = source["notifyOnComplete"];
	        this.cloudDeploy = source["cloudDeploy"];
	        this.theme = source["theme"];
	        this.language = source["language"];
	    }
	}
	
	
	
	export class SystemDefaultConfig {
	    jdkPath: string;
	    mavenPath: string;
	    mavenSettingsPath: string;
	    mavenRepoPath: string;
	    mavenArgs: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemDefaultConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.jdkPath = source["jdkPath"];
	        this.mavenPath = source["mavenPath"];
	        this.mavenSettingsPath = source["mavenSettingsPath"];
	        this.mavenRepoPath = source["mavenRepoPath"];
	        this.mavenArgs = source["mavenArgs"];
	    }
	}

}

