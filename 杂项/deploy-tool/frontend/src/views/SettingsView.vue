<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useSettingsStore } from "../stores/settings";
import { useWailsEvent } from "@/lib/useWailsEvent";

const settingsStore = useSettingsStore();
const activeTab = ref("general");
const mavenCommandInput = ref("");
const parsing = ref(false);
const detectedJdks = ref<{ path: string; source: string }[]>([]);
const detectingJdk = ref(false);
const errorMsg = ref("");
const successMsg = ref("");

onMounted(async () => {
  await settingsStore.fetchAll();
  if (!settingsStore.systemDefaults.jdkPath) {
    await detectJdk();
  }
});

useWailsEvent(
  "jdk-detection-result",
  (jdks: { path: string; source: string }[] | null) => {
    const safeJdks = Array.isArray(jdks) ? jdks : [];
    console.log("Received JDK detection result:", safeJdks);
    detectingJdk.value = false;
    detectedJdks.value = safeJdks;
    if (safeJdks.length > 0) {
      settingsStore.systemDefaults.jdkPath = safeJdks[0].path;
      successMsg.value = "检测到 " + safeJdks.length + " 个 JDK";
    } else {
      errorMsg.value = "未检测到 JDK，请手动配置";
    }
  }
);

async function saveSettings() {
  try {
    await settingsStore.saveAll();
  } catch (error) {
    console.error("Save failed:", error);
  }
}

function resetToDefaults() {
  settingsStore.globalSettings = {
    defaultTimeout: 0,
    logRetentionDays: 0,
    backupEnabled: false,
    backupCleanup: true,
    notifyOnComplete: false,
    cloudDeploy: false,
    offlineBuild: true,
    lightLog: true,
    theme: "",
    language: "",
  };
  settingsStore.systemDefaults = {
    jdkPath: "",
    mavenPath: "",
    mavenSettingsPath: "",
    mavenRepoPath: "",
    mavenArgs: [],
  };
}

async function parseMavenCommand() {
  errorMsg.value = "";
  successMsg.value = "";
  if (!mavenCommandInput.value.trim()) {
    return;
  }
  parsing.value = true;
  try {
    const { ParseMavenCommand } = await import("../../wailsjs/go/app/App");
    const resp = await ParseMavenCommand({ command: mavenCommandInput.value });
    const result = resp.code === 0 ? resp.data : null;
    console.log("Parse response:", resp);
    if (!result) {
      throw new Error(resp.message || "解析失败");
    }
    console.log(
      "[mvn-parse] mavenPath=%s settingsPath=%s repoLocal=%s",
      result?.mavenPath,
      result?.settingsPath,
      result?.repoLocal
    );
    console.log("Before update:", JSON.stringify(settingsStore.systemDefaults));

    if (result.mavenPath) {
      settingsStore.systemDefaults.mavenPath = result.mavenPath.replace(
        /"/g,
        ""
      );
    }
    if (result.settingsPath) {
      settingsStore.systemDefaults.mavenSettingsPath =
        result.settingsPath.replace(/"/g, "");
    }
    if (result.repoLocal) {
      settingsStore.systemDefaults.mavenRepoPath = result.repoLocal.replace(
        /"/g,
        ""
      );
    }
    if (!result.mavenPath) {
      console.warn(
        "[mvn-parse] 未识别 Maven 可执行文件路径；如果路径包含空格，建议加引号"
      );
    }
    if (!result.repoLocal) {
      console.warn("[mvn-parse] 未识别 -Dmaven.repo.local");
    }
    const settingsPath = result.settingsPath
      ? result.settingsPath.replace(/"/g, "")
      : "";
    const repoLocal = result.repoLocal
      ? result.repoLocal.replace(/"/g, "")
      : "";
    const params: string[] = ["clean", "package"];

    const hasSkipTests =
      (Array.isArray(result.argsArray) &&
        result.argsArray.some(
          (a: string) => a && a.startsWith("-DskipTests")
        )) ||
      typeof result.properties?.skipTests !== "undefined";
    if (!hasSkipTests) {
      params.push("-DskipTests");
    }

    if (settingsPath) {
      params.push("-s", settingsPath);
    }
    if (repoLocal) {
      params.push(`-Dmaven.repo.local=${repoLocal}`);
    }

    settingsStore.systemDefaults.mavenArgs = params;

    console.log("After update:", JSON.stringify(settingsStore.systemDefaults));
    successMsg.value = "解析成功！";
  } catch (error: any) {
    console.error("Parse failed:", error);
    errorMsg.value = "解析失败: " + (error.message || error);
  } finally {
    parsing.value = false;
  }
}

async function detectJdk() {
  errorMsg.value = "";
  successMsg.value = "";
  detectingJdk.value = true;
  detectedJdks.value = [];
  try {
    const { StartJDKDetection } = await import("../../wailsjs/go/app/App");
    const resp = await StartJDKDetection();
    if (resp.code !== 0) {
      throw new Error(resp.message || "检测失败");
    }
  } catch (error: any) {
    console.error("Detect JDK failed:", error);
    errorMsg.value = "检测失败: " + (error.message || error);
    detectingJdk.value = false;
  }
}

function selectJdk(jdk: { path: string; source: string }) {
  settingsStore.systemDefaults.jdkPath = jdk.path;
  detectedJdks.value = [];
}
</script>

<template>
  <div class="h-full overflow-y-auto p-6">
    <div class="mx-auto max-w-4xl space-y-6">
      <div class="flex items-start justify-between">
        <div>
          <h1 class="text-3xl font-bold bg-gradient-to-r from-primary to-blue-600 bg-clip-text text-transparent w-fit pb-1">系统设置</h1>
          <p class="text-muted-foreground mt-2 text-sm">配置应用程序的全局设置和默认值</p>
        </div>
        <div class="flex gap-3">
          <button
            class="rounded-md border border-blue-200 bg-blue-50 px-4 py-2 text-sm font-semibold text-blue-700 shadow-sm transition-all hover:bg-blue-100 hover:shadow-md hover:-translate-y-0.5"
            @click="resetToDefaults"
            type="button"
          >
            重置
          </button>
          <button
            class="rounded-md bg-gradient-to-r from-blue-600 to-cyan-500 px-5 py-2 text-sm font-semibold text-white shadow-md transition-all hover:from-blue-700 hover:to-cyan-600 hover:shadow-lg hover:-translate-y-0.5"
            @click="saveSettings"
            :disabled="settingsStore.saving"
            type="button"
          >
            {{ settingsStore.saving ? "保存中..." : "保存设置" }}
          </button>
        </div>
      </div>

      <!-- Tab switcher — segmented control style -->
      <div class="flex p-1 bg-muted rounded-xl gap-1">
        <button
          class="flex-1 py-2.5 px-4 rounded-lg text-sm font-medium transition-all duration-150"
          :class="
            activeTab === 'general'
              ? 'bg-background text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          "
          @click.prevent="activeTab = 'general'"
          type="button"
        >
          通用设置
        </button>
        <button
          class="flex-1 py-2.5 px-4 rounded-lg text-sm font-medium transition-all duration-150"
          :class="
            activeTab === 'defaults'
              ? 'bg-background text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          "
          @click.prevent="activeTab = 'defaults'"
          type="button"
        >
          默认配置
        </button>
      </div>

      <div v-show="activeTab === 'general'" class="space-y-4">
        <div class="rounded-md border p-4">
          <h3 class="text-lg font-medium">部署设置</h3>
          <p class="mb-4 text-sm text-muted-foreground">配置部署相关的全局参数</p>
          <div class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium"
                  >部署超时时间 (秒)</label
                >
                <input
                  v-model.number="settingsStore.globalSettings.defaultTimeout"
                  type="number"
                  class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                  placeholder="600"
                />
              </div>
              <div>
                <label class="block text-sm font-medium">日志保留天数</label>
                <input
                  v-model.number="settingsStore.globalSettings.logRetentionDays"
                  type="number"
                  class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                  placeholder="30"
                />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div
                class="flex items-center justify-between rounded-md border border-border p-3"
              >
                <div>
                  <div class="text-sm font-medium">部署前自动备份</div>
                  <div class="text-xs text-muted-foreground">
                    上传前先备份服务器上的旧文件
                  </div>
                </div>
                <label class="relative inline-flex cursor-pointer items-center">
                  <input
                    type="checkbox"
                    class="peer sr-only"
                    v-model="settingsStore.globalSettings.backupEnabled"
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
                  <div class="text-sm font-medium">云端部署（默认）</div>
                  <div class="text-xs text-muted-foreground">
                    开启后将上传文件并执行远程操作
                  </div>
                </div>
                <label class="relative inline-flex cursor-pointer items-center">
                  <input
                    type="checkbox"
                    class="peer sr-only"
                    v-model="settingsStore.globalSettings.cloudDeploy"
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
                  <div class="text-sm font-medium">离线打包（默认）</div>
                  <div class="text-xs text-muted-foreground">
                    Maven 添加 -o/--offline，不访问远程仓库
                  </div>
                </div>
                <label class="relative inline-flex cursor-pointer items-center">
                  <input
                    type="checkbox"
                    class="peer sr-only"
                    v-model="settingsStore.globalSettings.offlineBuild"
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
                  <div class="text-sm font-medium">轻量日志（默认）</div>
                  <div class="text-xs text-muted-foreground">减少日志输出，提升部署性能</div>
                </div>
                <label class="relative inline-flex cursor-pointer items-center">
                  <input
                    type="checkbox"
                    class="peer sr-only"
                    v-model="settingsStore.globalSettings.lightLog"
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
                  <div class="text-sm font-medium">部署完成后发送通知</div>
                  <div class="text-xs text-muted-foreground">
                    部署成功或失败后发送提醒
                  </div>
                </div>
                <label class="relative inline-flex cursor-pointer items-center">
                  <input
                    type="checkbox"
                    class="peer sr-only"
                    v-model="settingsStore.globalSettings.notifyOnComplete"
                  />
                  <div
                    class="h-6 w-11 rounded-full bg-input peer-checked:bg-primary after:absolute after:left-0.5 after:top-0.5 after:h-5 after:w-5 after:rounded-full after:bg-white after:transition-all peer-checked:after:translate-x-5"
                  ></div>
                </label>
              </div>
            </div>
          </div>
        </div>

        <!-- <div class="rounded-md border p-4">
          <h3 class="text-lg font-medium">外观设置</h3>
          <p class="mb-4 text-sm text-muted-foreground">配置应用程序的外观和语言</p>
          <div class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium">主题</label>
                <select
                  v-model="settingsStore.globalSettings.theme"
                  class="mt-1 block w-full rounded-md border border-input px-3 py-2"
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
                  class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                >
                  <option value="zh-Hans">简体中文</option>
                  <option value="zh-Hant">繁體中文</option>
                  <option value="en">English</option>
                </select>
              </div>
            </div>
          </div>
        </div> -->
      </div>

      <div v-show="activeTab === 'defaults'" class="space-y-4">
        <div class="rounded-md border p-4">
          <h3 class="text-lg font-medium">Maven 命令解析</h3>
          <p class="mb-4 text-sm text-muted-foreground">
            从 IDEA 的 Maven 工具窗口复制命令粘贴到下方解析
          </p>
          <div class="space-y-3">
            <div>
              <textarea
                v-model="mavenCommandInput"
                class="w-full rounded-md border border-input px-3 py-2 font-mono text-sm"
                rows="3"
                placeholder='例如: "D:\Program Files\JetBrains\...\mvn.cmd" -Didea.version=2025.3.3 -s "D:\java_tools\apache-maven-3.9.12\conf\settings_sgt0903.xml" -Dmaven.repo.local=D:\m2\repository package -f pom.xml'
              ></textarea>
            </div>
            <div class="flex items-center gap-2">
              <button
                class="rounded-md bg-blue-500 px-4 py-2 text-sm text-white hover:bg-blue-600 disabled:opacity-50"
                @click="parseMavenCommand"
                :disabled="parsing || !mavenCommandInput.trim()"
                type="button"
              >
                {{ parsing ? "解析中..." : "解析命令" }}
              </button>
              <span v-if="successMsg" class="text-sm text-green-600">{{
                successMsg
              }}</span>
              <span v-if="errorMsg" class="text-sm text-red-600">{{
                errorMsg
              }}</span>
            </div>
          </div>
        </div>

        <div class="rounded-md border p-4">
          <h3 class="text-lg font-medium">系统默认配置</h3>
          <p class="mb-4 text-sm text-muted-foreground">新环境创建时的默认配置值</p>
          <div class="rounded-md bg-yellow-50 border border-yellow-200 p-4 text-sm text-yellow-800 mb-4">
            <strong>提示：</strong>
            这些配置将作为新创建环境的默认值，但可以单独为每个环境覆盖。
          </div>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium">默认 JDK 路径</label>
              <div class="flex gap-2 mt-1">
                <input
                  v-model="settingsStore.systemDefaults.jdkPath"
                  class="flex-1 rounded-md border border-input px-3 py-2"
                  placeholder="C:\Program Files\Java\jdk1.8.0_202"
                />
                <button
                  class="rounded-md bg-green-500 px-4 py-2 text-sm text-white hover:bg-green-600 disabled:opacity-50"
                  @click="detectJdk"
                  :disabled="detectingJdk"
                  type="button"
                >
                  {{ detectingJdk ? "检测中..." : "自动检测" }}
                </button>
              </div>
              <div v-if="detectedJdks.length > 0" class="mt-2 space-y-1">
                <p class="text-xs text-muted-foreground">检测到以下 JDK：</p>
                <div
                  v-for="jdk in detectedJdks"
                  :key="jdk.path"
                  class="cursor-pointer rounded bg-muted px-2 py-1 text-xs hover:bg-muted"
                  @click="selectJdk(jdk)"
                >
                  {{ jdk.path }} ({{ jdk.source }})
                </div>
              </div>
              <div
                v-if="successMsg && detectedJdks.length > 0"
                class="mt-2 text-sm text-green-600"
              >
                {{ successMsg }}
              </div>
              <div v-if="errorMsg" class="mt-2 text-sm text-red-600">
                {{ errorMsg }}
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium">默认 Maven 路径</label>
              <input
                v-model="settingsStore.systemDefaults.mavenPath"
                class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                placeholder="从上方“解析命令”自动回填"
              />
            </div>
            <div>
              <label class="block text-sm font-medium"
                >默认 Maven settings.xml</label
              >
              <input
                v-model="settingsStore.systemDefaults.mavenSettingsPath"
                class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                placeholder="D:\maven\conf\settings.xml"
              />
            </div>
            <div>
              <label class="block text-sm font-medium">默认本地仓库路径</label>
              <input
                v-model="settingsStore.systemDefaults.mavenRepoPath"
                class="mt-1 block w-full rounded-md border border-input px-3 py-2"
                placeholder="D:\m2\repository"
              />
            </div>
            <div>
              <label class="block text-sm font-medium">默认 Maven 参数</label>
              <div class="mt-1 space-y-2">
                <div
                  v-for="(_, idx) in settingsStore.systemDefaults.mavenArgs"
                  :key="idx"
                  class="flex gap-2"
                >
                  <input
                    v-model="settingsStore.systemDefaults.mavenArgs[idx]"
                    class="flex-1 rounded-md border border-input px-3 py-2 font-mono text-sm"
                    placeholder="参数项"
                  />
                  <button
                    class="rounded-md border px-3 py-2 text-sm hover:bg-muted/50"
                    @click="
                      settingsStore.systemDefaults.mavenArgs.splice(idx, 1)
                    "
                    type="button"
                  >
                    删除
                  </button>
                </div>
                <button
                  class="rounded-md border px-3 py-2 text-sm hover:bg-muted/50"
                  @click="settingsStore.systemDefaults.mavenArgs.push('')"
                  type="button"
                >
                  添加参数
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
