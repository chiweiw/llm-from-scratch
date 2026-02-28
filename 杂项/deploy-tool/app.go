package main

import (
	"context"
	"deploy-tool/internal/models"
	"deploy-tool/internal/services"
	"embed"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

type App struct {
	ctx            context.Context
	configService  *services.ConfigService
	deployService  *services.DeployService
	historyService *services.HistoryService
}

func NewApp() *App {
	configService := services.NewConfigService()
	configService.Load()
	configService.Save()

	return &App{
		configService:  configService,
		deployService:  services.NewDeployService(),
		historyService: services.NewHistoryService(),
	}
}

func (a *App) GetEnvironments() []models.Environment {
	return a.configService.GetEnvironments()
}

func (a *App) GetEnvironment(id string) *models.Environment {
	return a.configService.GetEnvironment(id)
}

func (a *App) SaveEnvironment(env models.Environment) error {
	err := a.configService.SaveEnvironment(env)
	if err == nil {
		a.configService.Save()
	}
	return err
}

func (a *App) DeleteEnvironment(id string) error {
	err := a.configService.DeleteEnvironment(id)
	if err == nil {
		a.configService.Save()
	}
	return err
}

func (a *App) CheckEnvironment(envID string) *models.CheckResult {
	env := a.configService.GetEnvironment(envID)
	if env == nil {
		return nil
	}
	return services.CheckEnvironment(env)
}

func (a *App) StartDeploy(envID string, jarIDs []string) error {
	env := a.configService.GetEnvironment(envID)
	if env == nil {
		return fmt.Errorf("环境不存在")
	}
	return a.deployService.Start(envID, jarIDs)
}

func (a *App) CancelDeploy() {
	a.deployService.Cancel()
}

func (a *App) GetDeployProgress() *models.DeployProgress {
	return a.deployService.GetProgress()
}

func (a *App) GetDeployHistory() []models.DeployHistory {
	return a.historyService.GetList()
}

func (a *App) GetGlobalSettings() *models.GlobalSettings {
	return a.configService.GetSettings()
}

func (a *App) SaveGlobalSettings(settings models.GlobalSettings) error {
	err := a.configService.SaveSettings(settings)
	if err == nil {
		a.configService.Save()
	}
	return err
}

func (a *App) GetSystemDefaults() *models.SystemDefaultConfig {
	return a.configService.GetSystemDefaults()
}

func (a *App) SaveSystemDefaults(defaults models.SystemDefaultConfig) error {
	err := a.configService.SaveSystemDefaults(defaults)
	if err == nil {
		a.configService.Save()
	}
	return err
}

func (a *App) ExportConfig(envID string) (string, error) {
	return a.configService.Export(envID)
}

func (a *App) ImportConfig(jsonData string) error {
	err := a.configService.Import(jsonData)
	if err == nil {
		a.configService.Save()
	}
	return err
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "简易发包工具",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Linux: &linux.Options{
			Icon: nil,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
