package main

import (
	"context"

	"github.com/rgehrsitz/archon/internal/api"
	"github.com/rgehrsitz/archon/internal/logging"
)

// App struct contains the application services
type App struct {
	ctx              context.Context
	projectService   *api.ProjectService
	nodeService      *api.NodeService
	loggingService   *api.LoggingService
	migrationService *api.MigrationService
	gitService       *api.GitService
	snapshotService  *api.SnapshotService
}

// NewApp creates a new App application struct
func NewApp() *App {
	projectService := api.NewProjectService()
	nodeService := api.NewNodeService(projectService)
	loggingService := api.NewLoggingService(projectService)
	migrationService := api.NewMigrationService()
	gitService := api.NewGitService(projectService)
	snapshotService := api.NewSnapshotService(projectService)
	
	return &App{
		projectService:   projectService,
		nodeService:      nodeService,
		loggingService:   loggingService,
		migrationService: migrationService,
		gitService:       gitService,
		snapshotService:  snapshotService,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called when the app is quitting.
func (a *App) shutdown(ctx context.Context) {
    // Graceful logging shutdown
    logging.Shutdown()
}

// GetProjectService returns the project service for Wails binding
func (a *App) GetProjectService() *api.ProjectService {
	return a.projectService
}

// GetNodeService returns the node service for Wails binding
func (a *App) GetNodeService() *api.NodeService {
	return a.nodeService
}

// GetLoggingService returns the logging service for Wails binding
func (a *App) GetLoggingService() *api.LoggingService {
	return a.loggingService
}

// GetMigrationService returns the migration service for Wails binding
func (a *App) GetMigrationService() *api.MigrationService {
	return a.migrationService
}

// GetGitService returns the git service for Wails binding
func (a *App) GetGitService() *api.GitService {
	return a.gitService
}

// GetSnapshotService returns the snapshot service for Wails binding
func (a *App) GetSnapshotService() *api.SnapshotService {
	return a.snapshotService
}
