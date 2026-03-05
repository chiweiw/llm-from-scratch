export interface Environment {
  id: string;
  name: string;
  identifier: string;
  description: string;
  projectRoot: string;
  cloudDeploy: boolean;
  timeout: number;
  servers: ServerConfig[];
  targetFiles: TargetFile[];
  checkStatus: string;
  createdAt: number;
  updatedAt: number;
}

export interface ServerConfig {
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
}

export interface TargetFile {
  id: string;
  localPath: string;
  remoteName: string;
  defaultCheck: boolean;
}

export interface GlobalSettings {
  defaultTimeout: number;
  logRetentionDays: number;
  backupEnabled: boolean;
  backupCleanup: boolean;
  notifyOnComplete: boolean;
  cloudDeploy: boolean;
  theme: string;
  language: string;
}

export interface SystemDefaultConfig {
  jdkPath: string;
  mavenPath: string;
  mavenSettingsPath: string;
  mavenRepoPath: string;
  mavenArgs: string[];
}

export interface DeployHistory {
  id: string;
  environmentId: string;
  environmentName: string;
  startTime: number;
  endTime: number;
  status: string;
  files: string[];
  duration: number;
  errorMessage: string;
}

export interface DeployLog {
  id: string;
  deployId: string;
  level: string;
  message: string;
  timestamp: number;
  createdAt: number;
}

export interface DeployProgress {
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
}

export interface StepProgress {
  name: string;
  status: string;
  progress: number;
  message: string;
}

export interface CheckResult {
  success: boolean;
  checks: CheckItem[];
  summary: string;
}

export interface CheckItem {
  name: string;
  status: string;
  message: string;
}
