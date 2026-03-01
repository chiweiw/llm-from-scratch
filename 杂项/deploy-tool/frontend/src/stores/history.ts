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
        const { GetDeployHistory } = await import("../../wailsjs/go/main/App");
        this.histories = await GetDeployHistory();
      } finally {
        this.loading = false;
      }
    },
  },
});
