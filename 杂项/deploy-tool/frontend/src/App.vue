<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

const { availableLocales: languages, locale } = useI18n();
const route = useRoute();

const onclickLanguageHandle = (item: string) => {
  item !== locale.value ? (locale.value = item) : false;
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
    <aside class="w-64 border-r bg-card flex flex-col shadow-lg z-10">
      <div class="p-6 border-b flex items-center justify-center">
        <h1 class="text-xl font-bold bg-gradient-to-r from-primary to-blue-600 bg-clip-text text-transparent">简易发包工具 v2.0</h1>
      </div>
      <nav class="flex-1 px-4 py-8 flex flex-col gap-6">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="flex items-center gap-4 px-6 py-4 rounded-xl text-base font-medium transition-all duration-300 group relative"
          :class="[
            route.path === item.path
              ? 'bg-primary text-primary-foreground shadow-lg scale-105'
              : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground hover:translate-x-1',
          ]"
        >
          <span class="text-2xl group-hover:scale-110 transition-transform duration-300">{{ item.icon }}</span>
          <span class="tracking-wide">{{ item.name }}</span>
        </router-link>
      </nav>
      <div class="p-6 border-t bg-card/50 flex justify-center">
        <div class="flex gap-2 p-1 bg-background/50 rounded-lg">
          <button
            v-for="item in languages"
            :key="item"
            @click="onclickLanguageHandle(item)"
            class="px-4 py-1.5 text-xs font-medium rounded-md transition-all duration-200"
            :class="[
              item === locale
                ? 'bg-primary text-primary-foreground shadow-sm'
                : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
            ]"
          >
            {{ item === "zh-Hans" ? "中文" : "EN" }}
          </button>
        </div>
      </div>
    </aside>

    <main class="flex-1 flex flex-col overflow-hidden">
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
