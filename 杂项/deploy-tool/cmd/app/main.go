package appcmd

import (
	"context"
	"deploy-tool/internal/app"
	"deploy-tool/internal/db"
	"deploy-tool/internal/logger"
	"deploy-tool/internal/service"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

func Run(assets fs.FS) error {
	logger.Init(0)

	database, err := db.Init("")
	if err != nil {
		logger.Error("初始化数据库失败: %v", err)
		return err
	}
	defer database.Close()

	envDAO := db.NewEnvironmentDAO(database)
	globalSettingDAO := db.NewGlobalSettingDAO(database)
	systemDefaultDAO := db.NewSystemDefaultDAO(database)
	serverConfigDAO := db.NewServerConfigDAO(database)
	targetFileDAO := db.NewTargetFileDAO(database)
	deployHistoryDAO := db.NewDeployHistoryDAO(database)
	deployLogDAO := db.NewDeployLogDAO(database)

	configService := service.NewConfigService(
		envDAO,
		globalSettingDAO,
		systemDefaultDAO,
		serverConfigDAO,
		targetFileDAO,
	)
	configService.Load()

	historyService := service.NewHistoryService(deployHistoryDAO, deployLogDAO)

	deployService := service.NewDeployService(configService, historyService, context.Background())
	deployService.SetConfigService(configService)
	deployService.SetHistoryService(historyService)

	ipc := app.New(
		configService,
		deployService,
		historyService,
	)

	webviewUserDataPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "deploy-tool", "webview2")
	if strings.TrimSpace(os.Getenv("LOCALAPPDATA")) == "" {
		webviewUserDataPath = filepath.Join(os.TempDir(), "deploy-tool", "webview2")
	}
	disableGPU := strings.TrimSpace(strings.ToLower(os.Getenv("DEPLOY_TOOL_WEBVIEW2_DISABLE_GPU"))) == "1"
	disableRCI := strings.TrimSpace(strings.ToLower(os.Getenv("DEPLOY_TOOL_WEBVIEW2_DISABLE_RCI"))) == "1"

	err = wails.Run(&options.App{
		Title:  "简易发包工具",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnStartup:        ipc.Startup,
		Bind: []interface{}{
			ipc,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:                false,
			WindowIsTranslucent:                 false,
			DisableWindowIcon:                   false,
			WebviewUserDataPath:                 webviewUserDataPath,
			WebviewGpuIsDisabled:                disableGPU,
			WebviewDisableRendererCodeIntegrity: disableRCI,
		},
		Linux: &linux.Options{
			Icon: nil,
		},
	})
	return err
}
