export namespace entity {
	
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
	export class Environment {
	    id: string;
	    name: string;
	    identifier: string;
	    description: string;
	    projectRoot: string;
	    cloudDeploy: boolean;
	    timeout: number;
	    backupCleanup: boolean;
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
	        this.projectRoot = source["projectRoot"];
	        this.cloudDeploy = source["cloudDeploy"];
	        this.timeout = source["timeout"];
	        this.backupCleanup = source["backupCleanup"];
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
	    mavenArgs: string[];
	
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

export namespace request {
	
	export class CheckEnvironment {
	    id: string;
	
	    static createFrom(source: any = {}) {
	        return new CheckEnvironment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	    }
	}
	export class DeleteEnvironment {
	    id: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteEnvironment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	    }
	}
	export class ExportEnvironment {
	    id: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportEnvironment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	    }
	}
	export class ImportEnvironment {
	    json: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportEnvironment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.json = source["json"];
	    }
	}
	export class ParseMavenCommand {
	    command: string;
	
	    static createFrom(source: any = {}) {
	        return new ParseMavenCommand(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.command = source["command"];
	    }
	}
	export class SaveEnvironment {
	    environment: entity.Environment;
	
	    static createFrom(source: any = {}) {
	        return new SaveEnvironment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.environment = this.convertValues(source["environment"], entity.Environment);
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
	export class SaveGlobalSettings {
	    settings: entity.GlobalSettings;
	
	    static createFrom(source: any = {}) {
	        return new SaveGlobalSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.settings = this.convertValues(source["settings"], entity.GlobalSettings);
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
	export class SaveSystemDefaults {
	    defaults: entity.SystemDefaultConfig;
	
	    static createFrom(source: any = {}) {
	        return new SaveSystemDefaults(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defaults = this.convertValues(source["defaults"], entity.SystemDefaultConfig);
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
	export class StartDeploy {
	    environmentId: string;
	    jarIds: string[];
	
	    static createFrom(source: any = {}) {
	        return new StartDeploy(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.environmentId = source["environmentId"];
	        this.jarIds = source["jarIds"];
	    }
	}

}

export namespace response {
	
	export class Base {
	    code: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Base(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	    }
	}
	export class Data__deploy_tool_internal_model_entity_CheckResult_ {
	    code: number;
	    message: string;
	    data?: entity.CheckResult;
	
	    static createFrom(source: any = {}) {
	        return new Data__deploy_tool_internal_model_entity_CheckResult_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], entity.CheckResult);
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
	export class Data__deploy_tool_internal_model_entity_DeployProgress_ {
	    code: number;
	    message: string;
	    data?: entity.DeployProgress;
	
	    static createFrom(source: any = {}) {
	        return new Data__deploy_tool_internal_model_entity_DeployProgress_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], entity.DeployProgress);
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
	export class Data__deploy_tool_internal_model_entity_Environment_ {
	    code: number;
	    message: string;
	    data?: entity.Environment;
	
	    static createFrom(source: any = {}) {
	        return new Data__deploy_tool_internal_model_entity_Environment_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], entity.Environment);
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
	export class Data__deploy_tool_internal_service_MavenParseResult_ {
	    code: number;
	    message: string;
	    data?: service.MavenParseResult;
	
	    static createFrom(source: any = {}) {
	        return new Data__deploy_tool_internal_service_MavenParseResult_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], service.MavenParseResult);
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
	export class Data___deploy_tool_internal_model_entity_DeployHistory_ {
	    code: number;
	    message: string;
	    data: entity.DeployHistory[];
	
	    static createFrom(source: any = {}) {
	        return new Data___deploy_tool_internal_model_entity_DeployHistory_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], entity.DeployHistory);
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
	export class Data___deploy_tool_internal_model_entity_Environment_ {
	    code: number;
	    message: string;
	    data: entity.Environment[];
	
	    static createFrom(source: any = {}) {
	        return new Data___deploy_tool_internal_model_entity_Environment_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], entity.Environment);
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
	export class Data___map_string_string_ {
	    code: number;
	    message: string;
	    data: any[];
	
	    static createFrom(source: any = {}) {
	        return new Data___map_string_string_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = source["data"];
	    }
	}
	export class Data_deploy_tool_internal_model_entity_GlobalSettings_ {
	    code: number;
	    message: string;
	    data: entity.GlobalSettings;
	
	    static createFrom(source: any = {}) {
	        return new Data_deploy_tool_internal_model_entity_GlobalSettings_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], entity.GlobalSettings);
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
	export class Data_deploy_tool_internal_model_entity_SystemDefaultConfig_ {
	    code: number;
	    message: string;
	    data: entity.SystemDefaultConfig;
	
	    static createFrom(source: any = {}) {
	        return new Data_deploy_tool_internal_model_entity_SystemDefaultConfig_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = this.convertValues(source["data"], entity.SystemDefaultConfig);
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
	export class Data_string_ {
	    code: number;
	    message: string;
	    data: string;
	
	    static createFrom(source: any = {}) {
	        return new Data_string_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.data = source["data"];
	    }
	}

}

export namespace service {
	
	export class DeployService {
	
	
	    static createFrom(source: any = {}) {
	        return new DeployService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class MavenParseResult {
	    mavenPath: string;
	    settingsPath: string;
	    repoLocal: string;
	    pomFile: string;
	    goals: string[];
	    properties: Record<string, string>;
	    argsArray: string[];
	
	    static createFrom(source: any = {}) {
	        return new MavenParseResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mavenPath = source["mavenPath"];
	        this.settingsPath = source["settingsPath"];
	        this.repoLocal = source["repoLocal"];
	        this.pomFile = source["pomFile"];
	        this.goals = source["goals"];
	        this.properties = source["properties"];
	        this.argsArray = source["argsArray"];
	    }
	}

}

