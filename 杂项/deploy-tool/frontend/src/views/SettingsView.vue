<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useSettingsStore } from '../stores/settings';

const settingsStore = useSettingsStore();
const activeTab = ref('general');

onMounted(async () => {
  await settingsStore.fetchAll();
});

async function saveSettings() {
  try {
    await settingsStore.saveAll();
  } catch (error) {
    console.error('Save failed:', error);
  }
}

function resetToDefaults() {
  settingsStore.globalSettings = {
    defaultTimeout: 600,
    logRetentionDays: 30,
    backupEnabled: true,
    notifyOnComplete: true,
    cloudDeploy: true,
    theme: 'system',
    language: 'zh-Hans',
  };
  settingsStore.systemDefaults = {
    jdkPath: '',
    mavenPath: '',
    mavenSettingsPath: '',
    mavenRepoPath: '',
    mavenArgs: 'clean package -DskipTests',
  };
}
</script>

<template>
  <div class="h-full overflow-y-auto p-6">
    <div class="mx-auto max-w-4xl space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold">系统设置</h1>
          <p class="text-gray-500">配置应用程序的全局设置和默认值</p>
        </div>
        <div class="flex gap-2">
          <button 
            class="rounded-md border px-3 py-2 text-sm hover:bg-gray-50"
            @click="resetToDefaults"
          >
            重置
          </button>
          <button 
            class="rounded-md bg-primary px-4 py-2 text-sm text-white hover:bg-primary/90"
            @click="saveSettings" 
            :disabled="settingsStore.saving"
          >
            {{ settingsStore.saving ? '保存中...' : '保存设置' }}
          </button>
        </div>
      </div>

      <div class="border-b">
        <nav class="-mb-px flex space-x-4">
          <button
            class="whitespace-nowrap border-b-2 px-1 py-4 text-sm font-medium"
            :class="activeTab === 'general' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'"
            @click="activeTab = 'general'"
          >
            通用设置
          </button>
          <button
            class="whitespace-nowrap border-b-2 px-1 py-4 text-sm font-medium"
            :class="activeTab === 'defaults' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'"
            @click="activeTab = 'defaults'"
          >
            默认配置
          </button>
        </nav>
      </div>

      <div v-show="activeTab === 'general'" class="space-y-4">
        <div class="rounded-md border p-4">
          <h3 class="text-lg font-medium">部署设置</h3>
          <p class="mb-4 text-sm text-gray-500">配置部署相关的全局参数</p>
          <div class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium">部署超时时间 (秒)</label>
                <input 
                  v-model.number="settingsStore.globalSettings.defaultTimeout" 
                  type="number"
                  class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
                  placeholder="600"
                />
              </div>
              <div>
                <label class="block text-sm font-medium">日志保留天数</label>
                <input 
                  v-model.number="settingsStore.globalSettings.logRetentionDays" 
                  type="number"
                  class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
                  placeholder="30"
                />
              </div>
            </div>
            <label class="flex items-center gap-2">
              <input type="checkbox" v-model="settingsStore.globalSettings.backupEnabled" class="rounded" />
              <span class="text-sm">部署前自动备份</span>
            </label>
            <label class="flex items-center gap-2">
              <input type="checkbox" v-model="settingsStore.globalSettings.cloudDeploy" class="rounded" />
              <span class="text-sm">云端部署（默认）</span>
            </label>
            <label class="flex items-center gap-2">
              <input type="checkbox" v-model="settingsStore.globalSettings.notifyOnComplete" class="rounded" />
              <span class="text-sm">部署完成后发送通知</span>
            </label>
          </div>
        </div>

        <div class="rounded-md border p-4">
          <h3 class="text-lg font-medium">外观设置</h3>
          <p class="mb-4 text-sm text-gray-500">配置应用程序的外观和语言</p>
          <div class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium">主题</label>
                <select 
                  v-model="settingsStore.globalSettings.theme" 
                  class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
                >
                  <option value="light">浅色</option>
                  <option value="dark">深色</option>
                  <option value="system">跟随系统</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium">语言</label>
                <select 
                  v-model="settingsStore.globalSettings.language" 
                  class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
                >
                  <option value="zh-Hans">简体中文</option>
                  <option value="zh-Hant">繁體中文</option>
                  <option value="en">English</option>
                </select>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-show="activeTab === 'defaults'" class="space-y-4">
        <div class="rounded-md border p-4">
          <h3 class="text-lg font-medium">系统默认配置</h3>
          <p class="mb-4 text-sm text-gray-500">新环境创建时的默认配置值</p>
          <div class="rounded-md bg-yellow-50 p-4 text-sm text-yellow-800 mb-4">
            <strong>提示：</strong>这些配置将作为新创建环境的默认值，但可以单独为每个环境覆盖。
          </div>
          <div class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium">默认 JDK 路径</label>
                <input 
                  v-model="settingsStore.systemDefaults.jdkPath" 
                  class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
                  placeholder="C:\Program Files\Java\jdk1.8.0_202\bin"
                />
              </div>
              <div>
                <label class="block text-sm font-medium">默认 Maven 路径</label>
                <input 
                  v-model="settingsStore.systemDefaults.mavenPath" 
                  class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
                  placeholder="D:\maven\bin\mvn.cmd"
                />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium">默认 Maven settings.xml</label>
              <input 
                v-model="settingsStore.systemDefaults.mavenSettingsPath" 
                class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
                placeholder="D:\maven\conf\settings.xml"
              />
            </div>
            <div>
              <label class="block text-sm font-medium">默认本地仓库路径</label>
              <input 
                v-model="settingsStore.systemDefaults.mavenRepoPath" 
                class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
                placeholder="D:\m2\repository"
              />
            </div>
            <div>
              <label class="block text-sm font-medium">默认 Maven 参数</label>
              <input 
                v-model="settingsStore.systemDefaults.mavenArgs" 
                class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
                placeholder="clean package -DskipTests"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
