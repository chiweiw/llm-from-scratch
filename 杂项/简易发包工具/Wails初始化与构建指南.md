# Wails 初始化与构建指南（Windows 10/11）

## 目标
- 在 Windows 上使用 Wails 初始化前后端并完成开发与打包
- 支持两类项目：
  - 最小模板（vanilla）：嵌入静态资源，适合无前端依赖的场景
  - 前端框架（React/Vue/Svelte/Vite）：使用 Node 与 npm 进行前端构建

## 环境准备
- Go：建议 1.22.x（并关闭自动工具链下载）
  - 关闭自动工具链：在终端执行  
    `go env -w GOTOOLCHAIN=local`
- Node：安装 LTS（推荐 v24.14.0；Node 20 也可用）
  - 下载页（Windows MSI/ZIP）：https://nodejs.org/en/download/
  - 验证：`node -v` 和 `npm -v`
- WebView2 Runtime：Windows 桌面渲染必需（Win11通常自带；Win10若缺失请安装）
  - 检测与安装可用 Wails Doctor 或安装包
- Wails CLI：  
  `go install github.com/wailsapp/wails/v2/cmd/wails@v2.11.0`  
  验证：`wails version`
- 可选工具
  - UPX（压缩 exe）：https://upx.github.io/
  - NSIS（安装包）：https://wails.io/docs/guides/windows-installer/

## 自检
- 运行：`wails doctor`  
  - WebView2 Installed  
  - Node/npm Available（或 Installed）  
  - 若缺失，按提示安装

## 初始化项目
### 1) 最小模板（vanilla，无需前端依赖）
- 在目标目录执行：  
  `wails init -n MinimalApp -t vanilla`
- 结构说明：
  - 生成 wails.json、frontend 与 build 目录
  - 默认配置了 `frontend:install` 与 `frontend:build`，但可按需跳过

### 2) 前端框架模板（以 React 为例）
- 在目标目录执行：  
  `wails init -n MyApp -t react`
- 其他可用模板：`vue`、`svelte`、`vite`

## 开发与调试
- 开发模式（前后端一起运行，热更新）：
  - 进入项目目录：`cd <项目目录>`
  - 执行：`wails dev`
  - 说明：会自动启动前端 dev server 与后端（dev 标签构建）
- VSCode 调试（Go 断点，dev 标签）：
  - 在 `.vscode/launch.json` 增加：
    ```json
    {
      "version": "0.2.0",
      "configurations": [
        {
          "name": "Wails Dev (Go Debug)",
          "type": "go",
          "request": "launch",
          "mode": "debug",
          "program": "${workspaceFolder}/<项目目录>",
          "buildFlags": "-tags=dev"
        }
      ]
    }
    ```

## 构建发行版
- 标准构建（完整前端 + 资源打包）：  
  `wails build`
- 跳过前端依赖安装（依赖已安装且 package.json 未变化）：  
  `wails build -s`
- 仅嵌入资源（不打包图标/manifest 等，适合 go:embed 静态前端）：  
  `wails build -s -nopackage`
- 产物位置：`build/bin/<项目名>.exe`
- 可选：使用 UPX 压缩 exe（需安装 UPX）  
  `upx build/bin/<项目名>.exe`
- 可选：生成 NSIS 安装包（需安装 NSIS）  
  `wails build -nsis`
  - 安装器位于 `build/bin`

## 常见问题
- npm/Node 未找到
  - 确认 `node -v` 与 `npm -v` 正常
  - 若终端 PATH 未刷新，重启终端/VSCode 或将 `C:\Program Files\nodejs` 加入 PATH
  - 也可手动安装依赖：
    ```
    cd frontend
    npm install
    cd ..
    wails build
    ```
- Go 自动下载 toolchain（例如尝试 go1.23）
  - 设置本地工具链模式：`go env -w GOTOOLCHAIN=local`
  - 如 go.mod 出现 `go 1.23`，改为 `go 1.22` 并执行 `go mod tidy`
- 启动出现控制台窗口
  - 手动编译时需使用 GUI 子系统：  
    `go build -tags "desktop,production" -ldflags "-H=windowsgui -s -w" -o bin/app.exe`
  - 使用 Wails CLI 的 `wails build` 默认会设置合适的参数
- libpng 警告（iCCP/cHRM）
  - 替换为标准 sRGB PNG 图标（重新导出，去掉非 sRGB 配置）

## 验证流程（建议）
1. `wails doctor` 检查依赖（WebView2/Node/npm）
2. 初始化模板（vanilla 或 react）
3. `wails dev` 本地开发与调试
4. `wails build` 生成发行版（可选 `-s` 加速）
5. 运行 `build/bin/<项目名>.exe` 验证
6. 可选：`wails build -nsis` 生成安装包；`upx` 压缩体积

## 备注
- Wails 作者正在构建 v3（来自项目维护者介绍），如后续升级，请参考官方迁移指南
- 生产环境建议保留 `wails.json` 中的 Info 字段（公司、产品名、版本、版权）以便安装包生成与元数据管理
