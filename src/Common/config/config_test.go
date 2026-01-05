package config

import (
	"encoding/json"
	"os"
	"testing"
)

func TestConfigAIEmbeddingModel(t *testing.T) {
	// Create a temporary config file
	tmpFile := "config_test.json"
	cfg := AppConfig{
		AIEmbeddingModel: "doubao-embedding-vision-251215",
	}
	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(tmpFile, data, 0644)
	defer os.Remove(tmpFile)

	// Initialize config
	err := InitConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to init config: %v", err)
	}

	if GlobalConfig.AIEmbeddingModel != "doubao-embedding-vision-251215" {
		t.Errorf("Expected doubao-embedding-vision-251215, got %s", GlobalConfig.AIEmbeddingModel)
	}
}

func TestConfigFromEnv(t *testing.T) {
	os.Setenv("AI_EMBEDDING_MODEL", "env-model-id")
	defer os.Unsetenv("AI_EMBEDDING_MODEL")

	loadConfigFromEnv()

	if GlobalConfig.AIEmbeddingModel != "env-model-id" {
		t.Errorf("Expected env-model-id, got %s", GlobalConfig.AIEmbeddingModel)
	}
}
