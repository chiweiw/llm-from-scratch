export interface Environment {
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
  checkStatus?: 'pass' | 'warning' | 'error' | 'unchecked';
  createdAt: number;
  updatedAt: number;
}

export interface LocalConfig {
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
  mavenArgs: string;
}

export interface DeployHistory {
  id: string;
  environmentId: string;
  environmentName: string;
  status: 'success' | 'failed' | 'cancelled';
  startTime: number;
  endTime: number;
  logPath: string;
  errorMsg?: string;
}

export interface DeployProgress {
  phase: 'idle' | 'checking' | 'building' | 'uploading' | 'restarting' | 'completed' | 'failed';
  progress: number;
  message: string;
  currentServer?: string;
  currentFile?: string;
}

export interface CheckResult {
  success: boolean;
  checks: CheckItem[];
  summary: string;
}

export interface CheckItem {
  name: string;
  status: 'pass' | 'warning' | 'error';
  message: string;
}
