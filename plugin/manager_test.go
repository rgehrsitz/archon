package plugin

import (
	"os"
	"testing"
)

func TestNewPluginManager(t *testing.T) {
	pm := NewPluginManager()
	if pm == nil {
		t.Error("NewPluginManager() returned nil")
	}
	if pm.plugins == nil {
		t.Error("plugins map not initialized")
	}
	if len(pm.plugins) != 0 {
		t.Error("plugins map should be empty")
	}
}

func TestLoadPlugin(t *testing.T) {
	pm := NewPluginManager()

	// Test loading non-existent plugin
	err := pm.LoadPlugin("nonexistent.wasm")
	if err == nil {
		t.Error("Expected error when loading non-existent plugin")
	}

	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test*.wasm")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Test loading existing plugin
	err = pm.LoadPlugin(tmpFile.Name())
	if err != nil {
		t.Errorf("LoadPlugin() error = %v", err)
	}

	// Verify plugin was loaded
	pm.mu.RLock()
	_, exists := pm.plugins[tmpFile.Name()]
	pm.mu.RUnlock()
	if !exists {
		t.Error("Plugin not found after loading")
	}

	// Test loading duplicate plugin (should succeed)
	err = pm.LoadPlugin(tmpFile.Name())
	if err != nil {
		t.Errorf("LoadPlugin() duplicate error = %v", err)
	}
}

func TestExecute(t *testing.T) {
	pm := NewPluginManager()

	// Test executing non-existent plugin
	_, err := pm.Execute("nonexistent", nil)
	if err == nil {
		t.Error("Expected error when executing non-existent plugin")
	}

	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test*.wasm")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Test executing existing plugin
	err = pm.LoadPlugin(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadPlugin() error = %v", err)
	}

	params := map[string]interface{}{
		"key": "value",
	}
	result, err := pm.Execute(tmpFile.Name(), params)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	if result == nil {
		t.Error("Execute() returned nil result")
	}

	// Verify result matches input params (current implementation)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Error("Execute() result is not a map")
	}
	if resultMap["key"] != "value" {
		t.Error("Execute() result does not match input params")
	}
}
