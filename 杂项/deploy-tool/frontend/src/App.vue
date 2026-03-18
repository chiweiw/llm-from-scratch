<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { FolderOpen, Rocket, History, Settings } from "lucide-vue-next";

const { availableLocales: languages, locale } = useI18n();
const route = useRoute();

const onclickLanguageHandle = (item: string) => {
  item !== locale.value ? (locale.value = item) : false;
};

const navItems = [
  { path: "/", name: "环境管理", icon: FolderOpen },
  { path: "/deploy", name: "部署中心", icon: Rocket },
  { path: "/history", name: "历史记录", icon: History },
  { path: "/settings", name: "系统设置", icon: Settings },
];
</script>

<template>
  <div class="flex h-screen bg-background text-foreground">
    <!-- Sidebar -->
    <aside class="w-72 border-r bg-card flex flex-col shrink-0">
      <!-- Logo -->
      <div class="px-6 py-6 border-b">
        <div class="flex items-center gap-4">
          <div
            class="w-10 h-10 rounded-xl bg-gradient-to-br from-primary to-primary/75 flex items-center justify-center shrink-0 shadow-sm"
          >
            <Rocket :size="20" class="text-primary-foreground" />
          </div>
          <div>
            <div class="text-base font-bold leading-tight">简易发包工具</div>
            <div class="text-xs text-muted-foreground mt-0.5">Deployment Tool v2.0</div>
          </div>
        </div>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 p-4 space-y-1">
        <div class="px-3 pb-3 pt-1">
          <span class="text-xs font-semibold uppercase tracking-widest text-muted-foreground/50">导航</span>
        </div>
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="group relative flex items-center gap-3 px-3 py-3 rounded-xl text-sm transition-all duration-200"
          :class="[
            route.path === item.path
              ? 'bg-primary text-primary-foreground font-medium shadow-md shadow-primary/20'
              : 'font-medium text-muted-foreground hover:bg-muted/80 hover:text-foreground',
          ]"
        >
          <div
            class="flex items-center justify-center w-6 h-6 shrink-0 transition-all duration-200"
            :class="[
              route.path === item.path
                ? 'text-primary-foreground'
                : 'text-muted-foreground group-hover:text-foreground',
            ]"
          >
            <component :is="item.icon" :size="18" />
          </div>
          <span>{{ item.name }}</span>
        </router-link>
      </nav>

      <!-- Language switcher -->
      <div class="px-6 pb-6 pt-4 border-t">
        <div class="px-1 mb-2">
          <span class="text-xs font-semibold uppercase tracking-widest text-muted-foreground/50">Language</span>
        </div>
        <div class="flex p-1.5 bg-muted rounded-xl">
          <button
            v-for="item in languages"
            :key="item"
            @click="onclickLanguageHandle(item)"
            class="flex-1 py-2 text-xs font-medium rounded-lg transition-all duration-200"
            :class="[
              item === locale
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground',
            ]"
          >
            {{ item === "zh-Hans" ? "中文" : "EN" }}
          </button>
        </div>
      </div>
    </aside>

    <!-- Main content -->
    <main class="flex-1 overflow-hidden">
      <div class="h-full overflow-auto">
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
