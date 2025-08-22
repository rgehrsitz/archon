package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"wailts/internal/api"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()
	// Instantiate Wails-bound services (stubs for now)
	projectService := api.NewProjectService()
	nodeService := api.NewNodeService()
	snapshotService := api.NewSnapshotService()
	diffService := api.NewDiffService()
	mergeService := api.NewMergeService()
	gitService := api.NewGitService()
	importService := api.NewImportService()
	indexService := api.NewIndexService()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Archon",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// BackgroundColour: &options.RGBA{R: 128, G: 38, B: 54, A: 1},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
			projectService,
			nodeService,
			snapshotService,
			diffService,
			mergeService,
			gitService,
			importService,
			indexService,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
