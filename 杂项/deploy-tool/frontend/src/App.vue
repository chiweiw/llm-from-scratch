<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { computed } from "vue";

const { availableLocales: languages, locale } = useI18n();
const route = useRoute();

const currentTitle = computed(() => {
  return (route.meta.title as string) || "简易发包工具";
});

const onclickLanguageHandle = (item: string) => {
  item !== locale.value ? (locale.value = item) : false;
};

const onclickMinimise = () => {
  window.runtime.WindowMinimise();
};

const onclickQuit = () => {
  window.runtime.Quit();
};

const navItems = [
  { path: "/", name: "环境管理", icon: "📁" },
  { path: "/deploy", name: "部署中心", icon: "🚀" },
  { path: "/history", name: "历史记录", icon: "📋" },
  { path: "/settings", name: "系统设置", icon: "⚙️" },
];
</script>

<template>
  <div class="flex h-screen bg-background text-foreground">
    <aside class="w-56 border-r bg-card flex flex-col">
      <div class="p-4 border-b">
        <h1 class="text-lg font-bold">简易发包工具 v2.0</h1>
      </div>
      <nav class="flex-1 p-2">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
          :class="[
            route.path === item.path
              ? 'bg-primary text-primary-foreground'
              : 'hover:bg-accent hover:text-accent-foreground',
          ]"
        >
          <span>{{ item.icon }}</span>
          <span>{{ item.name }}</span>
        </router-link>
      </nav>
      <div class="p-4 border-t">
        <div class="flex gap-1">
          <button
            v-for="item in languages"
            :key="item"
            @click="onclickLanguageHandle(item)"
            class="px-2 py-1 text-xs rounded transition-colors"
            :class="[
              item === locale
                ? 'bg-primary text-primary-foreground'
                : 'hover:bg-accent',
            ]"
          >
            {{ item === "zh-Hans" ? "中文" : "EN" }}
          </button>
        </div>
      </div>
    </aside>

    <main class="flex-1 flex flex-col overflow-hidden">
      <header
        class="h-12 border-b flex items-center justify-between px-4 bg-card/50"
        style="--wails-draggable: drag"
      >
        <span class="text-sm text-muted-foreground">{{ currentTitle }}</span>
        <div class="flex gap-2" style="--wails-draggable: no-drag">
          <button
            @click="onclickMinimise"
            class="w-8 h-8 rounded hover:bg-accent flex items-center justify-center text-sm"
          >
            ─
          </button>
          <button
            @click="onclickQuit"
            class="w-8 h-8 rounded hover:bg-destructive hover:text-destructive-foreground flex items-center justify-center text-sm"
          >
            ✕
          </button>
        </div>
      </header>

      <div class="flex-1 overflow-auto">
        <router-view v-slot="{ Component }">
          <keep-alive>
            <component :is="Component" />
          </keep-alive>
        </router-view>
      </div>
    </main>
  </div>
</template>

<style lang="scss">
@import url("./assets/css/reset.css");
@import url("./assets/css/font.css");
@import url("./assets/css/globals.css");

html,
body {
  width: 100%;
  height: 100%;
  margin: 0;
  padding: 0;
  font-family: "JetBrainsMono", system-ui, -apple-system, sans-serif;
}

#app {
  width: 100%;
  height: 100%;
  overflow: hidden;
}
</style>
