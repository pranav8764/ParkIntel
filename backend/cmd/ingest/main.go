package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pranav8764/ParkIntel/backend/db"
)

func main() {
	log.Println("Starting independent data ingestion script...")

	// Load configuration from .env file
	_ = godotenv.Load()

	// Also try loading .env from parent directories if run from backend/cmd/ingest/
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load("../.env")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// 1. Initialize PostgreSQL database connection pool
	err := db.InitDB(databaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	// 2. Locate schema.sql file
	schemaPath := "schema.sql"
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		schemaPath = "backend/schema.sql"
	}
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		schemaPath = "../backend/schema.sql"
	}
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		schemaPath = "../../schema.sql"
	}

	// 3. Set up Database Schema
	err = db.SetupSchema(schemaPath)
	if err != nil {
		log.Fatalf("Failed to set up database schema: %v", err)
	}

	// 4. Locate CSV files for ingestion
	trainCSV := "../ml-python/train.csv"
	testCSV := "../ml-python/test.csv"
	predictionsCSV := "../ml-python/predictions.csv"

	// Fallback paths depending on current working directory
	if _, err := os.Stat(trainCSV); os.IsNotExist(err) {
		trainCSV = "ml-python/train.csv"
		testCSV = "ml-python/test.csv"
		predictionsCSV = "ml-python/predictions.csv"
	}
	if _, err := os.Stat(trainCSV); os.IsNotExist(err) {
		trainCSV = "../../ml-python/train.csv"
		testCSV = "../../ml-python/test.csv"
		predictionsCSV = "../../ml-python/predictions.csv"
	}

	// 5. Run ingestion
	err = db.IngestData(trainCSV, testCSV, predictionsCSV)
	if err != nil {
		log.Fatalf("CSV data ingestion failed: %v", err)
	}

	log.Println("Data ingestion script completed successfully!")
}
