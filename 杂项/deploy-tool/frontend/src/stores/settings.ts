import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { GlobalSettings, SystemDefaultConfig } from '../types';

export const useSettingsStore = defineStore('settings', () => {
  const globalSettings = ref<GlobalSettings>({
    defaultTimeout: 600,
    logRetentionDays: 30,
    backupEnabled: true,
    notifyOnComplete: true,
    theme: 'system',
    language: 'zh-Hans',
  });

  const systemDefaults = ref<SystemDefaultConfig>({
    jdkPath: '',
    mavenPath: '',
    mavenSettingsPath: '',
    mavenRepoPath: '',
    mavenArgs: 'clean package -DskipTests',
  });

  const loading = ref(false);
  const saving = ref(false);

  async function fetchGlobalSettings() {
    loading.value = true;
    try {
      const { GetGlobalSettings } = await import('../../wailsjs/go/main/App');
      globalSettings.value = await GetGlobalSettings();
    } catch (error) {
      console.error('Failed to fetch global settings:', error);
    } finally {
      loading.value = false;
    }
  }

  async function saveGlobalSettings() {
    saving.value = true;
    try {
      const { SaveGlobalSettings } = await import('../../wailsjs/go/main/App');
      await SaveGlobalSettings(globalSettings.value);
    } catch (error) {
      console.error('Failed to save global settings:', error);
      throw error;
    } finally {
      saving.value = false;
    }
  }

  async function fetchSystemDefaults() {
    loading.value = true;
    try {
      const { GetSystemDefaults } = await import('../../wailsjs/go/main/App');
      systemDefaults.value = await GetSystemDefaults();
    } catch (error) {
      console.error('Failed to fetch system defaults:', error);
    } finally {
      loading.value = false;
    }
  }

  async function saveSystemDefaults() {
    saving.value = true;
    try {
      const { SaveSystemDefaults } = await import('../../wailsjs/go/main/App');
      await SaveSystemDefaults(systemDefaults.value);
    } catch (error) {
      console.error('Failed to save system defaults:', error);
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
