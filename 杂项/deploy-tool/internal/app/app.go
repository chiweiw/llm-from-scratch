package app

import (
	"context"
	"deploy-tool/internal/model/entity"
	"deploy-tool/internal/model/request"
	"deploy-tool/internal/model/response"
	"deploy-tool/internal/service"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	configService *service.ConfigService
	deployService *service.DeployService
	history       *service.HistoryService
}

func New(cfg *service.ConfigService, deploy *service.DeployService, history *service.HistoryService) *App {
	return &App{
		configService: cfg,
		deployService: deploy,
		history:       history,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetEnvironments() response.Data[[]entity.Environment] {
	return response.OKData(a.configService.GetEnvironments())
}

func (a *App) GetEnvironment(id string) response.Data[*entity.Environment] {
	if id == "" {
		return response.FailData("id 不能为空", (*entity.Environment)(nil))
	}
	env := a.configService.GetEnvironment(id)
	if env == nil {
		return response.FailData("环境不存在", (*entity.Environment)(nil))
	}
	return response.OKData(env)
}

func (a *App) SaveEnvironment(req request.SaveEnvironment) response.Base {
	if req.Environment.ID == "" {
		return response.Fail("环境 id 不能为空")
	}
	if err := a.configService.UpsertEnvironment(req.Environment); err != nil {
		return response.Fail(err.Error())
	}
	return response.OK()
}

func (a *App) DeleteEnvironment(req request.DeleteEnvironment) response.Base {
	if req.ID == "" {
		return response.Fail("id 不能为空")
	}
	if err := a.configService.DeleteEnvironment(req.ID); err != nil {
		return response.Fail(err.Error())
	}
	return response.OK()
}

func (a *App) CheckEnvironment(req request.CheckEnvironment) response.Data[*entity.CheckResult] {
	if req.ID == "" {
		return response.FailData("id 不能为空", (*entity.CheckResult)(nil))
	}
	env := a.configService.GetEnvironment(req.ID)
	if env == nil {
		return response.FailData("环境不存在", (*entity.CheckResult)(nil))
	}
	defaults := a.configService.GetSystemDefaults()
	result := service.CheckEnvironment(env, defaults)
	return response.OKData(result)
}

func (a *App) StartDeploy(req request.StartDeploy) response.Base {
	if req.EnvironmentID == "" {
		return response.Fail("environmentId 不能为空")
	}
	if err := a.deployService.Start(req.EnvironmentID, req.JarIDs); err != nil {
		return response.Fail(err.Error())
	}
	return response.OK()
}

func (a *App) CancelDeploy() response.Base {
	a.deployService.Cancel()
	return response.OK()
}

func (a *App) GetDeployProgress() response.Data[*entity.DeployProgress] {
	return response.OKData(a.deployService.GetProgress())
}

func (a *App) GetDeployHistory() response.Data[[]entity.DeployHistory] {
	return response.OKData(a.history.GetList())
}

func (a *App) GetGlobalSettings() response.Data[entity.GlobalSettings] {
	return response.OKData(a.configService.GetSettings())
}

func (a *App) SaveGlobalSettings(req request.SaveGlobalSettings) response.Base {
	if err := a.configService.SaveSettings(req.Settings); err != nil {
		return response.Fail(err.Error())
	}
	return response.OK()
}

func (a *App) GetSystemDefaults() response.Data[entity.SystemDefaultConfig] {
	return response.OKData(a.configService.GetSystemDefaults())
}

func (a *App) SaveSystemDefaults(req request.SaveSystemDefaults) response.Base {
	if err := a.configService.SaveSystemDefaults(req.Defaults); err != nil {
		return response.Fail(err.Error())
	}
	return response.OK()
}

func (a *App) ParseMavenCommand(req request.ParseMavenCommand) response.Data[*service.MavenParseResult] {
	return response.OKData(service.ParseMavenCommand(req.Command))
}

func (a *App) GetAutoDetectJDK() response.Data[[]map[string]string] {
	return response.OKData(service.AutoDetectJDK())
}

func (a *App) StartJDKDetection() response.Base {
	go func() {
		jdks := service.DetectJDK()
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "jdk-detection-result", jdks)
		}
	}()
	return response.OK()
}

func (a *App) ExportConfig(req request.ExportEnvironment) response.Data[string] {
	if req.ID == "" {
		return response.FailData("id 不能为空", "")
	}
	data, err := a.configService.ExportEnvironment(req.ID)
	if err != nil {
		return response.FailData(err.Error(), "")
	}
	return response.OKData(data)
}

func (a *App) ImportConfig(req request.ImportEnvironment) response.Base {
	if req.JSON == "" {
		return response.Fail("json 不能为空")
	}
	if err := a.configService.ImportEnvironment(req.JSON); err != nil {
		return response.Fail(err.Error())
	}
	return response.OK()
}

