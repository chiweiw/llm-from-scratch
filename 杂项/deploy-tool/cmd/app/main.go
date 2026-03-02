package appcmd

import (
	"deploy-tool/internal/app"
	"deploy-tool/internal/config"
	"deploy-tool/internal/dao"
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
	appCfg := config.Default()
	configDAO := dao.NewFileConfigDAO(appCfg)

	configService := service.NewConfigService(configDAO)
	configService.Load()

	ipc := app.New(
		configService,
		service.NewDeployService(),
		service.NewHistoryService(),
	)

	webviewUserDataPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "deploy-tool", "webview2")
	if strings.TrimSpace(os.Getenv("LOCALAPPDATA")) == "" {
		webviewUserDataPath = filepath.Join(os.TempDir(), "deploy-tool", "webview2")
	}
	disableGPU := strings.TrimSpace(strings.ToLower(os.Getenv("DEPLOY_TOOL_WEBVIEW2_DISABLE_GPU"))) == "1"
	disableRCI := strings.TrimSpace(strings.ToLower(os.Getenv("DEPLOY_TOOL_WEBVIEW2_DISABLE_RCI"))) == "1"

	err := wails.Run(&options.App{
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
