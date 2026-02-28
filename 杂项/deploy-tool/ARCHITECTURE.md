# 简易发包工具桌面版 - 架构文档

## 项目概述

**项目名称**: 简易发包工具桌面版 (DeployTool Desktop)  
**技术栈**: Wails v2.11.0 + Go 1.24.13 + Vue 3 + TypeScript + Tailwind CSS  
**版本号**: v2.0  
**日期**: 2026-02-28

---

## 1. 系统架构

### 1.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                         桌面应用 (Wails)                             │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                      前端层 (Vue.js)                         │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │   │
│  │  │   Views     │  │   Stores    │  │     Components      │  │   │
│  │  │  (页面视图)  │  │  (状态管理)  │  │     (UI组件)        │  │   │
│  │  │             │  │             │  │                     │  │   │
│  │  │ Environment │  │ environment │  │ shadcn-vue          │  │   │
│  │  │ Deploy      │  │ deploy      │  │ Tailwind CSS        │  │   │
│  │  │ History     │  │ history     │  │ Lucide Icons        │  │   │
│  │  │ Settings    │  │ settings    │  │                     │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│                         Wails Bridge                               │
│                              │                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                      后端层 (Go)                             │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │   │
│  │  │   Models    │  │  Services   │  │   App Bindings      │  │   │
│  │  │  (数据模型)  │  │  (业务逻辑)  │  │   (Wails绑定)       │  │   │
│  │  │             │  │             │  │                     │  │   │
│  │  │ Environment │  │ Environment │  │ GetEnvironments     │  │   │
│  │  │ Deploy      │  │ Deploy      │  │ StartDeploy         │  │   │
│  │  │ History     │  │ History     │  │ GetDeployProgress   │  │   │
│  │  │ Config      │  │ Config      │  │ ...                 │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

### 1.2 技术选型

| 层级 | 技术 | 版本 | 用途 |
|------|------|------|------|
| 桌面框架 | Wails | v2.11.0 | 跨平台桌面应用框架 |
| 后端语言 | Go | 1.24.13 | 业务逻辑处理 |
| 前端框架 | Vue | 3.2.38 | UI 框架 |
| 语言 | TypeScript | 4.7.4 | 类型安全 |
| 状态管理 | Pinia | 2.0.21 | 全局状态管理 |
| 路由 | Vue Router | 4.1.5 | 页面路由 |
| UI 样式 | Tailwind CSS | 3.1.8 | 原子化 CSS |
| 组件库 | shadcn-vue | latest | 现代化 UI 组件 |
| 图标 | Lucide Vue | latest | 图标库 |
| 国际化 | Vue I18n | 9.2.2 | 多语言支持 |
| 构建工具 | Vite | 3.0.9 | 前端构建 |

---

## 2. 前端架构

### 2.1 目录结构

```
frontend/
├── src/
│   ├── assets/                 # 静态资源
│   │   ├── css/               # 样式文件
│   │   │   ├── globals.css    # 全局样式 (shadcn-vue 变量)
│   │   │   ├── font.css       # 字体定义
│   │   │   └── reset.css      # CSS 重置
│   │   ├── fonts/             # JetBrainsMono 字体
│   │   └── images/            # 图片资源
│   │
│   ├── components/            # Vue 组件
│   │   └── HelloWorld.vue     # 示例组件
│   │
│   ├── i18n/                  # 国际化
│   │   ├── locales/
│   │   │   ├── en.json        # 英文翻译
│   │   │   └── zh-Hans.json   # 简体中文翻译
│   │   └── index.ts           # i18n 配置
│   │
│   ├── lib/                   # 工具库
│   │   └── utils.ts           # cn() 工具函数 (shadcn-vue)
│   │
│   ├── router/                # 路由配置
│   │   └── index.ts           # Vue Router 配置
│   │
│   ├── stores/                # Pinia 状态管理
│   │   ├── environment.ts     # 环境管理 Store
│   │   ├── deploy.ts          # 部署 Store
│   │   ├── history.ts         # 历史记录 Store
│   │   └── counter.ts         # 示例 Store
│   │
│   ├── types/                 # TypeScript 类型定义
│   │   └── index.ts           # 全局类型定义
│   │
│   ├── views/                 # 页面视图
│   │   ├── EnvironmentView.vue   # 环境管理页面
│   │   ├── DeployView.vue        # 部署中心页面
│   │   ├── HistoryView.vue       # 历史记录页面
│   │   ├── SettingsView.vue      # 系统设置页面
│   │   ├── HomeView.vue          # 首页 (模板自带)
│   │   └── AboutView.vue         # 关于页面 (模板自带)
│   │
│   ├── App.vue                # 根组件
│   ├── main.ts                # 入口文件
│   └── style.scss             # 全局 SCSS
│
├── wailsjs/                   # Wails 自动生成
│   ├── go/
│   │   └── main/
│   │       ├── App.d.ts       # Go 方法 TypeScript 定义
│   │       └── App.js         # Go 方法 JavaScript 绑定
│   └── runtime/
│       ├── runtime.d.ts       # Wails Runtime 类型
│       └── runtime.js         # Wails Runtime 方法
│
├── tailwind.config.cjs        # Tailwind CSS 配置
├── postcss.config.js          # PostCSS 配置
├── vite.config.ts             # Vite 配置
└── package.json               # 依赖管理
```

### 2.2 状态管理 (Pinia)

#### environment.ts - 环境管理 Store
```typescript
state: {
  environments: Environment[],      // 环境列表
  currentEnvironment: Environment | null,  // 当前选中环境
  checkResult: CheckResult | null,  // 自检结果
  loading: boolean                  // 加载状态
}

actions:
  - fetchEnvironments()             // 获取环境列表
  - saveEnvironment(env)            // 保存环境
  - deleteEnvironment(id)           // 删除环境
  - checkEnvironment(id)            // 环境自检
```

#### deploy.ts - 部署 Store
```typescript
state: {
  progress: DeployProgress | null,  // 部署进度
  isDeploying: boolean,             // 是否部署中
  selectedJarIds: string[]          // 选中的 Jar 包
}

actions:
  - setSelectedJars(jarIds)         // 设置选中 Jar
  - startDeploy(envId)              // 开始部署
  - cancelDeploy()                  // 取消部署
  - fetchProgress()                 // 获取进度
```

#### history.ts - 历史记录 Store
```typescript
state: {
  histories: DeployHistory[],       // 历史记录列表
  currentHistory: DeployHistory | null,
  loading: boolean
}

actions:
  - fetchHistories(filter)          // 获取历史记录
  - fetchHistoryDetail(id)          // 获取详情
```

### 2.3 路由配置

```typescript
routes: [
  { path: "/", name: "environment", component: EnvironmentView },
  { path: "/deploy", name: "deploy", component: DeployView },
  { path: "/history", name: "history", component: HistoryView },
  { path: "/settings", name: "settings", component: SettingsView }
]
```

### 2.4 UI 设计系统

#### 颜色变量 (globals.css)
```css
:root {
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  --card: 0 0% 100%;
  --card-foreground: 222.2 84% 4.9%;
  --primary: 222.2 47.4% 11.2%;
  --primary-foreground: 210 40% 98%;
  --secondary: 210 40% 96.1%;
  --secondary-foreground: 222.2 47.4% 11.2%;
  --muted: 210 40% 96.1%;
  --muted-foreground: 215.4 16.3% 46.9%;
  --accent: 210 40% 96.1%;
  --accent-foreground: 222.2 47.4% 11.2%;
  --destructive: 0 84.2% 60.2%;
  --destructive-foreground: 210 40% 98%;
  --border: 214.3 31.8% 91.4%;
  --input: 214.3 31.8% 91.4%;
  --ring: 222.2 84% 4.9%;
  --radius: 0.5rem;
}
```

---

## 3. 后端架构

### 3.1 目录结构

```
deploy-tool/
├── internal/                  # 内部包
│   ├── models/               # 数据模型
│   │   ├── environment.go    # 环境配置模型
│   │   ├── check.go          # 自检结果模型
│   │   ├── deploy.go         # 部署进度模型
│   │   ├── history.go        # 历史记录模型
│   │   └── config.go         # 全局配置模型
│   │
│   └── services/             # 业务服务层
│       ├── environment_service.go   # 环境管理服务
│       ├── deploy_service.go        # 部署服务
│       ├── config_service.go        # 配置服务
│       └── history_service.go       # 历史记录服务
│
├── main.go                   # 应用入口
├── app.go                    # Wails App 绑定
├── go.mod                    # Go 模块定义
└── wails.json                # Wails 配置
```

### 3.2 数据模型

#### Environment - 环境配置
```go
type Environment struct {
    ID          string         // 环境ID
    Name        string         // 环境名称
    Identifier  string         // 环境标识 (dev/test/prod)
    Description string         // 环境描述
    Local       LocalConfig    // 本地配置
    Servers     []ServerConfig // 服务器列表
    TargetFiles []TargetFile   // 目标文件列表
    RenameRule  string         // 改名规则
}
```

#### LocalConfig - 本地环境配置
```go
type LocalConfig struct {
    ProjectRoot        string // 项目根目录
    JdkPath           string // JDK 路径
    MavenPath         string // Maven 路径
    MavenSettingsPath string // Maven settings.xml 路径
    MavenRepoPath     string // Maven 本地仓库
    MavenArgs         string // Maven 参数
    QuietMode         bool   // 安静模式
    VerboseOutput     bool   // 精简日志
    SpecifyPom        bool   // 显式指定 pom
    OfflineBuild      bool   // 离线构建
}
```

#### ServerConfig - 服务器配置
```go
type ServerConfig struct {
    ID            string // 服务器ID
    Name          string // 服务器名称
    Host          string // 主机地址
    Port          int    // SSH 端口
    Username      string // 用户名
    Password      string // 密码 (加密存储)
    DeployDir     string // 远程部署目录
    RestartScript string // 重启脚本路径
    EnableRestart bool   // 是否启用重启
    UseSudo       bool   // 是否使用 sudo
}
```

#### DeployProgress - 部署进度
```go
type DeployProgress struct {
    EnvironmentID string         // 环境ID
    Status        string         // 状态 (idle/running/success/failed/canceled)
    CurrentStep   string         // 当前步骤
    TotalProgress int            // 总进度 0-100
    Steps         []StepProgress // 步骤进度
    CurrentFile   string         // 当前上传文件
    FileProgress  int            // 文件进度
    Speed         string         // 上传速度
    StartTime     int64          // 开始时间
    EndTime       int64          // 结束时间
    ErrorMessage  string         // 错误信息
}
```

### 3.3 服务层

#### EnvironmentService - 环境管理
```go
- GetAll() []Environment                    // 获取所有环境
- GetByID(id string) *Environment           // 根据ID获取环境
- Save(env Environment) error               // 保存环境
- Delete(id string) error                   // 删除环境
- Duplicate(id string) (*Environment, error) // 复制环境
- CheckLocal(envID string) *CheckResult     // 本地环境自检
- CheckRemote(envID string) *CheckResult    // 远程环境自检
```

#### DeployService - 部署服务
```go
- Start(envID string, jarIDs []string) error // 开始部署
- Cancel()                                  // 取消部署
- GetProgress() *DeployProgress             // 获取部署进度
```

#### ConfigService - 配置服务
```go
- Load()                                    // 加载配置
- Save()                                    // 保存配置
- Export(envID string) (string, error)      // 导出配置
- Import(jsonData string) error             // 导入配置
- GetSettings() *GlobalSettings             // 获取全局设置
- SaveSettings(settings GlobalSettings) error // 保存全局设置
```

#### HistoryService - 历史记录服务
```go
- GetList(filter HistoryFilter) []DeployHistory // 获取历史列表
- GetDetail(id string) *DeployHistory           // 获取历史详情
- Add(record DeployHistory)                     // 添加历史记录
```

### 3.4 Wails 绑定方法

```go
// App.go 中暴露给前端的方法

// 环境管理
GetEnvironments() []models.Environment
GetEnvironment(id string) *models.Environment
SaveEnvironment(env models.Environment) error
DeleteEnvironment(id string) error
DuplicateEnvironment(id string) (*models.Environment, error)

// 环境自检
CheckLocalEnvironment(envID string) *models.CheckResult
CheckRemoteEnvironment(envID string) *models.CheckResult

// 部署操作
StartDeploy(envID string, jarIDs []string) error
CancelDeploy()
GetDeployProgress() *models.DeployProgress

// 历史记录
GetDeployHistory(filter models.HistoryFilter) []models.DeployHistory
GetHistoryDetail(id string) *models.DeployHistory

// 配置
ExportConfig(envID string) (string, error)
ImportConfig(jsonData string) error
GetGlobalSettings() *models.GlobalSettings
SaveGlobalSettings(settings models.GlobalSettings) error
```

---

## 4. 通信机制

### 4.1 前端调用后端

```typescript
// 示例：获取环境列表
const { GetEnvironments } = await import('../../wailsjs/go/main/App');
const environments = await GetEnvironments();

// 示例：开始部署
const { StartDeploy } = await import('../../wailsjs/go/main/App');
await StartDeploy(envId, selectedJarIds);
```

### 4.2 后端调用前端 (Events)

```go
// 后端发送事件
runtime.EventsEmit(ctx, "deploy:progress", progress)

// 前端监听事件
window.runtime.EventsOn("deploy:progress", (progress) => {
  console.log(progress);
});
```

---

## 5. 功能模块

### 5.1 环境管理模块

```
┌─────────────────────────────────────────────────────────────┐
│ 环境管理                                                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌───────────────┐  ┌─────────────────────────────────────┐ │
│  │ 环境列表       │  │ 环境详情                            │ │
│  │               │  │                                     │ │
│  │ • 开发环境     │  │  📋 基本信息                        │ │
│  │ • 测试环境     │  │  📁 本地配置                        │ │
│  │ • 生产环境     │  │  🖥️ 服务器配置                      │ │
│  │               │  │  🏷️ 改名规则                        │ │
│  │ [+ 添加环境]   │  │  📦 目标文件                        │ │
│  │               │  │                                     │ │
│  └───────────────┘  │ [保存] [删除] [复制] [自检]         │ │
│                     └─────────────────────────────────────┘ │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 5.2 部署中心模块

```
┌─────────────────────────────────────────────────────────────┐
│ 部署中心                                                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  选择环境: [开发环境 ▼]                     [🔄 刷新状态]   │
│                                                             │
│  自检状态:                                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ ✅ 项目根目录存在                                   │   │
│  │ ✅ JDK 环境正常                                    │   │
│  │ ✅ Maven 可用                                      │   │
│  │ ✅ 服务器连接正常                                  │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  📦 选择要部署的文件:                                       │
│  ☑ platform-startup-project.jar                            │
│  ☑ platform-startup-system.jar                             │
│                                                             │
│  部署进度:                                                  │
│  [████████████████░░░░] 80%                                │
│                                                             │
│  ✅ 环境自检     ████████████ 100%                          │
│  ✅ Maven 打包   ████████████ 100%                          │
│  🔄 文件上传     ████████░░░░ 80%                           │
│  ⏸️ 远程重启     ░░░░░░░░░░░░ 0%                            │
│                                                             │
│  [开始部署] [取消部署]                                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 6. 开发规范

### 6.1 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件 | PascalCase | `EnvironmentView.vue` |
| 文件 | camelCase | `environmentService.go` |
| 类型 | PascalCase | `Environment`, `DeployProgress` |
| 变量 | camelCase | `currentEnvironment` |
| 常量 | UPPER_SNAKE_CASE | `DEPLOY_STATUS_RUNNING` |
| 方法 | camelCase | `getEnvironments()` |

### 6.2 目录命名

- 小写字母
- 使用连字符分隔单词 (kebab-case)
- 示例: `environment-view`, `deploy-center`

### 6.3 代码组织

```
// Vue 组件结构
<script setup lang="ts">
// 1. imports
// 2. types/interfaces
// 3. props/emits
// 4. reactive state
// 5. computed
// 6. methods
// 7. lifecycle hooks
</script>

<template>
  <!-- 模板内容 -->
</template>

<style scoped>
/* 样式 */
</style>
```

---

## 7. 构建与部署

### 7.1 开发模式

```bash
# 启动开发服务器 (热重载)
cd deploy-tool && wails dev

# 前端单独开发
cd deploy-tool/frontend && npm run dev
```

### 7.2 生产构建

```bash
# 构建生产版本
cd deploy-tool && wails build

# 构建 Windows 安装包
cd deploy-tool && wails build -platform windows/amd64

# 构建时清理
cd deploy-tool && wails build -clean
```

### 7.3 输出目录

```
build/
├── bin/
│   └── deploy-tool.exe        # Windows 可执行文件
└── installer/
    └── deploy-tool-installer.exe  # Windows 安装包
```

---

## 8. 扩展计划

### 8.1 后续功能

1. **Maven 打包集成** - 调用本地 Maven 执行打包
2. **SSH 文件上传** - 使用 SFTP 上传 Jar 包到服务器
3. **远程命令执行** - SSH 执行重启脚本
4. **实时日志流** - WebSocket 推送部署日志
5. **部署历史图表** - 可视化部署统计
6. **多语言完善** - 完善国际化支持

### 8.2 技术债务

1. 完善错误处理机制
2. 添加单元测试
3. 实现数据持久化 (SQLite/BoltDB)
4. 优化前端性能
5. 添加操作日志

---

## 9. 参考资料

- [Wails 官方文档](https://wails.io/docs/)
- [Vue 3 文档](https://vuejs.org/)
- [Tailwind CSS 文档](https://tailwindcss.com/)
- [shadcn-vue 文档](https://www.shadcn-vue.com/)
- [Pinia 文档](https://pinia.vuejs.org/)

---

**文档版本**: 1.0  
**最后更新**: 2026-02-28  
**作者**: AI Assistant
