import { defineStore } from 'pinia';
import type { DeployProgress } from '@/types';

export const useDeployStore = defineStore('deploy', {
  state: () => ({
    progress: null as DeployProgress | null,
    isDeploying: false,
    selectedJarIds: [] as string[],
  }),

  getters: {
    currentStepIndex: (state) => {
      if (!state.progress) return -1;
      return state.progress.steps.findIndex((s) => s.status === 'running');
    },
  },

  actions: {
    setSelectedJars(jarIds: string[]) {
      this.selectedJarIds = jarIds;
    },

    async startDeploy(envId: string) {
      this.isDeploying = true;
      const { StartDeploy } = await import('../../wailsjs/go/main/App');
      await StartDeploy(envId, this.selectedJarIds);
    },

    async cancelDeploy() {
      const { CancelDeploy } = await import('../../wailsjs/go/main/App');
      await CancelDeploy();
      this.isDeploying = false;
    },

    async fetchProgress() {
      const { GetDeployProgress } = await import('../../wailsjs/go/main/App');
      this.progress = await GetDeployProgress();
      if (this.progress && ['success', 'failed', 'canceled'].includes(this.progress.status)) {
        this.isDeploying = false;
      }
    },
  },
});
