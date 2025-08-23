package main

import (
	"context"

	"github.com/rgehrsitz/archon/internal/api"
)

// App struct contains the application services
type App struct {
	ctx            context.Context
	projectService *api.ProjectService
	nodeService    *api.NodeService
}

// NewApp creates a new App application struct
func NewApp() *App {
	projectService := api.NewProjectService()
	nodeService := api.NewNodeService(projectService)
	
	return &App{
		projectService: projectService,
		nodeService:    nodeService,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetProjectService returns the project service for Wails binding
func (a *App) GetProjectService() *api.ProjectService {
	return a.projectService
}

// GetNodeService returns the node service for Wails binding
func (a *App) GetNodeService() *api.NodeService {
	return a.nodeService
}
