import { defineStore } from 'pinia';
import type { DeployHistory, GlobalSettings } from '@/types';

export const useHistoryStore = defineStore('history', {
  state: () => ({
    histories: [] as DeployHistory[],
    currentHistory: null as DeployHistory | null,
    loading: false,
  }),

  actions: {
    async fetchHistories(filter = {}) {
      this.loading = true;
      try {
        const { GetDeployHistory } = await import('../../wailsjs/go/main/App');
        this.histories = await GetDeployHistory(filter);
      } finally {
        this.loading = false;
      }
    },

    async fetchHistoryDetail(id: string) {
      const { GetHistoryDetail } = await import('../../wailsjs/go/main/App');
      this.currentHistory = await GetHistoryDetail(id);
    },
  },
});

export const useSettingsStore = defineStore('settings', {
  state: () => ({
    settings: null as GlobalSettings | null,
    loading: false,
  }),

  actions: {
    async fetchSettings() {
      this.loading = true;
      try {
        const { GetGlobalSettings } = await import('../../wailsjs/go/main/App');
        this.settings = await GetGlobalSettings();
      } finally {
        this.loading = false;
      }
    },

    async saveSettings(settings: GlobalSettings) {
      const { SaveGlobalSettings } = await import('../../wailsjs/go/main/App');
      await SaveGlobalSettings(settings);
      this.settings = settings;
    },
  },
});
