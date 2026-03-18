<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch, computed } from "vue";
import { ChevronDown } from "lucide-vue-next";
import { useEnvironmentStore } from "@/stores/environment";
import { useDeployStore } from "@/stores/deploy";
import { useSettingsStore } from "@/stores/settings";
import { useWailsEvent } from "@/lib/useWailsEvent";
import type { CheckResult, CheckItem, DeployProgress } from "@/types";

const envStore = useEnvironmentStore();
const deployStore = useDeployStore();
const settingsStore = useSettingsStore();

const selectedEnvId = ref("");
const errorMessage = ref("");
const showSuccess = ref(false);
const pendingShellCommand = ref("");
const logs = ref<Array<{ level: string; message: string; time: string }>>([]);
const logContainer = ref<HTMLElement | null>(null);
const precheckStatus = ref<"idle" | "running" | "success" | "failed">("idle");
const precheckResult = ref<CheckResult | null>(null);
const selectedEnv = computed(() =>
  envStore.environments.find((env) => env.id === selectedEnvId.value)
);

const isLightLog = computed(() => settingsStore.globalSettings.lightLog !== false);

function formatDuration(totalSeconds: number): string {
  if (totalSeconds < 0) totalSeconds = 0;
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  const pad = (v: number) => (v < 10 ? `0${v}` : `${v}`);
  return `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;
}

// A reactive "current time in seconds" that ticks every second while deploying.
// Using a reactive ref (vs bare Date.now()) makes the computed properly re-evaluate.
const nowSeconds = ref(Math.floor(Date.now() / 1000));
let nowTimer: ReturnType<typeof setInterval> | null = null;

watch(
  () => deployStore.isDeploying,
  (running) => {
    if (running) {
      nowSeconds.value = Math.floor(Date.now() / 1000);
      nowTimer = setInterval(() => {
        nowSeconds.value = Math.floor(Date.now() / 1000);
      }, 1000);
    } else {
      if (nowTimer !== null) {
        clearInterval(nowTimer);
        nowTimer = null;
      }
    }
  },
  { immediate: true }
);

// deployDuration: ticks every second while running (via nowSeconds),
// then freezes at the backend-reported final value when done.
const deployDuration = computed(() => {
  const p = deployStore.progress;
  if (!p || !p.startTime) return "";
  if (p.status === "running") {
    return formatDuration(nowSeconds.value - p.startTime);
  }
  return formatDuration(p.elapsedSeconds ?? 0);
});

onMounted(async () => {
  await envStore.fetchEnvironments();
  await settingsStore.fetchGlobalSettings();
  // Recover any in-progress deployment state after a page navigation or reload.
  await deployStore.fetchProgress();
});

function handleLog(data: {
  level: string;
  message: string;
  ts?: string;
  line?: string;
}) {
  const timeStr =
    data.ts || new Date().toLocaleTimeString("zh-CN", { hour12: false });
  if (isLightLog.value && data.level === "DEBUG") {
    return;
  }
  logs.value.push({
    level: data.level,
    message: data.line || data.message,
    time: timeStr,
  });
  updatePendingShellCommand(data.line || data.message);

  const maxLogs = isLightLog.value ? 300 : 1000;
  if (logs.value.length > maxLogs) {
    logs.value = logs.value.slice(-maxLogs);
  }

  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight;
    }
  });
}

function updatePendingShellCommand(message?: string) {
  if (!message) {
    return;
  }
  const marker = "即将执行命令:";
  const idx = message.indexOf(marker);
  if (idx === -1) {
    return;
  }
  const cmd = message.slice(idx + marker.length).trim();
  if (cmd) {
    pendingShellCommand.value = cmd;
  }
}

useWailsEvent("log-event", handleLog);

// Backend pushes deploy-progress events on every progress change.
// This replaces the previous setInterval polling approach.
useWailsEvent("deploy-progress", (progress: DeployProgress) => {
  deployStore.setProgress(progress);
  if (progress.status === "success") {
    showSuccess.value = true;
  }
});

watch(logs, () => {
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight;
    }
  });
});

onUnmounted(() => {
  if (nowTimer !== null) {
    clearInterval(nowTimer);
    nowTimer = null;
  }
});

async function startDeploy() {
  errorMessage.value = "";
  showSuccess.value = false;
  pendingShellCommand.value = "";
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
        message: `环境检查通过，但存在 ${warnCount} 条通过，将继续部署`,
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
    // Progress updates arrive via the "deploy-progress" Wails event.
  } catch (error) {
    errorMessage.value =
      error instanceof Error ? error.message : "启动部署失败";
    console.error("部署启动失败:", error);
  }
}

async function cancelDeploy() {
  try {
    await deployStore.cancelDeploy();
    showSuccess.value = false;
    pendingShellCommand.value = "";
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : "取消失败";
  }
}
</script>

<template>
  <div class="h-full p-6 flex flex-col">
    <!-- Header -->
    <div class="mb-5 shrink-0">
      <h1 class="text-3xl font-bold bg-gradient-to-r from-primary to-blue-600 bg-clip-text text-transparent w-fit pb-1">部署中心</h1>
      <p class="text-muted-foreground mt-1.5 text-sm">选择环境并执行应用部署任务</p>
    </div>

    <!-- Alerts -->
    <div
      v-if="errorMessage"
      class="mb-3 shrink-0 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
    >
      {{ errorMessage }}
    </div>
    <div
      v-if="showSuccess"
      class="mb-3 shrink-0 rounded-md border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800"
    >
      部署成功！
    </div>

    <!-- Env selector row -->
    <div class="mb-4 shrink-0">
      <div class="flex items-center gap-3 flex-wrap">
        <select
          v-model="selectedEnvId"
          class="rounded-md border border-input bg-background px-3 py-2 text-sm max-w-xs focus:outline-none focus:ring-1 focus:ring-ring"
        >
          <option value="">请选择环境</option>
          <option
            v-for="env in envStore.environments"
            :key="env.id"
            :value="env.id"
          >
            {{ env.buildType === "frontend" ? "前端" : "后端" }} |
            {{ env.name }}
          </option>
        </select>
        <span
          v-if="selectedEnv"
          class="rounded-full px-2.5 py-0.5 text-xs font-medium"
          :class="{
            'bg-purple-100 text-purple-700': selectedEnv.buildType === 'frontend',
            'bg-blue-100 text-blue-700': !selectedEnv.buildType || selectedEnv.buildType === 'backend',
          }"
        >
          {{ selectedEnv.buildType === "frontend" ? "前端环境" : "后端环境" }}
        </span>
        <button
          @click="startDeploy"
          :disabled="!selectedEnvId || deployStore.isDeploying"
          class="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {{ deployStore.isDeploying ? "部署中..." : "开始部署" }}
        </button>
        <button
          v-if="deployStore.isDeploying"
          @click="cancelDeploy"
          :title="deployStore.progress?.currentStep !== 'Maven 打包' ? '仅在 Maven 打包阶段支持终止' : ''"
          class="rounded-md border px-4 py-2 text-sm font-medium transition-colors"
          :class="deployStore.progress?.currentStep === 'Maven 打包'
            ? 'border-red-300 text-red-600 hover:bg-red-50'
            : 'border-input text-muted-foreground cursor-not-allowed opacity-60'"
        >
          {{ deployStore.progress?.currentStep === 'Maven 打包' ? '终止打包' : '暂停部署' }}
        </button>
        <div v-if="deployDuration" class="ml-auto flex items-center gap-2 px-3 py-1.5 rounded-md bg-muted/40 border border-border/40">
          <span class="text-xs text-muted-foreground font-medium">耗时</span>
          <span class="font-mono text-lg font-bold text-primary tabular-nums tracking-tight">{{ deployDuration }}</span>
        </div>
      </div>
    </div>

    <!-- Two-column content area: 30% controls / 70% logs -->
    <div class="flex-1 min-h-0 grid grid-cols-[3fr_7fr] gap-5">
      <!-- Left column: server info + progress (fills same height as right col) -->
      <div class="flex flex-col gap-4 min-h-0">
        <!-- Server info -->
        <div
          v-if="selectedEnv"
          class="shrink-0 rounded-lg border border-blue-200 bg-blue-50/60 p-3.5 text-sm"
        >
          <div class="font-medium text-blue-800 mb-1.5">将影响的机器与部署目录</div>
          <div v-if="(selectedEnv.servers?.length ?? 0) === 0" class="text-blue-700/70">
            当前环境未配置服务器
          </div>
          <div v-else class="space-y-1">
            <div
              v-for="(server, idx) in selectedEnv.servers || []"
              :key="server.id || idx"
              class="flex flex-wrap items-center gap-2 text-blue-900"
            >
              <span class="font-medium">{{ server.name || `服务器 ${idx + 1}` }}</span>
              <span class="text-xs text-blue-600 font-mono">{{ server.host }}:{{ server.port }}</span>
              <span class="text-xs text-blue-600">→ {{ server.deployDir || "未设置" }}</span>
            </div>
          </div>
        </div>

        <!-- Progress panel: fills remaining height to match log panel -->
        <div class="rounded-lg border flex flex-col flex-1 min-h-0">
          <!-- Panel header -->
          <div class="flex items-center justify-between px-5 py-3.5 border-b shrink-0">
            <h3 class="text-sm font-semibold">部署进度</h3>
            <span
              v-if="deployStore.progress"
              class="text-xs font-mono font-bold tabular-nums"
              :class="{
                'text-green-600': deployStore.progress.status === 'success',
                'text-red-500': deployStore.progress.status === 'failed',
                'text-primary': !['success','failed'].includes(deployStore.progress.status),
              }"
            >{{ deployStore.progress.totalProgress }}%</span>
          </div>

          <!-- Active progress: steps with ChevronDown arrows as connectors -->
          <div v-if="deployStore.progress" class="flex-1 min-h-0 overflow-y-auto px-5 py-4">
            <template v-for="(step, idx) in deployStore.progress.steps" :key="step.name">
              <!-- Step row -->
              <div class="flex items-start gap-3">
                <!-- Status dot -->
                <div
                  class="mt-0.5 w-6 h-6 rounded-full flex items-center justify-center text-xs font-semibold shrink-0 border-2 bg-background transition-colors duration-300"
                  :class="{
                    'border-muted/60 text-muted-foreground/30': step.status === 'pending',
                    'border-primary text-primary': step.status === 'running',
                    'border-green-500 bg-green-50 text-green-600': step.status === 'success',
                    'border-red-500 bg-red-50 text-red-600': step.status === 'failed',
                    'border-muted bg-muted/30 text-muted-foreground/40': step.status === 'skipped',
                  }"
                >
                  <span v-if="step.status === 'success'">✓</span>
                  <span v-else-if="step.status === 'failed'">✗</span>
                  <span v-else-if="step.status === 'skipped'">—</span>
                  <span
                    v-else-if="step.status === 'running'"
                    class="inline-block w-2 h-2 rounded-full bg-primary animate-pulse"
                  ></span>
                  <span v-else class="inline-block w-1.5 h-1.5 rounded-full bg-muted/60"></span>
                </div>
                <!-- Step content -->
                <div class="flex-1 min-w-0">
                  <div class="flex items-center justify-between gap-2">
                    <span
                      class="text-sm font-medium leading-6 transition-colors duration-200"
                      :class="{
                        'text-foreground': step.status === 'running' || step.status === 'success' || step.status === 'failed',
                        'text-muted-foreground/50': step.status === 'pending' || step.status === 'skipped',
                      }"
                    >{{ step.name }}</span>
                    <span
                      v-if="step.status === 'running'"
                      class="shrink-0 text-[10px] font-medium text-primary bg-primary/10 px-1.5 py-0.5 rounded-md animate-pulse"
                    >进行中</span>
                  </div>
                  <div
                    v-if="step.message"
                    class="text-xs leading-relaxed"
                    :class="step.status === 'failed' ? 'text-red-600' : 'text-muted-foreground'"
                  >{{ step.message }}</div>
                </div>
              </div>
              <!-- ChevronDown connector (arrow indicating downward flow) -->
              <div
                v-if="idx < deployStore.progress.steps.length - 1"
                class="flex justify-start pl-[9px] py-0.5"
              >
                <ChevronDown
                  :size="14"
                  :class="{
                    'text-green-400': step.status === 'success',
                    'text-primary/50': step.status === 'running',
                    'text-muted-foreground/20': step.status === 'pending' || step.status === 'skipped',
                    'text-red-300': step.status === 'failed',
                  }"
                />
              </div>
            </template>

            <!-- Error message -->
            <div
              v-if="deployStore.progress.status === 'failed' && deployStore.progress.errorMessage"
              class="mt-4 rounded-md border border-red-200 bg-red-50 p-3"
            >
              <div class="text-xs font-semibold text-red-800 mb-1">错误信息</div>
              <div class="text-xs text-red-700 leading-relaxed font-mono break-all">{{ deployStore.progress.errorMessage }}</div>
            </div>
          </div>

          <!-- Precheck status (no active progress yet) -->
          <div v-else-if="precheckStatus !== 'idle'">
            <div class="flex gap-3">
              <div class="flex flex-col items-center shrink-0">
                <div
                  class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-semibold border-2"
                  :class="{
                    'border-primary bg-primary/10 text-primary': precheckStatus === 'running',
                    'border-green-500 bg-green-500 text-white': precheckStatus === 'success',
                    'border-red-500 bg-red-500 text-white': precheckStatus === 'failed',
                  }"
                >
                  <span v-if="precheckStatus === 'success'">✓</span>
                  <span v-else-if="precheckStatus === 'failed'">✗</span>
                  <span
                    v-else
                    class="inline-block w-2 h-2 rounded-full bg-primary animate-pulse"
                  ></span>
                </div>
              </div>
              <div class="flex-1">
                <div class="text-sm font-medium leading-6">环境检查</div>
                <div v-if="precheckStatus === 'running'" class="text-xs text-muted-foreground">
                  正在进行环境检查...
                </div>
                <div v-else-if="precheckStatus === 'failed'" class="text-xs text-red-600">
                  环境检查未通过
                </div>
                <div v-else-if="precheckStatus === 'success'" class="text-xs text-green-600">
                  检查通过
                  <template v-if="(precheckResult?.checks?.filter((c) => c.status === 'warning').length || 0) > 0">
                    （存在 {{ precheckResult?.checks?.filter((c) => c.status === 'warning').length }} 条警告，将继续部署）
                  </template>
                </div>
              </div>
            </div>

            <div
              v-if="precheckResult && precheckResult.checks && precheckResult.checks.length"
              class="mt-3 space-y-2"
            >
              <div
                v-for="(item, idx) in precheckResult.checks"
                :key="idx"
                class="flex items-start gap-2 rounded-md p-2 text-xs border"
                :class="{
                  'border-red-200 bg-red-50': item.status === 'error',
                  'border-yellow-200 bg-yellow-50': item.status === 'warning',
                  'border-green-200 bg-green-50': item.status === 'pass',
                }"
              >
                <span v-if="item.status === 'pass'" class="mt-0.5 text-green-600">✓</span>
                <span v-else-if="item.status === 'error'" class="mt-0.5 text-red-600">✗</span>
                <span v-else class="mt-0.5 text-yellow-600">⚠</span>
                <div class="flex-1">
                  <div class="font-medium">{{ item.name }}</div>
                  <div
                    v-if="item.message"
                    :class="item.status === 'error' ? 'text-red-700' : 'text-muted-foreground'"
                  >
                    {{ item.message }}
                  </div>
                </div>
              </div>
              <div
                v-if="precheckStatus === 'failed'"
                class="mt-2 rounded-md border border-red-300 bg-red-100 p-2.5 text-xs text-red-800"
              >
                请根据以上错误信息修改配置后重新自检
              </div>
            </div>
          </div>

          <!-- Empty state -->
          <div v-else class="text-center text-muted-foreground py-10 text-sm">
            选择环境后点击「开始部署」
          </div>
        </div>
      </div>

      <!-- Right column: log panel + shell command -->
      <div class="flex flex-col gap-3 min-h-0">
        <!-- Log panel -->
        <div class="rounded-lg border flex flex-col flex-1 min-h-0">
          <div class="flex items-center justify-between px-4 py-2.5 border-b shrink-0">
            <h3 class="text-sm font-semibold">实时日志</h3>
            <button
              @click="logs = []"
              class="text-xs text-muted-foreground hover:text-foreground transition-colors"
            >
              清空
            </button>
          </div>
          <div
            ref="logContainer"
            class="flex-1 min-h-0 overflow-y-auto bg-[#0d1117] rounded-b-lg p-3 font-mono text-xs leading-5"
          >
            <div v-if="logs.length === 0" class="text-center text-gray-600 py-8">
              暂无日志
            </div>
            <div v-else>
              <div
                v-for="(log, index) in logs"
                :key="index"
                class="mb-0.5"
                :class="{
                  'text-blue-400': log.level === 'INFO',
                  'text-yellow-400': log.level === 'WARN',
                  'text-red-400': log.level === 'ERROR',
                  'text-gray-500': log.level === 'DEBUG',
                }"
              >
                <span class="text-gray-600 select-none">[{{ log.time }}]</span>
                <span class="font-semibold ml-1.5 text-gray-400">[{{ log.level }}]</span>
                <span class="ml-1.5">{{ log.message }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Shell command preview -->
        <div
          v-if="pendingShellCommand"
          class="shrink-0 rounded-lg border border-amber-300 bg-amber-50 p-3"
        >
          <div class="text-xs font-medium text-amber-700 mb-1.5">即将执行的 Shell 命令</div>
          <pre class="text-xs text-amber-900 whitespace-pre-wrap break-all font-mono leading-4">{{ pendingShellCommand }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>
