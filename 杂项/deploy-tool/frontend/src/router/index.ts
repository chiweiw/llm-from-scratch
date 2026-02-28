import { createRouter, createWebHashHistory } from "vue-router";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: "/",
      name: "environment",
      component: () => import("../views/EnvironmentView.vue"),
      meta: { title: "环境管理" },
    },
    {
      path: "/deploy",
      name: "deploy",
      component: () => import("../views/DeployView.vue"),
      meta: { title: "部署中心" },
    },
    {
      path: "/history",
      name: "history",
      component: () => import("../views/HistoryView.vue"),
      meta: { title: "历史记录" },
    },
    {
      path: "/settings",
      name: "settings",
      component: () => import("../views/SettingsView.vue"),
      meta: { title: "系统设置" },
    },
  ],
});

export default router;
