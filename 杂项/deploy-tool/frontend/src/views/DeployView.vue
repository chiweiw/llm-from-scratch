<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from "vue";
import { useEnvironmentStore } from "@/stores/environment";
import { useDeployStore } from "@/stores/deploy";
import { EventsOn } from "../../wailsjs/runtime/runtime";

const envStore = useEnvironmentStore();
const deployStore = useDeployStore();

const selectedEnvId = ref("");
const errorMessage = ref("");
const showSuccess = ref(false);
const logs = ref<Array<{ level: string; message: string; time: string }>>([]);
const logContainer = ref<HTMLElement | null>(null);
let progressInterval: number | null = null;

onMounted(async () => {
  await envStore.fetchEnvironments();

  EventsOn("log-event", (data: { level: string; message: string }) => {
    const now = new Date();
    const timeStr = now.toLocaleTimeString("zh-CN", { hour12: false });
    logs.value.push({
      level: data.level,
      message: data.message,
      time: timeStr,
    });

    if (logs.value.length > 1000) {
      logs.value = logs.value.slice(-1000);
    }

    nextTick(() => {
      if (logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight;
      }
    });
  });
});

watch(logs, () => {
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight;
    }
  });
});

onUnmounted(() => {
  if (progressInterval) {
    clearInterval(progressInterval);
  }
});

async function startDeploy() {
  errorMessage.value = "";
  showSuccess.value = false;
  logs.value = [];

  if (!selectedEnvId.value) {
    errorMessage.value = "请先选择环境";
    return;
  }

  try {
    await deployStore.startDeploy(selectedEnvId.value);
    progressInterval = window.setInterval(() => {
      deployStore.fetchProgress();

      if (deployStore.progress && deployStore.progress.status === "success") {
        showSuccess.value = true;
        clearInterval(progressInterval);
        progressInterval = null;
      }
    }, 500);
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "启动部署失败";
    console.error("部署启动失败:", error);
  }
}

function cancelDeploy() {
  deployStore.cancelDeploy();
  if (progressInterval) {
    clearInterval(progressInterval);
    progressInterval = null;
  }
  showSuccess.value = false;
}
</script>

<template>
  <div class="h-full p-6">
    <h1 class="text-2xl font-bold mb-6">部署中心</h1>

    <div v-if="errorMessage" class="mb-4 rounded-md border border-red-200 bg-red-50 p-4 text-red-800">
      {{ errorMessage }}
    </div>

    <div v-if="showSuccess" class="mb-4 rounded-md border border-green-200 bg-green-50 p-4 text-green-800">
      部署成功！
    </div>

    <div class="mb-6">
      <label class="text-sm text-muted-foreground mb-2 block">选择环境</label>
      <select
        v-model="selectedEnvId"
        class="w-full max-w-md rounded-md border bg-background px-3 py-2"
      >
        <option value="">请选择环境</option>
        <option
          v-for="env in envStore.environments"
          :key="env.id"
          :value="env.id"
        >
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
          <div
            v-for="step in deployStore.progress.steps"
            :key="step.name"
            class="flex items-center gap-3"
          >
            <div
              class="w-6 h-6 rounded-full flex items-center justify-center text-xs"
              :class="{
                'bg-muted text-muted-foreground': step.status === 'pending',
                'bg-primary text-primary-foreground': step.status === 'running',
                'bg-green-500 text-white': step.status === 'success',
                'bg-red-500 text-white': step.status === 'failed',
                'bg-gray-400 text-white': step.status === 'skipped',
              }"
            >
              {{
                step.status === "success"
                  ? "✓"
                  : step.status === "failed"
                  ? "✗"
                  : step.status === "skipped"
                  ? "⊘"
                  : ""
              }}
            </div>
            <div class="flex-1">
              <div class="font-medium">{{ step.name }}</div>
              <div v-if="step.message" class="text-sm text-muted-foreground">
                {{ step.message }}
              </div>
            </div>
          </div>
        </div>

        <div v-if="deployStore.progress.errorMessage" class="mt-4 rounded-md border border-red-200 bg-red-50 p-4 text-red-800">
          <div class="font-medium">错误信息</div>
          <div class="text-sm">{{ deployStore.progress.errorMessage }}</div>
        </div>
      </div>
      <div v-else class="text-center text-muted-foreground py-8">
        选择环境后开始部署
      </div>
    </div>

    <div class="rounded-lg border p-6 mb-6">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg font-semibold">实时日志</h3>
        <button
          @click="logs = []"
          class="text-sm text-muted-foreground hover:text-foreground"
        >
          清空日志
        </button>
      </div>
      <div
        ref="logContainer"
        class="bg-gray-900 text-gray-100 rounded-md p-4 h-64 overflow-y-auto font-mono text-sm"
      >
        <div v-if="logs.length === 0" class="text-center text-gray-500 py-8">
          暂无日志
        </div>
        <div v-else>
          <div
            v-for="(log, index) in logs"
            :key="index"
            class="mb-1"
            :class="{
              'text-blue-400': log.level === 'INFO',
              'text-yellow-400': log.level === 'WARN',
              'text-red-400': log.level === 'ERROR',
              'text-gray-400': log.level === 'DEBUG',
            }"
          >
            <span class="text-gray-500">[{{ log.time }}]</span>
            <span class="font-bold ml-2">[{{ log.level }}]</span>
            <span class="ml-2">{{ log.message }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="flex gap-4">
      <button
        @click="startDeploy"
        :disabled="!selectedEnvId || deployStore.isDeploying"
        class="rounded-md bg-primary px-6 py-2 text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {{ deployStore.isDeploying ? "部署中..." : "开始部署" }}
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
