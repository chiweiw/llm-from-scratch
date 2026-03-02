import { defineStore } from "pinia";
import type { DeployHistory } from "@/types";

export const useHistoryStore = defineStore("history", {
  state: () => ({
    histories: [] as DeployHistory[],
    loading: false,
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
  },
});
