import { defineStore } from "pinia";
import type { DeployProgress } from "@/types";

export const useDeployStore = defineStore("deploy", {
  state: () => ({
    progress: null as DeployProgress | null,
    isDeploying: false,
    selectedJarIds: [] as string[],
  }),

  getters: {
    currentStepIndex: (state) => {
      if (!state.progress) return -1;
      return state.progress.steps.findIndex((s) => s.status === "running");
    },
  },

  actions: {
    setSelectedJars(jarIds: string[]) {
      this.selectedJarIds = jarIds;
    },

    async startDeploy(envId: string) {
      this.isDeploying = true;
      const { StartDeploy } = await import("../../wailsjs/go/app/App");
      const resp = await StartDeploy({
        environmentId: envId,
        jarIds: this.selectedJarIds,
      });
      if (resp.code !== 0) {
        this.isDeploying = false;
        throw new Error(resp.message || "启动部署失败");
      }
    },

    async cancelDeploy() {
      const { CancelDeploy } = await import("../../wailsjs/go/app/App");
      await CancelDeploy();
      this.isDeploying = false;
    },

    async fetchProgress() {
      const { GetDeployProgress } = await import("../../wailsjs/go/app/App");
      const resp = await GetDeployProgress();
      this.progress = resp.code === 0 ? resp.data : null;
      if (
        this.progress &&
        ["success", "failed", "canceled"].includes(this.progress.status)
      ) {
        this.isDeploying = false;
      }
    },
  },
});
