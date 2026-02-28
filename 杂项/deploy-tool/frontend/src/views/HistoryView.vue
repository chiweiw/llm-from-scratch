<script setup lang="ts">
import { onMounted } from 'vue';
import { useHistoryStore } from '@/stores/history';

const historyStore = useHistoryStore();

onMounted(async () => {
  await historyStore.fetchHistories();
});

function formatTime(timestamp: number): string {
  return new Date(timestamp * 1000).toLocaleString('zh-CN');
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}秒`;
  const minutes = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${minutes}分${secs}秒`;
}
</script>

<template>
  <div class="h-full p-6">
    <h1 class="text-2xl font-bold mb-6">历史记录</h1>
    
    <div class="rounded-lg border">
      <table class="w-full">
        <thead class="bg-muted/50">
          <tr>
            <th class="px-4 py-3 text-left font-medium">时间</th>
            <th class="px-4 py-3 text-left font-medium">环境</th>
            <th class="px-4 py-3 text-left font-medium">状态</th>
            <th class="px-4 py-3 text-left font-medium">耗时</th>
            <th class="px-4 py-3 text-left font-medium">文件数</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="history in historyStore.histories" :key="history.id" class="border-t hover:bg-muted/30">
            <td class="px-4 py-3">{{ formatTime(history.startTime) }}</td>
            <td class="px-4 py-3">{{ history.environmentName }}</td>
            <td class="px-4 py-3">
              <span 
                class="inline-flex items-center rounded-full px-2 py-1 text-xs font-medium"
                :class="{
                  'bg-green-100 text-green-800': history.status === 'success',
                  'bg-red-100 text-red-800': history.status === 'failed',
                  'bg-yellow-100 text-yellow-800': history.status === 'canceled',
                }"
              >
                {{ history.status === 'success' ? '成功' : history.status === 'failed' ? '失败' : '已取消' }}
              </span>
            </td>
            <td class="px-4 py-3">{{ formatDuration(history.duration) }}</td>
            <td class="px-4 py-3">{{ history.files.length }}</td>
          </tr>
          <tr v-if="historyStore.histories.length === 0">
            <td colspan="5" class="px-4 py-8 text-center text-muted-foreground">
              暂无历史记录
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
