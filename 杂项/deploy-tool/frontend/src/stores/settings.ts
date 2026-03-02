import { defineStore } from "pinia";
import { ref } from "vue";
import type { GlobalSettings, SystemDefaultConfig } from "../types";

export const useSettingsStore = defineStore("settings", () => {
  const globalSettings = ref<GlobalSettings>({
    defaultTimeout: 0,
    logRetentionDays: 0,
    backupEnabled: false,
    notifyOnComplete: false,
    cloudDeploy: false,
    theme: "",
    language: "",
  });

  const systemDefaults = ref<SystemDefaultConfig>({
    jdkPath: "",
    mavenPath: "",
    mavenSettingsPath: "",
    mavenRepoPath: "",
    mavenArgs: [],
  });

  const loading = ref(false);
  const saving = ref(false);

  async function fetchGlobalSettings() {
    loading.value = true;
    try {
      const { GetGlobalSettings } = await import("../../wailsjs/go/app/App");
      const resp = await GetGlobalSettings();
      if (resp.code === 0) {
        globalSettings.value = resp.data;
      }
    } catch (error) {
      console.error("Failed to fetch global settings:", error);
    } finally {
      loading.value = false;
    }
  }

  async function saveGlobalSettings() {
    saving.value = true;
    try {
      const { SaveGlobalSettings } = await import("../../wailsjs/go/app/App");
      const resp = await SaveGlobalSettings({ settings: globalSettings.value });
      if (resp.code !== 0) {
        throw new Error(resp.message || "保存失败");
      }
    } catch (error) {
      console.error("Failed to save global settings:", error);
      throw error;
    } finally {
      saving.value = false;
    }
  }

  async function fetchSystemDefaults() {
    loading.value = true;
    try {
      const { GetSystemDefaults } = await import("../../wailsjs/go/app/App");
      const resp = await GetSystemDefaults();
      if (resp.code === 0) {
        systemDefaults.value = resp.data;
      }
    } catch (error) {
      console.error("Failed to fetch system defaults:", error);
    } finally {
      loading.value = false;
    }
  }

  async function saveSystemDefaults() {
    saving.value = true;
    try {
      const { SaveSystemDefaults } = await import("../../wailsjs/go/app/App");
      const resp = await SaveSystemDefaults({ defaults: systemDefaults.value });
      if (resp.code !== 0) {
        throw new Error(resp.message || "保存失败");
      }
    } catch (error) {
      console.error("Failed to save system defaults:", error);
      throw error;
    } finally {
      saving.value = false;
    }
  }

  async function fetchAll() {
    await Promise.all([fetchGlobalSettings(), fetchSystemDefaults()]);
  }

  async function saveAll() {
    await Promise.all([saveGlobalSettings(), saveSystemDefaults()]);
  }

  return {
    globalSettings,
    systemDefaults,
    loading,
    saving,
    fetchGlobalSettings,
    saveGlobalSettings,
    fetchSystemDefaults,
    saveSystemDefaults,
    fetchAll,
    saveAll,
  };
});
