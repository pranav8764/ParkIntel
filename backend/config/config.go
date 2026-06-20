package config

import (
	"os"

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
	onnxLib := os.Getenv("ONNX_RUNTIME_LIB_PATH")
	if onnxLib == "" {
		// Default to the path of libonnxruntime.so found in the python virtual environment
		onnxLib = "./ml-python/.venv/lib/python3.12/site-packages/onnxruntime/capi/libonnxruntime.so.1.27.0"
	}

	return &Config{
		DatabaseURL:        dbURL,
		OnnxModelDir:       modelDir,
		Port:               port,
		GinMode:            ginMode,
		OnnxRuntimeLibPath: onnxLib,
	}, nil
}
