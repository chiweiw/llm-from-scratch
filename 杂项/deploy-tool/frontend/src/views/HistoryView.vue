<script setup lang="ts">
import { onMounted, ref, watch, computed } from "vue";
import { useHistoryStore } from "@/stores/history";
import DateRange from "@/components/DateRange.vue";

const historyStore = useHistoryStore();
const selectedId = ref("");

const toYMD = (d: Date) => {
  const y = d.getFullYear();
  const m = d.getMonth() + 1;
  const dd = d.getDate();
  const mmStr = m < 10 ? "0" + m : "" + m;
  const ddStr = dd < 10 ? "0" + dd : "" + dd;
  return `${y}-${mmStr}-${ddStr}`;
};

onMounted(async () => {
  await historyStore.fetchHistories();
  const today = new Date();
  const end = new Date(today.getFullYear(), today.getMonth(), today.getDate());
  const start = new Date(end);
  start.setDate(end.getDate() - 6);
  dateRange.value = { start: toYMD(start), end: toYMD(end) };
});

function formatTime(timestamp: number): string {
  return new Date(timestamp * 1000).toLocaleString("zh-CN");
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}秒`;
  const minutes = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${minutes}分${secs}秒`;
}

watch(selectedId, async (val) => {
  const h = historyStore.histories.find((x) => x.id === val) || null;
  historyStore.setSelected(h);
  if (h) {
    await historyStore.fetchLogs(h.id);
  }
});

const filterEnv = ref("");
const filterStatus = ref("");
const dateRange = ref<{ start: string; end: string }>({ start: "", end: "" });

const envOptions = computed(() => {
  const set = new Set<string>();
  historyStore.histories.forEach((h) => set.add(h.environmentName));
  return Array.from(set);
});

function parseDateToUnixStart(s: string): number | null {
  if (!s) return null;
  const parts = s.split("-");
  if (parts.length !== 3) return null;
  const y = Number(parts[0]);
  const m = Number(parts[1]);
  const d = Number(parts[2]);
  const dt = new Date(y, m - 1, d, 0, 0, 0);
  if (isNaN(dt.getTime())) return null;
  return Math.floor(dt.getTime() / 1000);
}
function parseDateToUnixEnd(s: string): number | null {
  if (!s) return null;
  const parts = s.split("-");
  if (parts.length !== 3) return null;
  const y = Number(parts[0]);
  const m = Number(parts[1]);
  const d = Number(parts[2]);
  const dt = new Date(y, m - 1, d, 23, 59, 59);
  if (isNaN(dt.getTime())) return null;
  return Math.floor(dt.getTime() / 1000);
}

const filteredHistories = computed(() => {
  const startTs = parseDateToUnixStart(dateRange.value.start);
  const endTs = parseDateToUnixEnd(dateRange.value.end);
  return historyStore.histories.filter((h) => {
    if (filterEnv.value && h.environmentName !== filterEnv.value) return false;
    if (filterStatus.value && h.status !== filterStatus.value) return false;
    if (startTs !== null && h.startTime < startTs) return false;
    if (endTs !== null && h.startTime > endTs) return false;
    return true;
  });
});
</script>

<template>
  <div class="h-full p-6">
    <div class="mb-8">
      <h1 class="text-3xl font-bold bg-gradient-to-r from-primary to-blue-600 bg-clip-text text-transparent w-fit pb-1">历史记录</h1>
      <p class="text-muted-foreground mt-2 text-sm">查看过往的部署记录与执行日志</p>
    </div>

    <div class="rounded-lg border p-4 mb-4">
      <div class="grid grid-cols-4 gap-3 items-end">
        <div>
          <label class="block text-xs text-muted-foreground mb-1">环境</label>
          <select v-model="filterEnv" class="w-full rounded border border-input bg-background px-2 py-1 text-sm">
            <option value="">全部</option>
            <option v-for="env in envOptions" :key="env" :value="env">
              {{ env }}
            </option>
          </select>
        </div>
        <div>
          <label class="block text-xs text-muted-foreground mb-1">状态</label>
          <select
            v-model="filterStatus"
            class="w-full rounded border border-input bg-background px-2 py-1 text-sm"
          >
            <option value="">全部</option>
            <option value="success">成功</option>
            <option value="failed">失败</option>
            <option value="canceled">已取消</option>
          </select>
        </div>
        <div class="col-span-2">
          <label class="block text-xs text-muted-foreground mb-1"
            >时间区间</label
          >
          <DateRange v-model="dateRange" />
        </div>
      </div>
    </div>

    <div class="rounded-lg border mb-6">
      <table class="w-full">
        <thead class="bg-muted/50">
          <tr>
            <th class="px-4 py-3 text-left font-medium">时间</th>
            <th class="px-4 py-3 text-left font-medium">环境</th>
            <th class="px-4 py-3 text-left font-medium">状态</th>
            <th class="px-4 py-3 text-left font-medium">耗时</th>
            <th class="px-4 py-3 text-left font-medium">文件数</th>
            <th class="px-4 py-3 text-left font-medium">操作</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="history in filteredHistories" :key="history.id">
            <!-- 数据行 -->
            <tr class="border-t hover:bg-muted/30">
              <td class="px-4 py-3">{{ formatTime(history.startTime) }}</td>
              <td class="px-4 py-3">{{ history.environmentName }}</td>
              <td class="px-4 py-3">
                <span
                  class="inline-flex items-center rounded-full px-2 py-1 text-xs font-medium"
                  :class="{
                    'bg-green-100 text-green-800': history.status === 'success',
                    'bg-red-100 text-red-800': history.status === 'failed',
                    'bg-yellow-100 text-yellow-800':
                      history.status === 'canceled',
                  }"
                >
                  {{
                    history.status === "success"
                      ? "成功"
                      : history.status === "failed"
                      ? "失败"
                      : "已取消"
                  }}
                </span>
              </td>
              <td class="px-4 py-3">{{ formatDuration(history.duration) }}</td>
              <td class="px-4 py-3">{{ history.files.length }}</td>
              <td class="px-4 py-3">
                <button
                  class="rounded border px-2 py-1 text-xs hover:bg-accent"
                  @click="
                    selectedId = selectedId === history.id ? '' : history.id
                  "
                >
                  {{ selectedId === history.id ? "收起" : "查看" }}
                </button>
              </td>
            </tr>
            <!-- 内联展开详情行 -->
            <tr v-if="selectedId === history.id" class="bg-muted/20">
              <td colspan="6" class="px-6 py-4">
                <div class="flex items-center justify-between mb-2">
                  <div class="text-sm font-semibold">
                    详情：{{ history.environmentName }}
                  </div>
                  <div class="text-xs text-muted-foreground">
                    开始 {{ formatTime(history.startTime) }} ・ 结束
                    {{ formatTime(history.endTime) }}
                  </div>
                </div>
                <div class="mb-3" v-if="history.errorMessage">
                  <div class="text-xs text-muted-foreground mb-1">失败原因</div>
                  <div
                    class="rounded border p-2 text-xs text-red-600 bg-red-50"
                  >
                    {{ history.errorMessage }}
                  </div>
                </div>
                <div>
                  <div class="flex items-center justify-between mb-1">
                    <div class="text-xs text-muted-foreground">整体日志</div>
                    <div
                      class="text-xs text-muted-foreground"
                      v-if="historyStore.loadingLogs"
                    >
                      加载中...
                    </div>
                  </div>
                  <div
                    class="rounded border bg-gray-900 text-gray-100 p-3 h-60 overflow-auto text-xs font-mono"
                  >
                    <div
                      v-if="historyStore.logs.length === 0"
                      class="text-gray-400"
                    >
                      暂无日志
                    </div>
                    <div v-else>
                      <div
                        v-for="log in historyStore.logs"
                        :key="log.id"
                        class="mb-1"
                        :class="{
                          'text-blue-400': log.level === 'INFO',
                          'text-yellow-400': log.level === 'WARN',
                          'text-red-400': log.level === 'ERROR',
                          'text-gray-400': log.level === 'DEBUG',
                        }"
                      >
                        <span class="text-gray-500"
                          >[{{ formatTime(log.timestamp) }}]</span
                        >
                        <span class="font-bold ml-2">[{{ log.level }}]</span>
                        <span class="ml-2">{{ log.message }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </td>
            </tr>
          </template>
          <tr v-if="filteredHistories.length === 0">
            <td colspan="6" class="px-4 py-8 text-center text-muted-foreground">
              暂无历史记录
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
