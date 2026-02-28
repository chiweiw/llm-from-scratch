<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useEnvironmentStore } from '@/stores/environment';
import { useDeployStore } from '@/stores/deploy';

const envStore = useEnvironmentStore();
const deployStore = useDeployStore();

const selectedEnvId = ref('');
let progressInterval: number | null = null;

onMounted(async () => {
  await envStore.fetchEnvironments();
});

onUnmounted(() => {
  if (progressInterval) {
    clearInterval(progressInterval);
  }
});

function startDeploy() {
  if (selectedEnvId.value) {
    deployStore.startDeploy(selectedEnvId.value);
    progressInterval = window.setInterval(() => {
      deployStore.fetchProgress();
    }, 500);
  }
}

function cancelDeploy() {
  deployStore.cancelDeploy();
  if (progressInterval) {
    clearInterval(progressInterval);
  }
}
</script>

<template>
  <div class="h-full p-6">
    <h1 class="text-2xl font-bold mb-6">部署中心</h1>
    
    <div class="mb-6">
      <label class="text-sm text-muted-foreground mb-2 block">选择环境</label>
      <select v-model="selectedEnvId" class="w-full max-w-md rounded-md border bg-background px-3 py-2">
        <option value="">请选择环境</option>
        <option v-for="env in envStore.environments" :key="env.id" :value="env.id">
          {{ env.name }}
        </option>
      </select>
    </div>

    <div class="rounded-lg border p-6 mb-6">
      <h3 class="text-lg font-semibold mb-4">部署进度</h3>
      <div v-if="deployStore.progress">
        <div class="mb-4">
          <div class="flex justify-between text-sm mb-1">
            <span>总进度</span>
            <span>{{ deployStore.progress.totalProgress }}%</span>
          </div>
          <div class="h-2 rounded-full bg-muted overflow-hidden">
            <div 
              class="h-full bg-primary transition-all duration-300"
              :style="{ width: `${deployStore.progress.totalProgress}%` }"
            ></div>
          </div>
        </div>
        <div class="space-y-3">
          <div v-for="step in deployStore.progress.steps" :key="step.name" class="flex items-center gap-3">
            <div 
              class="w-6 h-6 rounded-full flex items-center justify-center text-xs"
              :class="{
                'bg-muted text-muted-foreground': step.status === 'pending',
                'bg-primary text-primary-foreground': step.status === 'running',
                'bg-green-500 text-white': step.status === 'success',
                'bg-red-500 text-white': step.status === 'failed',
              }"
            >
              {{ step.status === 'success' ? '✓' : step.status === 'failed' ? '✗' : '' }}
            </div>
            <div class="flex-1">
              <div class="font-medium">{{ step.name }}</div>
              <div v-if="step.message" class="text-sm text-muted-foreground">{{ step.message }}</div>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="text-center text-muted-foreground py-8">
        选择环境后开始部署
      </div>
    </div>

    <div class="flex gap-4">
      <button 
        @click="startDeploy"
        :disabled="!selectedEnvId || deployStore.isDeploying"
        class="rounded-md bg-primary px-6 py-2 text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        开始部署
      </button>
      <button 
        v-if="deployStore.isDeploying"
        @click="cancelDeploy"
        class="rounded-md border px-6 py-2 hover:bg-accent"
      >
        取消部署
      </button>
    </div>
  </div>
</template>
