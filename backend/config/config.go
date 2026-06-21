package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/joho/godotenv"
)

// Config holds all configuration parameters for the application
type Config struct {
	DatabaseURL        string
	OnnxModelDir       string
	Port               string
	GinMode            string
	OnnxRuntimeLibPath string
}

// LoadConfig loads the configuration from the environment and .env file
func LoadConfig() (*Config, error) {
	// Attempt to load .env file, ignore error if it doesn't exist
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	modelDir := os.Getenv("ONNX_MODEL_DIR")
	if modelDir == "" {
		modelDir = "./models/onnx"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "release"
	}
	onnxLib := resolveOnnxRuntimeLibPath(os.Getenv("ONNX_RUNTIME_LIB_PATH"))
	if _, err := os.Stat(onnxLib); err != nil {
		return nil, fmt.Errorf("ONNX runtime shared library not found at %q; set ONNX_RUNTIME_LIB_PATH to the absolute libonnxruntime.so path: %w", onnxLib, err)
	}

	return &Config{
		DatabaseURL:        dbURL,
		OnnxModelDir:       modelDir,
		Port:               port,
		GinMode:            ginMode,
		OnnxRuntimeLibPath: onnxLib,
	}, nil
}

func resolveOnnxRuntimeLibPath(configured string) string {
	if configured != "" {
		return configured
	}

	for _, candidate := range []string{
		"/usr/lib/libonnxruntime.so",
		"/usr/local/lib/libonnxruntime.so",
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	for _, pattern := range []string{
		"../ml-python/.venv/lib/python*/site-packages/onnxruntime/capi/libonnxruntime.so.*",
		"./ml-python/.venv/lib/python*/site-packages/onnxruntime/capi/libonnxruntime.so.*",
	} {
		matches, _ := filepath.Glob(pattern)
		if len(matches) == 0 {
			continue
		}
		sort.Strings(matches)
		return matches[len(matches)-1]
	}

	return "/usr/lib/libonnxruntime.so"
}
