package main

import (
	"context"
	"testing"
)

func TestNewApp(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Error("NewApp() returned nil")
	}
	if app.configVault != nil {
		t.Error("configVault should be nil before startup")
	}
	if app.snapshotMgr != nil {
		t.Error("snapshotMgr should be nil before startup")
	}
	if app.pluginMgr != nil {
		t.Error("pluginMgr should be nil before startup")
	}
}

func TestAppStartup(t *testing.T) {
	app := NewApp()
	ctx := context.Background()
	app.Startup(ctx) // Corrected to use exported Startup method

	if app.ctx != ctx {
		t.Error("Context not properly set during startup")
	}
	if app.configVault == nil {
		t.Error("configVault not initialized during startup")
	}
	if app.snapshotMgr == nil {
		t.Error("snapshotMgr not initialized during startup")
	}
	if app.pluginMgr == nil {
		t.Error("pluginMgr not initialized during startup")
	}
}

func TestLoadProject(t *testing.T) {
	app := NewApp()
	ctx := context.Background()
	app.Startup(ctx) // Corrected to use exported Startup method

	// Test loading non-existent project
	err := app.LoadProject("nonexistent")
	if err == nil {
		t.Error("Expected error when loading non-existent project")
	}

	// TODO: Add test for successful project load once storage package is implemented
}

func TestGetComponentTree(t *testing.T) {
	app := NewApp()
	ctx := context.Background()
	app.Startup(ctx) // Corrected to use exported Startup method
	// Test getting component tree with empty project (should succeed)
	tree, err := app.GetComponentTree()
	if err != nil {
		t.Errorf("Unexpected error when getting component tree: %v", err)
	}
	if tree == nil {
		t.Error("Expected non-nil tree even for empty project")
	}

	// TODO: Add test for successful component tree retrieval once storage package is implemented
}

func TestCreateSnapshot(t *testing.T) {
	app := NewApp()
	ctx := context.Background()
	app.Startup(ctx) // Corrected to use exported Startup method
	// Test creating snapshot with initialized project (should succeed)
	snap, err := app.CreateSnapshot("test snapshot")
	if err != nil {
		t.Errorf("Unexpected error when creating snapshot: %v", err)
	}
	if snap == nil {
		t.Error("Expected non-nil snapshot")
	}

	// TODO: Add test for successful snapshot creation once snapshot package is implemented
}

func TestGetSnapshots(t *testing.T) {
	app := NewApp()
	ctx := context.Background()
	app.Startup(ctx) // Corrected to use exported Startup method
	// Test getting snapshots with initialized project (should succeed)
	snapshots, err := app.GetSnapshots()
	if err != nil {
		t.Errorf("Unexpected error when getting snapshots: %v", err)
	}
	if snapshots == nil {
		t.Error("Expected non-nil snapshots slice even if empty")
	}

	// TODO: Add test for successful snapshot retrieval once snapshot package is implemented
}

func TestPluginOperations(t *testing.T) {
	app := NewApp()
	ctx := context.Background()
	app.Startup(ctx) // Corrected to use exported Startup method

	// Test loading non-existent plugin
	err := app.LoadPlugin("nonexistent.wasm")
	if err == nil {
		t.Error("Expected error when loading non-existent plugin")
	}

	// Test executing non-existent plugin
	result, err := app.ExecutePlugin("nonexistent", nil)
	if err == nil {
		t.Error("Expected error when executing non-existent plugin")
	}
	if result != nil {
		t.Error("Expected nil result when executing non-existent plugin")
	}

	// TODO: Add test for successful plugin operations once plugin package is implemented
}
