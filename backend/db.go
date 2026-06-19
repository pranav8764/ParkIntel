package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func IngestCSV(db *gorm.DB, filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) < 2 {
		return fmt.Errorf("CSV is empty or only has header")
	}

	var batch []ZonePrediction
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 18 {
			continue // skip malformed row
		}

		lat, _ := strconv.ParseFloat(row[1], 64)
		lon, _ := strconv.ParseFloat(row[2], 64)
		hour, _ := strconv.Atoi(row[5])
		dayOfWeek, _ := strconv.Atoi(row[6])
		month, _ := strconv.Atoi(row[7])

		hourBinStr := strings.TrimSpace(row[8])
		hourBin, _ := time.Parse("2006-01-02 15:04:05-07:00", hourBinStr)

		v1h, _ := strconv.Atoi(row[9])
		v24h, _ := strconv.Atoi(row[10])
		v7d, _ := strconv.Atoi(row[11])
		highProb, _ := strconv.ParseFloat(row[13], 64)
		priorityScore, _ := strconv.ParseFloat(row[14], 64)
		
		reasonsStr := row[17]
		reasonsStr = strings.ReplaceAll(reasonsStr, "'", "\"")

		pred := ZonePrediction{
			ZoneID:               row[0],
			Latitude:             lat,
			Longitude:            lon,
			PoliceStation:        row[3],
			JunctionName:         row[4],
			Hour:                 hour,
			DayOfWeek:            dayOfWeek,
			Month:                month,
			HourBin:              hourBin,
			ViolationsLast1H:     v1h,
			ViolationsLast24H:    v24h,
			ViolationsLast7D:     v7d,
			PredictedHotspotRisk: row[12],
			HighProb:             highProb,
			PriorityScore:        priorityScore,
			PriorityLevel:        row[15],
			RecommendedAction:    row[16],
			ReasonsJSON:          reasonsStr,
		}
		batch = append(batch, pred)
	}

	// Postgres can be sensitive to large batch inserts, using a safer batch size
	err = db.CreateInBatches(batch, 500).Error
	if err != nil {
		return err
	}

	fmt.Printf("Successfully ingested %d records.\n", len(batch))
	return nil
}

func SetupDB() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Println("Warning: DATABASE_URL not set. Please provide a Supabase Postgres connection string.")
		os.Exit(1)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&ZonePrediction{})

	var count int64
	db.Model(&ZonePrediction{}).Count(&count)
	if count == 0 {
		fmt.Println("Database is empty. Starting CSV ingestion...")
		// Assuming script is run from backend directory
		err := IngestCSV(db, "../model-r2 (2)/predictions.csv")
		if err != nil {
			fmt.Println("Error ingesting CSV:", err)
		}
	} else {
		fmt.Printf("Database already contains %d records. Skipping ingestion.\n", count)
	}

	return db
}
