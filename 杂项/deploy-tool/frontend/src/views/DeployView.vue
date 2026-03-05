<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from "vue";
import { useEnvironmentStore } from "@/stores/environment";
import { useDeployStore } from "@/stores/deploy";
import { useWailsEvent } from "@/lib/useWailsEvent";
import type { CheckResult, CheckItem } from "@/types";

const envStore = useEnvironmentStore();
const deployStore = useDeployStore();

const selectedEnvId = ref("");
const errorMessage = ref("");
const showSuccess = ref(false);
const logs = ref<Array<{ level: string; message: string; time: string }>>([]);
const logContainer = ref<HTMLElement | null>(null);
let progressInterval: number | undefined;
const precheckStatus = ref<"idle" | "running" | "success" | "failed">("idle");
const precheckResult = ref<CheckResult | null>(null);

onMounted(async () => {
  await envStore.fetchEnvironments();
});

function handleLog(data: {
  level: string;
  message: string;
  ts?: string;
  line?: string;
}) {
  const timeStr =
    data.ts || new Date().toLocaleTimeString("zh-CN", { hour12: false });
  logs.value.push({
    level: data.level,
    message: data.line || data.message,
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
}

useWailsEvent("log-event", handleLog);

watch(logs, () => {
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight;
    }
  });
});

onUnmounted(() => {
  if (progressInterval !== null) {
    clearInterval(progressInterval);
  }
});

async function startDeploy() {
  errorMessage.value = "";
  showSuccess.value = false;
  logs.value = [];
  deployStore.progress = null;
  precheckStatus.value = "idle";
  precheckResult.value = null;

  if (!selectedEnvId.value) {
    errorMessage.value = "请先选择环境";
    return;
  }

  try {
    precheckStatus.value = "running";
    logs.value.push({
      level: "INFO",
      message: "开始环境检查...",
      time: new Date().toLocaleTimeString("zh-CN", { hour12: false }),
    });
    const result = await envStore.checkEnvironment(selectedEnvId.value);
    precheckResult.value = result;
    if (!result?.success) {
      precheckStatus.value = "failed";
      logs.value.push({
        level: "ERROR",
        message: "环境检查未通过",
        time: new Date().toLocaleTimeString("zh-CN", { hour12: false }),
      });
      if (result.checks && result.checks.length) {
        for (const item of result.checks) {
          const level =
            item.status === "error"
              ? "ERROR"
              : item.status === "warning"
              ? "WARN"
              : "INFO";
          logs.value.push({
            level,
            message: `${item.name}${item.message ? " - " + item.message : ""}`,
            time: new Date().toLocaleTimeString("zh-CN", { hour12: false }),
          });
        }
      }
      return;
    }
    precheckStatus.value = "success";
    const warnCount =
      result.checks?.filter((c: CheckItem) => c.status === "warning").length ||
      0;
    if (warnCount > 0) {
      logs.value.push({
        level: "WARN",
        message: `环境检查通过，但存在 ${warnCount} 条警告，将继续部署`,
        time: new Date().toLocaleTimeString("zh-CN", { hour12: false }),
      });
    } else {
      logs.value.push({
        level: "INFO",
        message: "环境检查通过，开始部署...",
        time: new Date().toLocaleTimeString("zh-CN", { hour12: false }),
      });
    }
    await deployStore.startDeploy(selectedEnvId.value);
    progressInterval = window.setInterval(() => {
      deployStore.fetchProgress();

      if (deployStore.progress && deployStore.progress.status === "success") {
        showSuccess.value = true;
        if (progressInterval !== undefined) {
          clearInterval(progressInterval);
          progressInterval = undefined;
        }
      }
    }, 500);
  } catch (error) {
    errorMessage.value =
      error instanceof Error ? error.message : "启动部署失败";
    console.error("部署启动失败:", error);
  }
}

function cancelDeploy() {
  deployStore.cancelDeploy();
  if (progressInterval !== undefined) {
    clearInterval(progressInterval);
    progressInterval = undefined;
  }
  showSuccess.value = false;
}
</script>

<template>
  <div class="h-full p-6">
    <h1 class="text-2xl font-bold mb-6">部署中心</h1>

    <div
      v-if="errorMessage"
      class="mb-4 rounded-md border border-red-200 bg-red-50 p-4 text-red-800"
    >
      {{ errorMessage }}
    </div>

    <div
      v-if="showSuccess"
      class="mb-4 rounded-md border border-green-200 bg-green-50 p-4 text-green-800"
    >
      部署成功！
    </div>

    <div class="mb-6">
      <label class="text-sm text-muted-foreground mb-2 block">选择环境</label>
      <div class="flex items-center gap-3">
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
        <button
          @click="startDeploy"
          :disabled="!selectedEnvId || deployStore.isDeploying"
          class="rounded-md bg-primary px-4 py-2 text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ deployStore.isDeploying ? "部署中..." : "开始部署" }}
        </button>
        <button
          v-if="deployStore.isDeploying"
          @click="cancelDeploy"
          class="rounded-md border px-4 py-2 hover:bg-accent"
        >
          暂停部署
        </button>
      </div>
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

        <div
          v-if="deployStore.progress.errorMessage"
          class="mt-4 rounded-md border border-red-200 bg-red-50 p-4 text-red-800"
        >
          <div class="font-medium">错误信息</div>
          <div class="text-sm">{{ deployStore.progress.errorMessage }}</div>
        </div>
      </div>
      <div v-else>
        <div v-if="precheckStatus !== 'idle'">
          <div class="space-y-3">
            <div class="flex items-center gap-3">
              <div
                class="w-6 h-6 rounded-full flex items-center justify-center text-xs"
                :class="{
                  'bg-primary text-primary-foreground':
                    precheckStatus === 'running',
                  'bg-green-500 text-white': precheckStatus === 'success',
                  'bg-red-500 text-white': precheckStatus === 'failed',
                }"
              >
                {{
                  precheckStatus === "success"
                    ? "✓"
                    : precheckStatus === "failed"
                    ? "✗"
                    : ""
                }}
              </div>
              <div class="flex-1">
                <div class="font-medium">环境检查</div>
                <div
                  v-if="precheckStatus === 'running'"
                  class="text-sm text-muted-foreground"
                >
                  正在进行环境检查...
                </div>
                <div
                  v-else-if="precheckStatus === 'failed'"
                  class="text-sm text-red-700"
                >
                  环境检查未通过
                </div>
                <div
                  v-else-if="precheckStatus === 'success'"
                  class="text-sm text-green-700"
                >
                  检查通过
                  <template
                    v-if="
                      (precheckResult?.checks?.filter(
                        (c) => c.status === 'warning'
                      ).length || 0) > 0
                    "
                  >
                    （存在
                    {{
                      precheckResult?.checks?.filter(
                        (c) => c.status === "warning"
                      ).length
                    }}
                    条警告，将继续部署）
                  </template>
                </div>
              </div>
            </div>
            <div
              v-if="
                precheckResult &&
                precheckResult.checks &&
                precheckResult.checks.length
              "
              class="mt-2 space-y-2"
            >
              <div
                v-for="(item, idx) in precheckResult.checks"
                :key="idx"
                class="flex items-start gap-2 rounded bg-white p-2 text-sm border"
                :class="{
                  'border-red-200 bg-red-50': item.status === 'error',
                  'border-yellow-200 bg-yellow-50': item.status === 'warning',
                  'border-green-200 bg-green-50': item.status === 'pass',
                }"
              >
                <span
                  v-if="item.status === 'pass'"
                  class="mt-0.5 text-green-600"
                  >✓</span
                >
                <span
                  v-else-if="item.status === 'error'"
                  class="mt-0.5 text-red-600"
                  >✗</span
                >
                <span v-else class="mt-0.5 text-yellow-600">⚠</span>
                <div class="flex-1">
                  <div class="font-medium">{{ item.name }}</div>
                  <div
                    v-if="item.message"
                    :class="
                      item.status === 'error' ? 'text-red-700' : 'text-gray-700'
                    "
                  >
                    {{ item.message }}
                  </div>
                </div>
              </div>
              <div
                v-if="precheckStatus === 'failed'"
                class="mt-3 rounded border border-red-300 bg-red-100 p-3 text-red-800"
              >
                请根据以上错误信息修改配置后重新自检
              </div>
            </div>
          </div>
        </div>
        <div v-else class="text-center text-muted-foreground py-8">
          选择环境后开始部署
        </div>
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

    <!-- 操作按钮已移动到环境选择处 -->
  </div>
</template>
