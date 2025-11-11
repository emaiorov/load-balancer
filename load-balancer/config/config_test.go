package config

import (
	"os"
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("success", func(t *testing.T) {

		testConfigJSON := `
{
    "app": {"algorythm": "test_alg", "port": ":9090", "health_check_seconds": 5},
    "servers": [{"url": "http://test.com", "health": "/", "weight": 100}]
}`
		// Use t.TempDir() to create a temporary directory for test file.
		tempDir := t.TempDir()
		tempFile, err := os.CreateTemp(tempDir, "config-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		_, err = tempFile.Write([]byte(testConfigJSON))
		if err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tempFile.Close()

		// Expected output struct
		expectedConfig := &Config{
			App: struct {
				Handler            string `json:"algorythm"`
				Port               string `json:"port"`
				HealthCheckSeconds int    `json:"health_check_seconds"`
			}{
				Handler:            "test_alg",
				Port:               ":9090",
				HealthCheckSeconds: 5,
			},
			Servers: []ServerConfig{
				{
					Url:    "http://test.com",
					Health: "/",
					Weight: 100,
				},
			},
		}

		loadedConfig, err := Load(tempFile.Name())

		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		// Compare results
		if !reflect.DeepEqual(loadedConfig, expectedConfig) {
			t.Errorf("Loaded config does not match expected config.")
			t.Logf("Expected: %+v\n", expectedConfig)
			t.Logf("Got:      %+v\n", loadedConfig)
		}
	})

	t.Run("file_not_found", func(t *testing.T) {
		config, err := Load("fake_path.json")

		if err == nil {
			t.Error("Expected an error for missing file, but got nil")
		}

		if config != nil {
			t.Error("Expected config to be nil, but it was not")
		}
	})

	t.Run("broken_json", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFile, err := os.CreateTemp(tempDir, "bad-config-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		_, err = tempFile.Write([]byte(`{"app": {"algorythm": "test_alg", "port": ":9090"}`))
		if err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tempFile.Close()

		loadedConfig, err := Load(tempFile.Name())

		if err == nil {
			t.Error("Expected an error for bad JSON, but got nil")
		}

		if loadedConfig != nil {
			t.Error("Expected loaded config to be nil, but it does not")
		}
	})
}
