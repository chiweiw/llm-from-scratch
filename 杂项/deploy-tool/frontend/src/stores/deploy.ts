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

    // Called by the deploy-progress Wails event handler and by fetchProgress.
    setProgress(progress: DeployProgress | null) {
      this.progress = progress;
      if (
        progress &&
        ["success", "failed", "canceled"].includes(progress.status)
      ) {
        this.isDeploying = false;
      }
    },

    async startDeploy(envId: string) {
      this.isDeploying = true;
      this.progress = null;
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
      const resp = await CancelDeploy();
      if (resp.code !== 0) {
        throw new Error(resp.message || "取消失败");
      }
      this.isDeploying = false;
    },

    // Pulls the current progress once (used on mount to recover in-progress state).
    async fetchProgress() {
      const { GetDeployProgress } = await import("../../wailsjs/go/app/App");
      const resp = await GetDeployProgress();
      const p = resp.code === 0 ? (resp.data as DeployProgress | null) : null;
      this.setProgress(p);
      if (p && p.status === "running") {
        this.isDeploying = true;
      }
    },
  },
});
