import { defineStore } from "pinia";
import { ref } from "vue";
import type {
  Environment,
  ServerConfig,
  TargetFile,
  CheckResult,
  CheckItem,
} from "../types";

export const useEnvironmentStore = defineStore("environment", () => {
  const environments = ref<Environment[]>([]);
  const currentEnvironment = ref<Environment | null>(null);
  const checkResult = ref<CheckResult | null>(null);
  const loading = ref(false);
  const saving = ref(false);

  async function fetchEnvironments() {
    loading.value = true;
    try {
      const { GetEnvironments } = await import("../../wailsjs/go/app/App");
      const resp = await GetEnvironments();
      environments.value = resp.code === 0 ? resp.data : [];
    } catch (error) {
      console.error("Failed to fetch environments:", error);
    } finally {
      loading.value = false;
    }
  }

  async function fetchEnvironment(id: string) {
    try {
      const { GetEnvironment } = await import("../../wailsjs/go/app/App");
      const resp = await GetEnvironment(id);
      currentEnvironment.value = resp.code === 0 ? resp.data : null;
    } catch (error) {
      console.error("Failed to fetch environment:", error);
    }
  }

  function applyStableOrder(list: Environment[], order: string[]): Environment[] {
    const indexMap = new Map(order.map((id, idx) => [id, idx]));
    return [...list].sort((a, b) => {
      const ai = indexMap.get(a.id);
      const bi = indexMap.get(b.id);
      const aIndex = ai === undefined ? Number.MAX_SAFE_INTEGER : ai;
      const bIndex = bi === undefined ? Number.MAX_SAFE_INTEGER : bi;
      return aIndex - bIndex;
    });
  }

  async function saveEnvironment(env: Environment) {
    saving.value = true;
    try {
      const { SaveEnvironment } = await import("../../wailsjs/go/app/App");
      const resp = await SaveEnvironment({ environment: env } as any);
      if (resp.code !== 0) {
        throw new Error(resp.message || "保存失败");
      }
      await fetchEnvironments();
    } catch (error) {
      console.error("Failed to save environment:", error);
      throw error;
    } finally {
      saving.value = false;
    }
  }

  async function deleteEnvironment(id: string) {
    try {
      const { DeleteEnvironment } = await import("../../wailsjs/go/app/App");
      const resp = await DeleteEnvironment({ id });
      if (resp.code !== 0) {
        throw new Error(resp.message || "删除失败");
      }
      if (currentEnvironment.value?.id === id) {
        currentEnvironment.value = null;
      }
      await fetchEnvironments();
    } catch (error) {
      console.error("Failed to delete environment:", error);
      throw error;
    }
  }

  async function checkEnvironment(id: string) {
    try {
      const { CheckEnvironment } = await import("../../wailsjs/go/app/App");
      const resp = await CheckEnvironment({ id });
      if (resp.code !== 0 || !resp.data) {
        throw new Error(resp.message || "自检失败");
      }
      const result = resp.data;
      checkResult.value = result;

      const envIndex = environments.value.findIndex((e) => e.id === id);
      if (envIndex !== -1) {
        if (result.success) {
          const hasWarning = result.checks.some(
            (c: CheckItem) => c.status === "warning"
          );
          environments.value[envIndex].checkStatus = hasWarning
            ? "warning"
            : "pass";
        } else {
          environments.value[envIndex].checkStatus = "error";
        }
      }

      return result;
    } catch (error) {
      console.error("Failed to check environment:", error);
      throw error;
    }
  }

  async function createNewEnvironment(): Promise<Environment> {
    const now = Date.now();
    return {
      id: `env_${now}`,
      name: "新环境",
      identifier: "new",
      description: "",
      projectRoot: "",
      buildType: "backend",
      cloudDeploy: true,
      timeout: 600,
      servers: [],
      targetFiles: [],
      checkStatus: "unchecked",
      createdAt: now,
      updatedAt: now,
    };
  }

  function createNewServer(): ServerConfig {
    return {
      id: `server_${Date.now()}`,
      name: "新服务器",
      host: "",
      port: 22,
      username: "",
      password: "",
      deployDir: "",
      restartScript: "",
      enableRestart: false,
      useSudo: false,
    };
  }

  function createNewTargetFile(): TargetFile {
    return {
      id: `jar_${Date.now()}`,
      localPath: "",
      remoteName: "",
      urlPath: "",
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
