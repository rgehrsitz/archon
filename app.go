package main

import (
	"context"
	"sync"

	"github.com/rgehrsitz/archon/model"
	"github.com/rgehrsitz/archon/plugin"
	"github.com/rgehrsitz/archon/snapshot"
	"github.com/rgehrsitz/archon/storage"
)

// App struct holds the application state and dependencies
type App struct {
	ctx context.Context
	mu  sync.RWMutex

	// Core services
	configVault *storage.ConfigVault
	snapshotMgr *snapshot.Manager
	pluginMgr   *plugin.PluginManager

	// State
	currentProject string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.configVault = storage.NewConfigVault()
	a.snapshotMgr = snapshot.NewManager(a.configVault)
	a.pluginMgr = plugin.NewPluginManager()
}

// LoadProject loads a project from the given path
func (a *App) LoadProject(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.configVault.Load(path); err != nil {
		return err
	}
	a.currentProject = path
	return nil
}

// GetComponentTree returns the current component tree
func (a *App) GetComponentTree() (*model.ComponentTree, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.configVault.GetComponentTree()
}

// CreateSnapshot creates a new snapshot of the current state
func (a *App) CreateSnapshot(message string) (*snapshot.Snapshot, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.snapshotMgr.Create(message)
}

// GetSnapshots returns all snapshots for the current project
func (a *App) GetSnapshots() ([]*snapshot.Snapshot, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.snapshotMgr.List()
}

// LoadPlugin loads a WASM plugin
func (a *App) LoadPlugin(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.pluginMgr.LoadPlugin(path)
}

// ExecutePlugin runs a loaded plugin with the given parameters
func (a *App) ExecutePlugin(pluginID string, params map[string]interface{}) (interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.pluginMgr.Execute(pluginID, params)
}
