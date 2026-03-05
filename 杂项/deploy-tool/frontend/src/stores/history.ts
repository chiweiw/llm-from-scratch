import { defineStore } from "pinia";
import type { DeployHistory, DeployLog } from "@/types";

export const useHistoryStore = defineStore("history", {
  state: () => ({
    histories: [] as DeployHistory[],
    loading: false,
    selected: null as DeployHistory | null,
    logs: [] as DeployLog[],
    loadingLogs: false,
  }),

  actions: {
    async fetchHistories() {
      this.loading = true;
      try {
        const { GetDeployHistory } = await import("../../wailsjs/go/app/App");
        const resp = await GetDeployHistory();
        this.histories = resp.code === 0 ? resp.data : [];
      } finally {
        this.loading = false;
      }
    },
    setSelected(history: DeployHistory | null) {
      this.selected = history;
      this.logs = [];
    },
    async fetchLogs(id: string) {
      this.loadingLogs = true;
      try {
        const { GetDeployLogs } = await import("../../wailsjs/go/app/App");
        const resp = await GetDeployLogs(id);
        this.logs = resp.code === 0 ? resp.data : [];
      } finally {
        this.loadingLogs = false;
      }
    },
  },
});
