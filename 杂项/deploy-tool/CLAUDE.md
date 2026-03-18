# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**简易发包工具桌面版** (DeployTool) — a cross-platform desktop deployment tool built with **Wails v2 + Go + Vue 3 + TypeScript**. It automates backend (Maven) and frontend (npm) build-and-deploy pipelines to remote servers via SFTP/SSH.

## Commands

### Development

```bash
# Start full dev mode (Go backend + Vue frontend, hot reload)
wails dev

# Frontend only (runs at localhost:5173)
cd frontend && npm run dev
```

### Build

```bash
# Production build (outputs to build/bin/)
wails build

# Windows amd64
wails build -platform windows/amd64

# Clean build
wails build -clean
```

### Frontend

```bash
cd frontend
npm install          # install deps
npm run build        # type-check + build
npm run build-only   # build without type-check
npm run lint         # ESLint fix
```

### Go Tests

```bash
# Run all Go tests
go test ./...

# Run a specific test file
go test ./internal/service/ -run TestParseMavenCommand_IDEACommandLine -v
```

## Architecture

### Communication Flow

```
Vue frontend  →  Wails JS bridge  →  internal/app/App  →  service layer  →  DAO layer  →  SQLite DB
```

The Wails bridge auto-generates TypeScript bindings in `frontend/wailsjs/go/main/App.{js,d.ts}` from the exported methods on `App` in `internal/app/app.go`. The frontend calls these like regular async functions. The backend emits real-time events (log lines, JDK detection results) via `wailsRuntime.EventsEmit`.

### Entry Points

- `main.go` + `cmd/app/main.go` — embeds the `frontend/dist` filesystem and calls `wails.Run()`
- `internal/app/app.go` — the `App` struct; all methods on it become IPC-callable from the frontend
- `wails.json` — Wails config; build entry is `./cmd/app`

### Backend (`internal/`)

| Package | Role |
|---|---|
| `app` | Wails-bound IPC handler; wires services together; dispatches log events |
| `service` | Business logic: `ConfigService`, `DeployService`, `HistoryService`, `MavenBuildService`, `CheckEnvironment` |
| `dao` | `ConfigDAO` — thin wrappers over `db` DAOs |
| `db` | SQLite init, schema creation, migrations, DAO implementations |
| `model/entity` | Shared domain types: `Environment`, `DeployProgress`, `GlobalSettings`, `SystemDefaultConfig`, etc. |
| `model/request` | Request wrapper structs used as IPC method parameters |
| `model/response` | Generic `Base` / `Data[T]` response envelope (`Code 0` = OK, `Code 1` = error) |
| `utils` | SFTP client, JDK auto-detect, ZIP utilities |
| `logger` | Structured logger; hooks into Wails event emitter for real-time log streaming to frontend |

**All IPC methods return `response.Base` or `response.Data[T]`** — never raw errors. The frontend checks `resp.code === 0`.

### Deploy Pipeline

`DeployService.Start()` runs asynchronously in a goroutine. Two paths:

- **Backend** (`buildType == "backend"`): Environment check → Maven build → SFTP upload each target file → SSH restart script
- **Frontend** (`buildType == "frontend"`): Environment check → `npm run build` → zip `dist/` → SFTP upload `dist.zip` → remote backup + unzip via SSH

Progress is tracked in `entity.DeployProgress` with per-step `StepProgress` structs, protected by `sync.RWMutex`. The frontend polls `GetDeployProgress()` to render live status.

### Database

SQLite via `modernc.org/sqlite` (pure Go, no CGO). Database file lives next to the executable (`deploy-tool.db`). Schema is created in `internal/db/database.go`; incremental migrations run via a `migrations` table. Key tables: `environments`, `server_configs`, `target_files`, `deploy_history`, `deploy_logs`, `global_settings`, `system_defaults`.

### Frontend (`frontend/src/`)

- **Views**: `EnvironmentView`, `DeployView`, `HistoryView`, `SettingsView`
- **Stores** (Pinia): `environment`, `deploy`, `history`, `settings`
- **Types**: `frontend/src/types/index.ts` — mirrors Go entity structs
- **UI**: shadcn-vue components + Tailwind CSS + Lucide icons
- **i18n**: `frontend/src/i18n/locales/{en,zh-Hans}.json`

### Naming Conventions

- Go files: `snake_case`; exported types/methods: `PascalCase`
- Vue components: `PascalCase`; stores/utils: `camelCase`
- Constants: `UPPER_SNAKE_CASE`
- Vue SFC script order: imports → types → props/emits → reactive state → computed → methods → lifecycle hooks
