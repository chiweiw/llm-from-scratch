import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { Environment, LocalConfig, ServerConfig, TargetFile, CheckResult } from '../types';

export const useEnvironmentStore = defineStore('environment', () => {
  const environments = ref<Environment[]>([]);
  const currentEnvironment = ref<Environment | null>(null);
  const checkResult = ref<CheckResult | null>(null);
  const loading = ref(false);
  const saving = ref(false);

  async function fetchEnvironments() {
    loading.value = true;
    try {
      const { GetEnvironments } = await import('../../wailsjs/go/main/App');
      environments.value = await GetEnvironments();
    } catch (error) {
      console.error('Failed to fetch environments:', error);
    } finally {
      loading.value = false;
    }
  }

  async function fetchEnvironment(id: string) {
    try {
      const { GetEnvironment } = await import('../../wailsjs/go/main/App');
      currentEnvironment.value = await GetEnvironment(id);
    } catch (error) {
      console.error('Failed to fetch environment:', error);
    }
  }

  async function saveEnvironment(env: Environment) {
    saving.value = true;
    try {
      const { SaveEnvironment } = await import('../../wailsjs/go/main/App');
      await SaveEnvironment(env);
      await fetchEnvironments();
    } catch (error) {
      console.error('Failed to save environment:', error);
      throw error;
    } finally {
      saving.value = false;
    }
  }

  async function deleteEnvironment(id: string) {
    try {
      const { DeleteEnvironment } = await import('../../wailsjs/go/main/App');
      await DeleteEnvironment(id);
      if (currentEnvironment.value?.id === id) {
        currentEnvironment.value = null;
      }
      await fetchEnvironments();
    } catch (error) {
      console.error('Failed to delete environment:', error);
      throw error;
    }
  }

  async function checkEnvironment(id: string) {
    try {
      const { CheckEnvironment } = await import('../../wailsjs/go/main/App');
      const result = await CheckEnvironment(id);
      checkResult.value = result;
      
      const envIndex = environments.value.findIndex(e => e.id === id);
      if (envIndex !== -1) {
        if (result.success) {
          const hasWarning = result.checks.some(c => c.status === 'warning');
          environments.value[envIndex].checkStatus = hasWarning ? 'warning' : 'pass';
        } else {
          environments.value[envIndex].checkStatus = 'error';
        }
      }
      
      return result;
    } catch (error) {
      console.error('Failed to check environment:', error);
      throw error;
    }
  }

  function createNewEnvironment(): Environment {
    const now = Date.now();
    return {
      id: `env_${now}`,
      name: '新环境',
      identifier: 'new',
      description: '',
      cloudDeploy: true,
      timeout: 600,
      dryRun: false,
      backupCleanup: true,
      local: {
        projectRoot: '',
        jdkPath: '',
        mavenPath: '',
        mavenSettingsPath: '',
        mavenRepoPath: '',
        mavenArgs: 'clean package -DskipTests',
        mavenQuiet: false,
        compactMvnLog: true,
        specifyPom: true,
        offlineBuild: false,
      },
      servers: [],
      targetFiles: [],
      checkStatus: 'unchecked',
      createdAt: now,
      updatedAt: now,
    };
  }

  function createNewServer(): ServerConfig {
    return {
      id: `server_${Date.now()}`,
      name: '新服务器',
      host: '',
      port: 22,
      username: '',
      password: '',
      deployDir: '',
      restartScript: '',
      enableRestart: false,
      useSudo: false,
    };
  }

  function createNewTargetFile(): TargetFile {
    return {
      id: `jar_${Date.now()}`,
      localPath: '',
      remoteName: '',
      defaultCheck: true,
    };
  }

  return {
    environments,
    currentEnvironment,
    checkResult,
    loading,
    saving,
    fetchEnvironments,
    fetchEnvironment,
    saveEnvironment,
    deleteEnvironment,
    checkEnvironment,
    createNewEnvironment,
    createNewServer,
    createNewTargetFile,
  };
});
