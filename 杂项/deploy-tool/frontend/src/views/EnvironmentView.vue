<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useEnvironmentStore } from '../stores/environment';
import type { Environment, ServerConfig, TargetFile } from '../types';

const envStore = useEnvironmentStore();
const activeTab = ref('basic');
const isChecking = ref(false);
const editingEnv = ref<Environment | null>(null);
const showDeleteModal = ref(false);
const deleteTargetName = ref('');
const checkResultCollapsed = ref(false);

onMounted(async () => {
  await envStore.fetchEnvironments();
  if (envStore.environments.length > 0) {
    selectEnvironment(envStore.environments[0]);
  }
});

function selectEnvironment(env: Environment) {
  editingEnv.value = JSON.parse(JSON.stringify(env));
  activeTab.value = 'basic';
  envStore.checkResult = null;
}

function addNewEnvironment() {
  const newEnv = envStore.createNewEnvironment();
  envStore.environments.push(newEnv);
  selectEnvironment(newEnv);
}

function addServer() {
  if (!editingEnv.value) return;
  editingEnv.value.servers.push(envStore.createNewServer());
}

function removeServer(index: number) {
  if (!editingEnv.value) return;
  editingEnv.value.servers.splice(index, 1);
}

function addTargetFile() {
  if (!editingEnv.value) return;
  editingEnv.value.targetFiles.push(envStore.createNewTargetFile());
}

function removeTargetFile(index: number) {
  if (!editingEnv.value) return;
  editingEnv.value.targetFiles.splice(index, 1);
}

async function saveEnvironment() {
  if (!editingEnv.value) return;
  try {
    await envStore.saveEnvironment(editingEnv.value);
  } catch (error) {
    console.error('Save failed:', error);
  }
}

async function deleteCurrentEnvironment() {
  if (!editingEnv.value) return;
  deleteTargetName.value = editingEnv.value.name;
  showDeleteModal.value = true;
}

async function confirmDelete() {
  try {
    await envStore.deleteEnvironment(editingEnv.value!.id);
    editingEnv.value = null;
    if (envStore.environments.length > 0) {
      selectEnvironment(envStore.environments[0]);
    }
  } catch (error) {
    console.error('Delete failed:', error);
  } finally {
    showDeleteModal.value = false;
  }
}

function cancelDelete() {
  showDeleteModal.value = false;
}

async function checkCurrentEnvironment() {
  if (!editingEnv.value) return;
  isChecking.value = true;
  try {
    await envStore.checkEnvironment(editingEnv.value.id);
  } finally {
    isChecking.value = false;
  }
}
</script>

<template>
  <div class="flex h-full">
    <aside class="w-72 border-r bg-gray-50 p-4">
      <div class="mb-4 flex items-center justify-between">
        <h2 class="text-lg font-semibold">环境列表</h2>
        <button 
          class="rounded-md bg-primary px-3 py-1.5 text-sm text-primary-foreground hover:bg-primary/90"
          @click="addNewEnvironment"
        >
          + 添加
        </button>
      </div>
      <div class="space-y-2">
        <div
          v-for="env in envStore.environments"
          :key="env.id"
          class="cursor-pointer rounded-md p-3 transition-colors hover:bg-gray-100"
          :class="{ 'bg-blue-100': editingEnv?.id === env.id }"
          @click="selectEnvironment(env)"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span
                class="h-2 w-2 rounded-full"
                :class="{
                  'bg-green-500': env.checkStatus === 'pass',
                  'bg-yellow-500': env.checkStatus === 'warning',
                  'bg-red-500': env.checkStatus === 'error',
                  'bg-gray-300': !env.checkStatus || env.checkStatus === 'unchecked',
                }"
              ></span>
              <div>
                <div class="font-medium">{{ env.name }}</div>
                <div class="text-sm text-gray-500">{{ env.identifier }}</div>
              </div>
            </div>
            <div class="text-right">
              <span
                class="rounded-full px-2 py-0.5 text-xs"
                :class="{
                  'bg-green-100 text-green-700': env.identifier === 'dev',
                  'bg-yellow-100 text-yellow-700': env.identifier === 'test',
                  'bg-red-100 text-red-700': env.identifier === 'prod',
                  'bg-gray-100 text-gray-700': !['dev', 'test', 'prod'].includes(env.identifier),
                }"
              >
                {{ env.identifier }}
              </span>
              <div class="mt-1 text-xs" :class="{
                'text-green-600': env.checkStatus === 'pass',
                'text-yellow-600': env.checkStatus === 'warning',
                'text-red-600': env.checkStatus === 'error',
                'text-gray-400': !env.checkStatus || env.checkStatus === 'unchecked',
              }">
                {{ env.checkStatus === 'pass' ? '可用' : env.checkStatus === 'warning' ? '有警告' : env.checkStatus === 'error' ? '未通过' : '未自检' }}
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-if="envStore.environments.length === 0" class="py-8 text-center text-gray-500">
        暂无环境，请添加
      </div>
    </aside>

    <main class="flex-1 overflow-y-auto p-6">
      <div v-if="!editingEnv" class="flex h-full items-center justify-center text-gray-500">
        请选择或创建一个环境
      </div>

      <div v-else class="space-y-4">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-2xl font-bold">{{ editingEnv.name }}</h2>
            <p class="text-gray-500">{{ editingEnv.description || '暂无描述' }}</p>
          </div>
          <div class="flex gap-2">
            <button 
              class="rounded-md border px-3 py-2 text-sm hover:bg-gray-50"
              @click="checkCurrentEnvironment" 
              :disabled="isChecking"
            >
              {{ isChecking ? '检查中...' : '自检' }}
            </button>
            <button 
              class="rounded-md border border-red-300 px-3 py-2 text-sm text-red-600 hover:bg-red-50"
              @click="deleteCurrentEnvironment"
            >
              删除
            </button>
            <button 
              class="rounded-md bg-primary px-4 py-2 text-sm text-white hover:bg-primary/90"
              @click="saveEnvironment" 
              :disabled="envStore.saving"
            >
              {{ envStore.saving ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>

        <div v-if="envStore.checkResult" class="rounded-md border p-4" :class="envStore.checkResult.success ? 'border-green-200 bg-green-50' : 'border-red-200 bg-red-50'">
          <div class="flex cursor-pointer items-center justify-between" @click="checkResultCollapsed = !checkResultCollapsed">
            <h3 class="font-semibold">自检结果</h3>
            <div class="flex items-center gap-2">
              <span v-if="envStore.checkResult.success" class="rounded bg-green-100 px-2 py-1 text-sm font-medium text-green-700">✓ 检查通过</span>
              <span v-else class="rounded bg-red-100 px-2 py-1 text-sm font-medium text-red-700">✗ 检查失败</span>
              <svg class="h-4 w-4 text-gray-500 transition-transform" :class="checkResultCollapsed ? '-rotate-90' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </div>
          </div>
          <div v-show="!checkResultCollapsed" class="mt-3 space-y-2">
            <div v-for="(item, idx) in envStore.checkResult.checks" :key="idx" class="flex items-start gap-2 rounded bg-white p-2 text-sm">
              <span v-if="item.status === 'pass'" class="mt-0.5 text-green-500">✓</span>
              <span v-else-if="item.status === 'error'" class="mt-0.5 text-red-500">✗</span>
              <span v-else class="mt-0.5 text-yellow-500">⚠</span>
              <div class="flex-1">
                <div class="font-medium">{{ item.name }}</div>
                <div v-if="item.message" :class="item.status === 'error' ? 'text-red-600' : 'text-gray-600'">{{ item.message }}</div>
              </div>
            </div>
          </div>
          <div v-if="!envStore.checkResult.success && !checkResultCollapsed" class="mt-3 rounded border border-red-300 bg-red-100 p-3">
            <div class="font-medium text-red-800">请根据以上错误信息修改配置后重新自检</div>
          </div>
          <div class="mt-2 text-sm text-gray-600">{{ envStore.checkResult.summary }}</div>
        </div>

        <div class="border-b">
          <nav class="-mb-px flex space-x-4">
            <button
              class="whitespace-nowrap border-b-2 px-1 py-4 text-sm font-medium"
              :class="activeTab === 'basic' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'"
              @click="activeTab = 'basic'"
            >
              基本信息
            </button>
            <button
              class="whitespace-nowrap border-b-2 px-1 py-4 text-sm font-medium"
              :class="activeTab === 'local' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'"
              @click="activeTab = 'local'"
            >
              本地配置
            </button>
            <button
              class="whitespace-nowrap border-b-2 px-1 py-4 text-sm font-medium"
              :class="activeTab === 'servers' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'"
              @click="activeTab = 'servers'"
            >
              服务器
            </button>
            <button
              class="whitespace-nowrap border-b-2 px-1 py-4 text-sm font-medium"
              :class="activeTab === 'targets' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'"
              @click="activeTab = 'targets'"
            >
              目标文件
            </button>
          </nav>
        </div>

        <div v-show="activeTab === 'basic'" class="space-y-4 rounded-md border p-4">
          <h3 class="text-lg font-medium">基本信息</h3>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium">环境名称</label>
              <input 
                v-model="editingEnv.name" 
                class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
                placeholder="例如: 开发环境"
              />
            </div>
            <div>
              <label class="block text-sm font-medium">环境标识</label>
              <input 
                v-model="editingEnv.identifier" 
                class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
                placeholder="例如: dev, test, prod"
              />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium">描述</label>
            <textarea 
              v-model="editingEnv.description" 
              class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
              rows="2"
              placeholder="环境描述信息"
            ></textarea>
          </div>
          <div class="flex items-center gap-4">
            <label class="flex items-center gap-2">
              <input type="checkbox" v-model="editingEnv.cloudDeploy" class="rounded" />
              <span class="text-sm font-medium">云端部署</span>
            </label>
            <span class="text-xs text-gray-500">（启用后将支持打包后上传服务器和远程重启）</span>
          </div>
          <div>
            <label class="block text-sm font-medium">超时时间 (秒)</label>
            <input 
              v-model.number="editingEnv.timeout" 
              type="number"
              class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
              placeholder="600"
            />
          </div>
          <div class="flex flex-wrap gap-4">
            <label class="flex items-center gap-2">
              <input type="checkbox" v-model="editingEnv.dryRun" class="rounded" />
              <span class="text-sm">干跑模式 (仅自检)</span>
            </label>
            <label class="flex items-center gap-2">
              <input type="checkbox" v-model="editingEnv.backupCleanup" class="rounded" />
              <span class="text-sm">部署前清理旧备份</span>
            </label>
          </div>
        </div>

        <div v-show="activeTab === 'local'" class="space-y-4 rounded-md border p-4">
          <h3 class="text-lg font-medium">本地环境配置</h3>
          <div>
            <label class="block text-sm font-medium">项目根目录</label>
            <input 
              v-model="editingEnv.local.projectRoot" 
              class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
              placeholder="D:\javaproject\backcode"
            />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium">JDK 路径</label>
              <input 
                v-model="editingEnv.local.jdkPath" 
                class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
                placeholder="C:\Program Files\Java\jdk1.8.0_202\bin"
              />
            </div>
            <div>
              <label class="block text-sm font-medium">Maven 可执行文件</label>
              <input 
                v-model="editingEnv.local.mavenPath" 
                class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
                placeholder="D:\maven\bin\mvn.cmd"
              />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium">Maven settings.xml</label>
            <input 
              v-model="editingEnv.local.mavenSettingsPath" 
              class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
              placeholder="D:\maven\conf\settings.xml"
            />
          </div>
          <div>
            <label class="block text-sm font-medium">本地仓库路径</label>
            <input 
              v-model="editingEnv.local.mavenRepoPath" 
              class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
              placeholder="D:\m2\repository"
            />
          </div>
          <div>
            <label class="block text-sm font-medium">Maven 参数</label>
            <input 
              v-model="editingEnv.local.mavenArgs" 
              class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2"
              placeholder="clean package -DskipTests"
            />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <label class="flex items-center gap-2">
              <input type="checkbox" v-model="editingEnv.local.mavenQuiet" class="rounded" />
              <span class="text-sm">安静模式 (-q)</span>
            </label>
            <label class="flex items-center gap-2">
              <input type="checkbox" v-model="editingEnv.local.compactMvnLog" class="rounded" />
              <span class="text-sm">精简日志输出</span>
            </label>
            <label class="flex items-center gap-2">
              <input type="checkbox" v-model="editingEnv.local.specifyPom" class="rounded" />
              <span class="text-sm">显式指定 pom.xml</span>
            </label>
            <label class="flex items-center gap-2">
              <input type="checkbox" v-model="editingEnv.local.offlineBuild" class="rounded" />
              <span class="text-sm">离线构建 (-o)</span>
            </label>
          </div>
        </div>

        <div v-show="activeTab === 'servers'" class="space-y-4">
          <div class="flex items-center justify-between rounded-md border p-4">
            <div>
              <h3 class="text-lg font-medium">服务器配置</h3>
              <p class="text-sm text-gray-500">SSH 远程服务器配置</p>
            </div>
            <button class="rounded-md bg-primary px-3 py-1.5 text-sm text-white hover:bg-primary/90" @click="addServer">
              + 添加服务器
            </button>
          </div>
          
          <div v-if="editingEnv.servers.length === 0" class="py-8 text-center text-gray-500">
            暂无服务器配置，请添加
          </div>
          <div v-else class="space-y-4">
            <div v-for="(server, index) in editingEnv.servers" :key="server.id" class="rounded-md border p-4">
              <div class="mb-4 flex items-center justify-between">
                <h4 class="font-medium">服务器 {{ index + 1 }}</h4>
                <button class="text-red-500 hover:text-red-700" @click="removeServer(index)">
                  删除
                </button>
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium">服务器名称</label>
                  <input v-model="server.name" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2" placeholder="开发服务器" />
                </div>
                <div>
                  <label class="block text-sm font-medium">主机地址</label>
                  <input v-model="server.host" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2" placeholder="192.168.1.100" />
                </div>
                <div>
                  <label class="block text-sm font-medium">SSH 端口</label>
                  <input v-model.number="server.port" type="number" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2" placeholder="22" />
                </div>
                <div>
                  <label class="block text-sm font-medium">用户名</label>
                  <input v-model="server.username" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2" placeholder="root" />
                </div>
                <div>
                  <label class="block text-sm font-medium">密码</label>
                  <input v-model="server.password" type="password" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2" placeholder="请输入密码" />
                </div>
                <div>
                  <label class="block text-sm font-medium">部署目录</label>
                  <input v-model="server.deployDir" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2" placeholder="/home/omp/jar/" />
                </div>
                <div class="col-span-2">
                  <label class="block text-sm font-medium">重启脚本</label>
                  <input v-model="server.restartScript" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2" placeholder="/home/omp/jar/restart.sh" />
                </div>
                <label class="flex items-center gap-2">
                  <input type="checkbox" v-model="server.enableRestart" class="rounded" />
                  <span class="text-sm">启用重启</span>
                </label>
                <label class="flex items-center gap-2">
                  <input type="checkbox" v-model="server.useSudo" class="rounded" />
                  <span class="text-sm">使用 Sudo</span>
                </label>
              </div>
            </div>
          </div>
        </div>

        <div v-show="activeTab === 'targets'" class="space-y-4">
          <div class="flex items-center justify-between rounded-md border p-4">
            <div>
              <h3 class="text-lg font-medium">目标文件配置</h3>
              <p class="text-sm text-gray-500">需要部署的 Jar 包列表</p>
            </div>
            <button class="rounded-md bg-primary px-3 py-1.5 text-sm text-white hover:bg-primary/90" @click="addTargetFile">
              + 添加文件
            </button>
          </div>
          
          <div v-if="editingEnv.targetFiles.length === 0" class="py-8 text-center text-gray-500">
            暂无目标文件，请添加
          </div>
          <div v-else class="space-y-4">
            <div v-for="(file, index) in editingEnv.targetFiles" :key="file.id" class="rounded-md border p-4">
              <div class="mb-4 flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <input type="checkbox" v-model="file.defaultCheck" class="rounded" />
                  <h4 class="font-medium">文件 {{ index + 1 }}</h4>
                </div>
                <button class="text-red-500 hover:text-red-700" @click="removeTargetFile(index)">
                  删除
                </button>
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium">本地路径</label>
                  <input v-model="file.localPath" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2" placeholder="startup\xxx\target\xxx.jar" />
                </div>
                <div>
                  <label class="block text-sm font-medium">远程文件名 (可选)</label>
                  <input v-model="file.remoteName" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2" placeholder="留空则使用原文件名" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>

    <div v-if="showDeleteModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div class="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
        <div class="mb-4 flex items-center gap-3">
          <div class="flex h-10 w-10 items-center justify-center rounded-full bg-red-100">
            <svg class="h-6 w-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <h3 class="text-lg font-semibold text-gray-900">确认删除</h3>
        </div>
        <p class="mb-6 text-gray-600">
          确定要删除环境 <span class="font-medium text-gray-900">"{{ deleteTargetName }}"</span> 吗？<br>
          <span class="text-sm text-red-500">此操作不可撤销</span>
        </p>
        <div class="flex justify-end gap-3">
          <button @click="cancelDelete" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50">
            取消
          </button>
          <button @click="confirmDelete" class="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700">
            确认删除
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
