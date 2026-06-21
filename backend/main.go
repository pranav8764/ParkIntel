package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pranav8764/ParkIntel/backend/config"
	"github.com/pranav8764/ParkIntel/backend/db"
	"github.com/pranav8764/ParkIntel/backend/handlers"
	"github.com/pranav8764/ParkIntel/backend/inference"
	"github.com/pranav8764/ParkIntel/backend/middleware"
)

func main() {
	log.Println("Starting ParkIntel Go Inference Service...")

	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// 2. Connect to PostgreSQL database
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	err = db.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	// Locate schema.sql file
	schemaPath := "schema.sql"
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		schemaPath = "backend/schema.sql"
	}
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		schemaPath = "../backend/schema.sql"
	}

	// 3. Set up Database Schema
	err = db.SetupSchema(schemaPath)
	if err != nil {
		log.Fatalf("Failed to set up database schema: %v", err)
	}

	// 4. Initialize ONNX runtime and sessions once at startup (blocking)
	if cfg.OnnxRuntimeLibPath == "" {
		log.Fatal("ONNX_RUNTIME_LIB_PATH environment variable is not set")
	}

	modelDir := cfg.OnnxModelDir
	// Fallback logic to locate the model directory automatically if the specified one doesn't exist/contain metadata
	metaFile := filepath.Join(modelDir, "onnx_meta.json")
	if _, err := os.Stat(metaFile); os.IsNotExist(err) {
		log.Printf("Specified model directory '%s' not found or missing onnx_meta.json. Checking fallbacks...\n", modelDir)
		fallbacks := []string{
			"models/onnx",
			"../models/onnx",
			"/app/models/onnx",
			"./models/onnx",
		}
		for _, fb := range fallbacks {
			if _, err := os.Stat(filepath.Join(fb, "onnx_meta.json")); err == nil {
				log.Printf("Found model directory at fallback path: %s\n", fb)
				modelDir = fb
				break
			}
		}
	}

	sessions, err := inference.InitModelSessions(cfg.OnnxRuntimeLibPath, modelDir)
	if err != nil {
		log.Fatalf("Failed to initialize ONNX sessions: %v", err)
	}
	defer sessions.Destroy()

	// Initialize handlers with the ONNX sessions singleton
	handlers.Init(sessions)

	// Locate CSV files for ingestion
	trainCSV := "../ml-python/train.csv"
	testCSV := "../ml-python/test.csv"
	predictionsCSV := "../ml-python/predictions.csv"

	if _, err := os.Stat(trainCSV); os.IsNotExist(err) {
		trainCSV = "ml-python/train.csv"
		testCSV = "ml-python/test.csv"
		predictionsCSV = "ml-python/predictions.csv"
	}

	// 5. Ingest data from CSVs if database tables are empty
	err = db.IngestData(trainCSV, testCSV, predictionsCSV)
	if err != nil {
		log.Printf("Warning: CSV data ingestion skipped or failed: %v", err)
	}

	// 6. Set up Gin router
	r := gin.New()

	// Register explicit middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// Register routes
	r.GET("/api/hotspots", handlers.GetHotspots)
	r.GET("/api/enforcement/ranking", handlers.GetRanking)
	r.GET("/api/zones/:zone_id/insights", handlers.GetZoneInsights)
	r.POST("/api/simulate", handlers.PostSimulate)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 7. Run Server with graceful shutdown support
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Listening and serving HTTP on :%s\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting gracefully")
}
