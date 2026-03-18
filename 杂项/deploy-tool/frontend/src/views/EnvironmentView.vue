<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { useEnvironmentStore } from "../stores/environment";
import type { Environment } from "../types";

const envStore = useEnvironmentStore();
const activeTab = ref("basic");
const isChecking = ref(false);
const editingEnv = ref<Environment | null>(null);
const showDeleteModal = ref(false);
const deleteTargetName = ref("");
const checkResultCollapsed = ref(false);

onMounted(async () => {
  await envStore.fetchEnvironments();
  if (envStore.environments.length > 0) {
    selectEnvironment(envStore.environments[0]);
  }
});

function selectEnvironment(env: Environment) {
  const copy = JSON.parse(JSON.stringify(env));
  copy.servers = Array.isArray(copy.servers) ? copy.servers : [];
  copy.targetFiles = Array.isArray(copy.targetFiles) ? copy.targetFiles : [];
  if (!copy.buildType) {
    copy.buildType = "backend";
  }
  if (copy.buildType === "frontend") {
    if (copy.targetFiles.length === 0) {
      copy.targetFiles = [
        {
          id: `file_${Date.now()}`,
          localPath: "dist.zip",
          remoteName: "dist.zip",
          urlPath: "",
          defaultCheck: true,
        },
      ];
    } else if (!copy.targetFiles[0].urlPath) {
      copy.targetFiles[0].urlPath = "";
    }
  }
  editingEnv.value = copy;
  activeTab.value = "basic";
  envStore.checkResult = null;
}

async function addNewEnvironment() {
  const newEnv = await envStore.createNewEnvironment();
  envStore.environments.push(newEnv);
  selectEnvironment(newEnv);
}

function addServer() {
  if (!editingEnv.value) return;
  if (!Array.isArray(editingEnv.value.servers)) {
    editingEnv.value.servers = [];
  }
  editingEnv.value.servers.push(envStore.createNewServer());
}

function removeServer(index: number) {
  if (!editingEnv.value) return;
  editingEnv.value.servers.splice(index, 1);
}

function addTargetFile() {
  if (!editingEnv.value) return;
  if (editingEnv.value.buildType === "frontend") return;
  if (!Array.isArray(editingEnv.value.targetFiles)) {
    editingEnv.value.targetFiles = [];
  }
  editingEnv.value.targetFiles.push(envStore.createNewTargetFile());
}

function removeTargetFile(index: number) {
  if (!editingEnv.value) return;
  editingEnv.value.targetFiles.splice(index, 1);
}

async function saveEnvironment() {
  if (!editingEnv.value) return;
  try {
    if (!editingEnv.value.buildType) {
      editingEnv.value.buildType = "backend";
    }
    if (editingEnv.value.buildType === "frontend") {
      editingEnv.value.targetFiles = [
        {
          id: editingEnv.value.targetFiles?.[0]?.id || `file_${Date.now()}`,
          localPath: "dist.zip",
          remoteName: "dist.zip",
          urlPath: editingEnv.value.targetFiles?.[0]?.urlPath || "",
          defaultCheck: true,
        },
      ];
    }
    if (editingEnv.value.buildType !== "frontend") {
      const invalidServer = (editingEnv.value.servers || []).find(
        (server) =>
          server.enableRestart && isInvalidRestartScript(server.restartScript)
      );
      if (invalidServer) {
        window.alert(
          `服务器【${invalidServer.name || invalidServer.host || "未命名"}】的重启脚本必须以 .sh 结尾`
        );
        return;
      }
    }
    await envStore.saveEnvironment(editingEnv.value);
  } catch (error) {
    console.error("Save failed:", error);
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
    console.error("Delete failed:", error);
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

function shouldExpandRestartScriptPath(script: string): boolean {
  if (!script) return false;
  if (/[ \t\r\n;&|><`()]/.test(script)) return false;
  const lower = script.toLowerCase();
  return lower.endsWith(".sh") && (script.includes("/") || script.includes("\\"));
}

function shellQuote(value: string): string {
  if (!value) return "''";
  return `'${value.replace(/'/g, `'\\''`)}'`;
}

function buildRestartCommand(script: string, useSudo: boolean): string {
  const trimmed = (script || "").trim();
  if (!trimmed) return "";

  let cmd = trimmed;
  if (shouldExpandRestartScriptPath(trimmed)) {
    const normalized = trimmed.replace(/\\/g, "/");
    const slashIndex = normalized.lastIndexOf("/");
    if (slashIndex > 0 && slashIndex < normalized.length - 1) {
      const dir = normalized.slice(0, slashIndex);
      const file = normalized.slice(slashIndex + 1);
      cmd = `cd ${shellQuote(dir)} && sh ${shellQuote("./" + file)}`;
    }
  }

  if (useSudo) {
    return `sudo sh -c ${shellQuote(cmd)}`;
  }
  return cmd;
}

function isInvalidRestartScript(script: string): boolean {
  const trimmed = (script || "").trim();
  if (!trimmed) return false;
  return !trimmed.toLowerCase().endsWith(".sh");
}

function normalizeRemotePath(path: string): string {
  const trimmed = (path || "").trim();
  if (!trimmed) return "/";
  const withoutTail = trimmed.replace(/\/$/, "");
  return withoutTail.startsWith("/") ? withoutTail : `/${withoutTail}`;
}

function getFrontendRemoteName(): string {
  const name = editingEnv.value?.targetFiles?.[0]?.remoteName?.trim();
  return name || "dist.zip";
}

function wrapSudo(cmd: string, useSudo: boolean): string {
  if (!useSudo) return cmd;
  return `sudo sh -c ${shellQuote(cmd)}`;
}

function buildFrontendBackupCommand(baseDir: string, useSudo: boolean): string {
  const cmd = [
    "set -e",
    `cd ${shellQuote(baseDir)}`,
    "TS=$(date +%Y%m%d%H%M%S)",
    "if [ -d dist ]; then mv dist dist.${TS}.bak; fi",
  ].join("\n");
  return wrapSudo(cmd, useSudo);
}

function buildFrontendUnzipCommand(baseDir: string, remoteName: string, useSudo: boolean): string {
  const cmd = [
    "set -e",
    `cd ${shellQuote(baseDir)}`,
    "TMP_DIR=__dist_tmp__$(date +%Y%m%d%H%M%S)",
    'rm -rf "$TMP_DIR"',
    "rm -rf dist",
    `(unzip -o ${shellQuote(remoteName)} -d "$TMP_DIR" >/dev/null 2>&1 || python3 -m zipfile -e ${shellQuote(remoteName)} "$TMP_DIR")`,
    'mv "$TMP_DIR"/dist dist',
    'rm -rf "$TMP_DIR"',
    `rm -f ${shellQuote(remoteName)}`,
  ].join("\n");
  return wrapSudo(cmd, useSudo);
}

watch(
  () => editingEnv.value?.buildType,
  (buildType) => {
    if (buildType === "frontend" && editingEnv.value) {
      if (!Array.isArray(editingEnv.value.targetFiles)) {
        editingEnv.value.targetFiles = [];
      }
      if (editingEnv.value.targetFiles.length === 0) {
        editingEnv.value.targetFiles = [
          {
            id: `file_${Date.now()}`,
            localPath: "dist.zip",
            remoteName: "dist.zip",
            urlPath: "",
            defaultCheck: true,
          },
        ];
      }
    }
    if (buildType === "frontend" && activeTab.value === "targets") {
      activeTab.value = "servers";
    }
  }
);
</script>

<template>
  <div class="flex h-full">
    <aside class="w-72 border-r bg-card p-4">
      <div class="mb-6 flex items-center justify-between">
        <h2 class="text-xl font-bold bg-gradient-to-r from-primary to-blue-600 bg-clip-text text-transparent w-fit pb-1">环境列表</h2>
        <button
          class="rounded-md bg-primary px-3 py-1.5 text-sm text-primary-foreground hover:bg-primary/90 shadow-sm transition-all hover:shadow-md"
          @click="addNewEnvironment"
        >
          + 添加
        </button>
      </div>
      <div class="space-y-2">
        <div
          v-for="env in envStore.environments"
          :key="env.id"
          class="cursor-pointer rounded-md p-3 transition-all duration-200 transform"
          :class="[
            'border-l-4',
            editingEnv?.id === env.id
              ? env.buildType === 'frontend'
                ? 'bg-orange-200/60 border border-orange-300 border-l-orange-700 shadow-md translate-x-1'
                : 'bg-blue-200/60 border border-blue-300 border-l-blue-700 shadow-md translate-x-1'
              : env.buildType === 'frontend'
                ? 'bg-white border border-transparent border-l-orange-400 shadow-sm hover:shadow-md hover:translate-x-1'
                : 'bg-white border border-transparent border-l-blue-400 shadow-sm hover:shadow-md hover:translate-x-1'
          ]"
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
                  'bg-gray-300':
                    !env.checkStatus || env.checkStatus === 'unchecked',
                }"
              ></span>
              <div>
                <div class="font-medium">{{ env.name }}</div>
                <div class="text-sm text-muted-foreground">{{ env.identifier }}</div>
              </div>
            </div>
            <div class="text-right">
              <div class="flex items-center justify-end gap-2">
                <span
                  class="inline-flex w-12 justify-center rounded-full px-2 py-0.5 text-xs"
                  :class="{
                    'bg-orange-100 text-orange-700':
                      env.buildType === 'frontend',
                    'bg-blue-100 text-blue-700':
                      !env.buildType || env.buildType === 'backend',
                  }"
                >
                  {{ env.buildType === "frontend" ? "前端" : "后端" }}
                </span>
                <span
                  class="inline-flex w-12 justify-center rounded-full px-2 py-0.5 text-xs"
                  :class="{
                    'bg-green-100 text-green-700': env.identifier === 'dev',
                    'bg-yellow-100 text-yellow-700': env.identifier === 'test',
                    'bg-red-100 text-red-700': env.identifier === 'prod',
                    'bg-muted text-muted-foreground': ![
                      'dev',
                      'test',
                      'prod',
                    ].includes(env.identifier),
                  }"
                >
                  {{ env.identifier }}
                </span>
              </div>
              <div
                class="mt-1 text-xs"
                :class="{
                  'text-green-600': env.checkStatus === 'pass',
                  'text-yellow-600': env.checkStatus === 'warning',
                  'text-red-600': env.checkStatus === 'error',
                  'text-muted-foreground/70':
                    !env.checkStatus || env.checkStatus === 'unchecked',
                }"
              >
                {{
                  env.checkStatus === "pass"
                    ? "可用"
                    : env.checkStatus === "warning"
                    ? "通过"
                    : env.checkStatus === "error"
                    ? "未通过"
                    : "未自检"
                }}
              </div>
            </div>
          </div>
        </div>
      </div>
      <div
        v-if="envStore.environments.length === 0"
        class="py-8 text-center text-muted-foreground"
      >
        暂无环境，请添加
      </div>
    </aside>

    <main class="flex-1 overflow-y-auto p-6">
      <div
        v-if="!editingEnv"
        class="flex h-full items-center justify-center text-muted-foreground"
      >
        请选择或创建一个环境
      </div>

      <div v-else class="space-y-6">
        <div class="flex items-start justify-between">
          <div>
            <h2 class="text-3xl font-bold bg-gradient-to-r from-primary to-blue-600 bg-clip-text text-transparent w-fit pb-1">{{ editingEnv.name }}</h2>
            <p class="text-muted-foreground mt-2 text-sm">
              {{ editingEnv.description || "暂无描述" }}
            </p>
          </div>
          <div class="flex gap-3">
            <button
              class="rounded-md border px-4 py-2 text-sm font-medium transition-all hover:bg-accent hover:shadow-sm"
              @click="checkCurrentEnvironment"
              :disabled="isChecking"
            >
              {{ isChecking ? "检查中..." : "自检" }}
            </button>
            <button
              class="rounded-md border border-red-200 px-4 py-2 text-sm font-medium text-red-600 transition-all hover:bg-red-50 hover:border-red-300 hover:shadow-sm"
              @click="deleteCurrentEnvironment"
            >
              删除
            </button>
            <button
              class="rounded-md bg-primary px-5 py-2 text-sm font-medium text-white shadow-sm transition-all hover:bg-primary/90 hover:shadow-md hover:-translate-y-0.5"
              @click="saveEnvironment"
              :disabled="envStore.saving"
            >
              {{ envStore.saving ? "保存中..." : "保存" }}
            </button>
          </div>
        </div>

        <div
          v-if="envStore.checkResult"
          class="rounded-md border p-4"
          :class="
            envStore.checkResult.success
              ? 'border-green-200 bg-green-50'
              : 'border-red-200 bg-red-50'
          "
        >
          <div
            class="flex cursor-pointer items-center justify-between"
            @click="checkResultCollapsed = !checkResultCollapsed"
          >
            <h3 class="font-semibold">自检结果</h3>
            <div class="flex items-center gap-2">
              <span
                v-if="envStore.checkResult.success"
                class="rounded bg-green-100 px-2 py-1 text-sm font-medium text-green-700"
                >✓ 检查通过</span
              >
              <span
                v-else
                class="rounded bg-red-100 px-2 py-1 text-sm font-medium text-red-700"
                >✗ 检查失败</span
              >
              <svg
                class="h-4 w-4 text-muted-foreground transition-transform"
                :class="checkResultCollapsed ? '-rotate-90' : ''"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M19 9l-7 7-7-7"
                />
              </svg>
            </div>
          </div>
          <div v-show="!checkResultCollapsed" class="mt-3 space-y-2">
            <div
              v-for="(item, idx) in envStore.checkResult.checks"
              :key="idx"
              class="flex items-start gap-2 rounded bg-card p-2 text-sm"
            >
              <span v-if="item.status === 'pass'" class="mt-0.5 text-green-500"
                >✓</span
              >
              <span
                v-else-if="item.status === 'error'"
                class="mt-0.5 text-red-500"
                >✗</span
              >
              <span v-else class="mt-0.5 text-yellow-500">⚠</span>
              <div class="flex-1">
                <div class="font-medium">{{ item.name }}</div>
                <div
                  v-if="item.message"
                  :class="
                    item.status === 'error' ? 'text-red-600' : 'text-muted-foreground'
                  "
                >
                  {{ item.message }}
                </div>
              </div>
            </div>
          </div>
          <div
            v-if="!envStore.checkResult.success && !checkResultCollapsed"
            class="mt-3 rounded border border-red-300 bg-red-100 p-3"
          >
            <div class="font-medium text-red-800">
              请根据以上错误信息修改配置后重新自检
            </div>
          </div>
          <div class="mt-2 text-sm text-muted-foreground">
            {{ envStore.checkResult.summary }}
          </div>
        </div>

        <div class="border-b">
          <nav class="-mb-px flex space-x-4">
            <button
              class="whitespace-nowrap border-b-2 px-1 py-4 text-sm font-medium"
              :class="
                activeTab === 'basic'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              "
              @click="activeTab = 'basic'"
            >
              基本信息
            </button>
            <button
              class="whitespace-nowrap border-b-2 px-1 py-4 text-sm font-medium"
              :class="
                activeTab === 'servers'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              "
              @click="activeTab = 'servers'"
            >
              服务器
            </button>
            <button
              v-if="editingEnv?.buildType !== 'frontend'"
              class="whitespace-nowrap border-b-2 px-1 py-4 text-sm font-medium"
              :class="
                activeTab === 'targets'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              "
              @click="activeTab = 'targets'"
            >
              目标文件
            </button>
          </nav>
        </div>

        <div
          v-show="activeTab === 'basic'"
          class="space-y-4 rounded-md border p-4"
        >
          <h3 class="text-lg font-medium">基本信息</h3>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium">环境名称</label>
              <input
                v-model="editingEnv.name"
                class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                placeholder="例如: 开发环境"
              />
            </div>
            <div>
              <label class="block text-sm font-medium">环境标识</label>
              <input
                v-model="editingEnv.identifier"
                class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                placeholder="例如: dev, test, prod"
              />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium">描述</label>
            <textarea
              v-model="editingEnv.description"
              class="mt-1 block w-full rounded-md border border-input px-3 py-2"
              rows="2"
              placeholder="环境描述信息"
            ></textarea>
          </div>
          <div>
            <label class="block text-sm font-medium">项目根目录</label>
            <input
              v-model="editingEnv.projectRoot"
              class="mt-1 block w-full rounded-md border border-input px-3 py-2"
              placeholder="D:\\javaproject\\backcode"
            />
          </div>
          <div>
            <label class="block text-sm font-medium">构建类型</label>
            <select
              v-model="editingEnv.buildType"
              class="mt-1 block w-full rounded-md border border-input px-3 py-2"
            >
              <option value="backend">后端（Maven）</option>
              <option value="frontend">前端（npm build）</option>
            </select>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <!-- 云端部署 Toggle -->
            <div
              class="flex items-center justify-between rounded-md border border-border p-3"
            >
              <div>
                <div class="text-sm font-medium">云端部署</div>
                <div class="text-xs text-muted-foreground">
                  启用后将支持打包后上传服务器和远程重启
                </div>
              </div>
              <label class="relative inline-flex cursor-pointer items-center">
                <input
                  type="checkbox"
                  class="peer sr-only"
                  v-model="editingEnv.cloudDeploy"
                />
                <div
                  class="h-6 w-11 rounded-full bg-input peer-checked:bg-primary after:absolute after:left-0.5 after:top-0.5 after:h-5 after:w-5 after:rounded-full after:bg-white after:transition-all peer-checked:after:translate-x-5"
                ></div>
              </label>
            </div>

            <!-- 超时时间 Input -->
            <div
              class="flex flex-col justify-center rounded-md border border-border p-3"
            >
              <label class="block text-sm font-medium">超时时间 (秒)</label>
              <input
                v-model.number="editingEnv.timeout"
                type="number"
                class="mt-1 block w-full rounded-md border border-input px-3 py-1.5 text-sm"
                placeholder="600"
              />
            </div>
          </div>

          <!-- 部署前清理旧备份 配置已移除，改用系统全局配置 -->
        </div>

        <div v-show="activeTab === 'servers'" class="space-y-4">
          <div class="flex items-center justify-between rounded-md border p-4">
            <div>
              <h3 class="text-lg font-medium">服务器配置</h3>
              <p class="text-sm text-muted-foreground">SSH 远程服务器配置</p>
            </div>
            <button
              class="rounded-md bg-primary px-3 py-1.5 text-sm text-white hover:bg-primary/90"
              @click="addServer"
            >
              + 添加服务器
            </button>
          </div>

          <div
            v-if="(editingEnv?.servers?.length ?? 0) === 0"
            class="py-8 text-center text-muted-foreground"
          >
            暂无服务器配置，请添加
          </div>
          <div v-else class="space-y-4">
            <div
              v-for="(server, index) in editingEnv?.servers || []"
              :key="server.id"
              class="rounded-md border p-4"
            >
              <div class="mb-4 flex items-center justify-between">
                <h4 class="font-medium">服务器 {{ index + 1 }}</h4>
                <button
                  class="text-red-500 hover:text-red-700"
                  @click="removeServer(index)"
                >
                  删除
                </button>
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium">服务器名称</label>
                  <input
                    v-model="server.name"
                    class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                    placeholder="开发服务器"
                  />
                </div>
                <div>
                  <label class="block text-sm font-medium">主机地址</label>
                  <input
                    v-model="server.host"
                    class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                    placeholder="192.168.1.100"
                  />
                </div>
                <div>
                  <label class="block text-sm font-medium">SSH 端口</label>
                  <input
                    v-model.number="server.port"
                    type="number"
                    class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                    placeholder="22"
                  />
                </div>
                <div>
                  <label class="block text-sm font-medium">用户名</label>
                  <input
                    v-model="server.username"
                    class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                    placeholder="root"
                  />
                </div>
                <div>
                  <label class="block text-sm font-medium">密码</label>
                  <input
                    v-model="server.password"
                    type="password"
                    class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                    placeholder="请输入密码"
                  />
                </div>
                <div>
                  <label class="block text-sm font-medium">部署目录</label>
                  <input
                    v-model="server.deployDir"
                    class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                    placeholder="/home/omp/jar/"
                  />
                </div>
                <template v-if="editingEnv?.buildType !== 'frontend'">
                  <div
                    class="flex items-center justify-between rounded-md border border-border p-3"
                  >
                    <div>
                      <div class="text-sm font-medium">启用重启</div>
                      <div class="text-xs text-muted-foreground">
                        部署后自动执行重启脚本
                      </div>
                    </div>
                    <label
                      class="relative inline-flex cursor-pointer items-center"
                    >
                      <input
                        type="checkbox"
                        class="peer sr-only"
                        v-model="server.enableRestart"
                      />
                      <div
                        class="h-6 w-11 rounded-full bg-input peer-checked:bg-primary after:absolute after:left-0.5 after:top-0.5 after:h-5 after:w-5 after:rounded-full after:bg-white after:transition-all peer-checked:after:translate-x-5"
                      ></div>
                    </label>
                  </div>
                  <div
                    class="flex items-center justify-between rounded-md border border-border p-3"
                  >
                    <div>
                      <div class="text-sm font-medium">使用 Sudo</div>
                      <div class="text-xs text-muted-foreground">
                        使用 sudo 权限执行命令
                      </div>
                    </div>
                    <label
                      class="relative inline-flex cursor-pointer items-center"
                    >
                      <input
                        type="checkbox"
                        class="peer sr-only"
                        v-model="server.useSudo"
                      />
                      <div
                        class="h-6 w-11 rounded-full bg-input peer-checked:bg-primary after:absolute after:left-0.5 after:top-0.5 after:h-5 after:w-5 after:rounded-full after:bg-white after:transition-all peer-checked:after:translate-x-5"
                      ></div>
                    </label>
                  </div>
                  <div v-if="server.enableRestart" class="col-span-2">
                    <label class="block text-sm font-medium">重启脚本</label>
                    <input
                      v-model="server.restartScript"
                      class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                      :class="{ 'border-red-300 focus:border-red-400': isInvalidRestartScript(server.restartScript) }"
                      placeholder="/home/omp/jar/restart.sh"
                    />
                    <p
                      v-if="isInvalidRestartScript(server.restartScript)"
                      class="mt-1 text-sm text-red-600"
                    >
                      重启脚本必须以 .sh 结尾
                    </p>
                  </div>
                  <div
                    v-if="server.enableRestart"
                    class="col-span-2 rounded-md border border-amber-300 bg-amber-50 p-3"
                  >
                    <div class="text-xs text-amber-700 mb-1">将要执行的 Shell 命令</div>
                    <pre class="text-sm text-amber-900 whitespace-pre-wrap break-all font-mono">{{
                      buildRestartCommand(server.restartScript, server.useSudo) || "请先填写重启脚本"
                    }}</pre>
                    <p class="mt-2 text-xs text-amber-700">请确认 sudo 权限与脚本路径正确</p>
                  </div>
                </template>
            </div>
          </div>
          <div
            v-if="editingEnv?.buildType === 'frontend'"
            class="rounded-md border p-4"
          >
            <div class="mb-3">
              <h4 class="font-medium">前端文件信息</h4>
              <p class="text-sm text-muted-foreground">部署目录直接使用每台服务器的“部署目录”字段，不再单独配置 URL 路径</p>
            </div>
            <div class="space-y-3">
              <div class="grid grid-cols-10 items-center gap-3">
                <label class="col-span-3 text-sm font-medium">本地路径</label>
                <input
                  value="dist.zip"
                  disabled
                  class="col-span-7 block w-full rounded-md border border-input px-3 py-2"
                />
              </div>
              <div class="col-span-2 rounded-md border border-blue-200 bg-blue-50 p-3">
                <div class="text-xs text-blue-700 mb-1">部署命令预览（上传 dist.zip 后）</div>
                <div
                  v-if="(editingEnv?.servers?.length ?? 0) === 0"
                  class="text-sm text-blue-900"
                >
                  请先添加服务器并填写部署目录
                </div>
                <div v-else class="space-y-3">
                  <div
                    v-for="(server, idx) in editingEnv?.servers || []"
                    :key="`frontend-cmd-${server.id}-${idx}`"
                    class="rounded border border-blue-200 bg-white p-3"
                  >
                    <div class="text-xs text-muted-foreground mb-1">{{ server.name || `服务器 ${idx + 1}` }} (目录: {{ normalizeRemotePath(server.deployDir) }})</div>
                    <div class="text-xs text-blue-700">备份旧 dist</div>
                    <pre class="text-sm text-blue-900 whitespace-pre-wrap break-all font-mono">{{ buildFrontendBackupCommand(normalizeRemotePath(server.deployDir), server.useSudo) }}</pre>
                    <div class="text-xs text-blue-700 mt-2">解压并覆盖 dist</div>
                    <pre class="text-sm text-blue-900 whitespace-pre-wrap break-all font-mono">{{ buildFrontendUnzipCommand(normalizeRemotePath(server.deployDir), getFrontendRemoteName(), server.useSudo) }}</pre>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        </div>

        <div
          v-show="activeTab === 'targets' && editingEnv?.buildType !== 'frontend'"
          class="space-y-4"
        >
          <div class="flex items-center justify-between rounded-md border p-4">
            <div>
              <h3 class="text-lg font-medium">目标文件配置</h3>
              <p class="text-sm text-muted-foreground">
                {{
                  editingEnv?.buildType === "frontend"
                    ? "前端部署仅上传 dist.zip"
                    : "需要部署的 Jar 包列表"
                }}
              </p>
            </div>
            <button
              v-if="editingEnv?.buildType !== 'frontend'"
              class="rounded-md bg-primary px-3 py-1.5 text-sm text-white hover:bg-primary/90"
              @click="addTargetFile"
            >
              + 添加文件
            </button>
          </div>

          <div
            v-if="editingEnv?.buildType === 'frontend'"
            class="rounded-md border p-4"
          >
            <div class="mb-4 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <input type="checkbox" checked disabled class="rounded" />
                <h4 class="font-medium">文件 1</h4>
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium">本地路径</label>
                <input
                  value="dist.zip"
                  disabled
                  class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                />
              </div>
              <div>
                  <label class="block text-sm font-medium">
                    远程文件名
                    <span class="block text-xs text-muted-foreground">(可选)</span>
                  </label>
                <input
                  value="dist.zip"
                  disabled
                  class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                />
              </div>
            </div>
          </div>

          <template v-else>
            <div
              v-if="(editingEnv?.targetFiles?.length ?? 0) === 0"
              class="py-8 text-center text-muted-foreground"
            >
              暂无目标文件，请添加
            </div>
            <div v-else class="space-y-4">
              <div
                v-for="(file, index) in editingEnv?.targetFiles || []"
                :key="file.id"
                class="rounded-md border p-4"
              >
                <div class="mb-4 flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <label class="relative inline-flex cursor-pointer items-center">
                      <input
                        type="checkbox"
                        class="peer sr-only"
                        v-model="file.defaultCheck"
                      />
                      <div
                        class="h-6 w-11 rounded-full bg-input peer-checked:bg-primary after:absolute after:left-0.5 after:top-0.5 after:h-5 after:w-5 after:rounded-full after:bg-white after:transition-all peer-checked:after:translate-x-5"
                      ></div>
                    </label>
                    <h4 class="font-medium">文件 {{ index + 1 }}</h4>
                  </div>
                  <button
                    class="text-red-500 hover:text-red-700"
                    @click="removeTargetFile(index)"
                  >
                    删除
                  </button>
                </div>
                <div class="space-y-3">
                  <div class="grid grid-cols-10 items-center gap-3">
                    <label class="col-span-3 text-sm font-medium">本地路径</label>
                    <input
                      v-model="file.localPath"
                      class="col-span-7 block w-full rounded-md border border-input px-3 py-2"
                      placeholder="startup\xxx\target\xxx.jar"
                    />
                  </div>
                  <div class="grid grid-cols-10 items-center gap-3">
                    <label class="col-span-3 text-sm font-medium">
                      远程文件名（可选）
                    </label>
                    <input
                      v-model="file.remoteName"
                      class="col-span-7 block w-full rounded-md border border-input px-3 py-2"
                      placeholder="留空则使用原文件名"
                    />
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
    </main>

    <div
      v-if="showDeleteModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
    >
      <div class="w-full max-w-md rounded-lg bg-card p-6 shadow-xl">
        <div class="mb-4 flex items-center gap-3">
          <div
            class="flex h-10 w-10 items-center justify-center rounded-full bg-red-100"
          >
            <svg
              class="h-6 w-6 text-red-600"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
          </div>
          <h3 class="text-lg font-semibold text-foreground">确认删除</h3>
        </div>
        <p class="mb-6 text-muted-foreground">
          确定要删除环境
          <span class="font-medium text-foreground"
            >"{{ deleteTargetName }}"</span
          >
          吗？<br />
          <span class="text-sm text-red-500">此操作不可撤销</span>
        </p>
        <div class="flex justify-end gap-3">
          <button
            @click="cancelDelete"
            class="rounded-md border border-input bg-white px-4 py-2 text-sm font-medium text-foreground hover:bg-muted/50"
          >
            取消
          </button>
          <button
            @click="confirmDelete"
            class="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
          >
            确认删除
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
