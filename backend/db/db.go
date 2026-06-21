package db

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var dbPool *sql.DB

// InitDB initializes the PostgreSQL connection pool
func InitDB(databaseURL string) error {
	var err error
	dbPool, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool limits
	dbPool.SetMaxOpenConns(25)
	dbPool.SetMaxIdleConns(25)
	dbPool.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err = dbPool.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return nil
}

// GetDB returns the database connection pool
func GetDB() *sql.DB {
	return dbPool
}

// CloseDB closes the database connection pool
func CloseDB() {
	if dbPool != nil {
		_ = dbPool.Close()
	}
}

// SetupSchema executes the schema SQL script
func SetupSchema(schemaPath string) error {
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	_, err = dbPool.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Println("Database schema checked/created successfully")
	return nil
}

// IngestData performs data ingestion if the tables are empty or incomplete
func IngestData(trainCSVPath, testCSVPath, predictionsCSVPath string) error {
	ctx := context.Background()
	conn, err := dbPool.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	defer conn.Close()

	var locked bool
	err = conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock(54321)").Scan(&locked)
	if err != nil {
		return fmt.Errorf("failed to acquire advisory lock: %w", err)
	}
	if !locked {
		log.Println("Another data ingestion process is already running. Skipping ingestion.")
		return nil
	}
	defer func() {
		_, _ = conn.ExecContext(ctx, "SELECT pg_advisory_unlock(54321)")
	}()

	var count int
	err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM zone_predictions").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check zone_predictions count: %w", err)
	}

	forceIngest := os.Getenv("FORCE_INGEST") == "true"
	expectedRows := 11280

	if count > 0 && !forceIngest {
		if count >= expectedRows {
			log.Printf("Database already contains %d records (expected %d). Skipping ingestion.\n", count, expectedRows)
			return nil
		}
		log.Printf("Database contains incomplete data (%d/%d records). Clearing tables to re-ingest complete data...\n", count, expectedRows)
		forceIngest = true
	}

	if forceIngest {
		_, err = conn.ExecContext(ctx, "DELETE FROM zone_predictions")
		if err != nil {
			return fmt.Errorf("failed to clear zone_predictions: %w", err)
		}
		_, err = conn.ExecContext(ctx, "DELETE FROM zone_time_features")
		if err != nil {
			return fmt.Errorf("failed to clear zone_time_features: %w", err)
		}
		log.Println("Database tables cleared successfully.")
	}

	log.Println("Starting CSV ingestion...")

	// 1. Load train.csv to calculate historical baselines: zone_hour_hist_mean and zone_dow_hist_mean
	log.Println("Step 1/2: Reading train.csv to compute historical means...")
	hourHist, dowHist, err := computeHistoricalMeans(trainCSVPath)
	if err != nil {
		return fmt.Errorf("failed to compute historical means: %w", err)
	}

	// 2. Ingest zone_time_features and zone_predictions in synchronized batches of 100
	log.Println("Step 2/2: Ingesting zone_time_features and zone_predictions in synchronized batches of 100...")
	err = ingestSynchronized(testCSVPath, predictionsCSVPath, hourHist, dowHist)
	if err != nil {
		return fmt.Errorf("failed synchronized CSV ingestion: %w", err)
	}

	log.Println("CSV Ingestion complete successfully!")
	return nil
}

type histKey struct {
	zoneID string
	keyVal int // hour or day of week
}

func computeHistoricalMeans(trainCSVPath string) (map[histKey]float64, map[histKey]float64, error) {
	file, err := os.Open(trainCSVPath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Parse header
	header, err := reader.Read()
	if err != nil {
		return nil, nil, err
	}

	// Map column names to indexes
	colIdx := make(map[string]int)
	for i, name := range header {
		colIdx[name] = i
	}

	requiredCols := []string{"zone_id", "hour", "day_of_week", "total_violations"}
	for _, col := range requiredCols {
		if _, ok := colIdx[col]; !ok {
			return nil, nil, fmt.Errorf("required column %s not found in train.csv", col)
		}
	}

	hourSums := make(map[histKey]float64)
	hourCounts := make(map[histKey]int)
	dowSums := make(map[histKey]float64)
	dowCounts := make(map[histKey]int)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}

		zoneID := record[colIdx["zone_id"]]
		hour, _ := strconv.Atoi(record[colIdx["hour"]])
		dow, _ := strconv.Atoi(record[colIdx["day_of_week"]])
		totalViolations, _ := strconv.ParseFloat(record[colIdx["total_violations"]], 64)

		hk := histKey{zoneID: zoneID, keyVal: hour}
		hourSums[hk] += totalViolations
		hourCounts[hk]++

		dk := histKey{zoneID: zoneID, keyVal: dow}
		dowSums[dk] += totalViolations
		dowCounts[dk]++
	}

	hourMeans := make(map[histKey]float64)
	for k, sum := range hourSums {
		hourMeans[k] = sum / float64(hourCounts[k])
	}

	dowMeans := make(map[histKey]float64)
	for k, sum := range dowSums {
		dowMeans[k] = sum / float64(dowCounts[k])
	}

	return hourMeans, dowMeans, nil
}

func ingestSynchronized(testCSVPath, predictionsCSVPath string, hourHist, dowHist map[histKey]float64) error {
	testFile, err := os.Open(testCSVPath)
	if err != nil {
		return err
	}
	defer testFile.Close()

	predFile, err := os.Open(predictionsCSVPath)
	if err != nil {
		return err
	}
	defer predFile.Close()

	testReader := csv.NewReader(testFile)
	testHeader, err := testReader.Read()
	if err != nil {
		return err
	}

	predReader := csv.NewReader(predFile)
	predHeader, err := predReader.Read()
	if err != nil {
		return err
	}

	testColIdx := make(map[string]int)
	for i, name := range testHeader {
		testColIdx[name] = i
	}

	predColIdx := make(map[string]int)
	for i, name := range predHeader {
		predColIdx[name] = i
	}

	// Helper to clean Python string list representation: e.g. ['General violation cluster'] -> ["General violation cluster"]
	cleanReasons := func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			return "[]"
		}
		s = strings.ReplaceAll(s, "'", "\"")
		var arr []string
		if err := json.Unmarshal([]byte(s), &arr); err == nil {
			bytes, _ := json.Marshal(arr)
			return string(bytes)
		}
		return "[]"
	}

	count := 0
	batchSize := 100

	for {
		var testBatch [][]string
		var predBatch [][]string
		var eof bool

		for i := 0; i < batchSize; i++ {
			testRecord, err := testReader.Read()
			if err == io.EOF {
				eof = true
				break
			}
			if err != nil {
				return fmt.Errorf("error reading test.csv: %w", err)
			}

			predRecord, err := predReader.Read()
			if err == io.EOF {
				eof = true
				break
			}
			if err != nil {
				return fmt.Errorf("error reading predictions.csv: %w", err)
			}

			testBatch = append(testBatch, testRecord)
			predBatch = append(predBatch, predRecord)
		}

		if len(testBatch) == 0 {
			break
		}

		tx, err := dbPool.Begin()
		if err != nil {
			return err
		}

		// 1. Ingest all features in this batch
		err = bulkInsertFeatures(tx, testBatch, testColIdx, hourHist, dowHist)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed synchronized CSV ingestion for features: %w", err)
		}

		// 2. Ingest all predictions in this batch
		err = bulkInsertPredictions(tx, predBatch, predColIdx, cleanReasons)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed synchronized CSV ingestion for predictions: %w", err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction batch: %w", err)
		}

		count += len(testBatch)
		log.Printf("   [Progress] Ingested synchronized batch of %d rows into both tables (total: %d)...\n", len(testBatch), count)

		if eof {
			break
		}
	}

	log.Printf("Successfully ingested %d synchronized rows into both tables.\n", count)
	return nil
}

func bulkInsertFeatures(tx *sql.Tx, batch [][]string, testColIdx map[string]int, hourHist, dowHist map[histKey]float64) error {
	if len(batch) == 0 {
		return nil
	}

	numCols := 30
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		INSERT INTO zone_time_features (
			zone_id, hour, day_of_week, police_station, junction_name, junction_flag,
			total_violations, wrong_parking_count, no_parking_count, main_road_count,
			double_parking_count, near_crossing_count, near_signal_count, footpath_count,
			heavy_vehicle_count, medium_vehicle_count, light_vehicle_count, two_wheel_count,
			avg_vio_severity, max_vio_severity, avg_veh_weight, violations_last_1h,
			violations_last_3h, violations_last_24h, violations_last_7d, repeat_hotspot_score,
			historical_zone_log_total, zone_hour_hist_mean, zone_dow_hist_mean, avg_confidence
		) VALUES `)

	args := make([]interface{}, 0, len(batch)*numCols)
	placeholderIdx := 1

	parseFloat := func(s string) float64 {
		v, _ := strconv.ParseFloat(s, 64)
		return v
	}

	parseInt := func(s string) int {
		v, _ := strconv.Atoi(s)
		return v
	}

	for i, record := range batch {
		if i > 0 {
			queryBuilder.WriteString(", ")
		}
		queryBuilder.WriteString("(")
		for j := 0; j < numCols; j++ {
			if j > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(fmt.Sprintf("$%d", placeholderIdx))
			placeholderIdx++
		}
		queryBuilder.WriteString(")")

		zoneID := record[testColIdx["zone_id"]]
		hour := parseInt(record[testColIdx["hour"]])
		dow := parseInt(record[testColIdx["day_of_week"]])

		hk := histKey{zoneID: zoneID, keyVal: hour}
		dk := histKey{zoneID: zoneID, keyVal: dow}
		hourHistMean := hourHist[hk]
		dowHistMean := dowHist[dk]

		args = append(args,
			zoneID,
			hour,
			dow,
			record[testColIdx["police_station"]],
			record[testColIdx["junction_name"]],
			parseInt(record[testColIdx["junction_flag"]]),
			parseFloat(record[testColIdx["total_violations"]]),
			parseFloat(record[testColIdx["wrong_parking_count"]]),
			parseFloat(record[testColIdx["no_parking_count"]]),
			parseFloat(record[testColIdx["main_road_count"]]),
			parseFloat(record[testColIdx["double_parking_count"]]),
			parseFloat(record[testColIdx["near_crossing_count"]]),
			parseFloat(record[testColIdx["near_signal_count"]]),
			parseFloat(record[testColIdx["footpath_count"]]),
			parseFloat(record[testColIdx["heavy_vehicle_count"]]),
			parseFloat(record[testColIdx["medium_vehicle_count"]]),
			parseFloat(record[testColIdx["light_vehicle_count"]]),
			parseFloat(record[testColIdx["two_wheel_count"]]),
			parseFloat(record[testColIdx["avg_vio_severity"]]),
			parseFloat(record[testColIdx["max_vio_severity"]]),
			parseFloat(record[testColIdx["avg_veh_weight"]]),
			parseFloat(record[testColIdx["violations_last_1h"]]),
			parseFloat(record[testColIdx["violations_last_3h"]]),
			parseFloat(record[testColIdx["violations_last_24h"]]),
			parseFloat(record[testColIdx["violations_last_7d"]]),
			parseFloat(record[testColIdx["repeat_hotspot_score"]]),
			parseFloat(record[testColIdx["historical_zone_log_total"]]),
			hourHistMean,
			dowHistMean,
			parseFloat(record[testColIdx["avg_confidence"]]),
		)
	}

	queryBuilder.WriteString(" ON CONFLICT (zone_id, hour) DO NOTHING")

	_, err := tx.Exec(queryBuilder.String(), args...)
	return err
}

func bulkInsertPredictions(tx *sql.Tx, batch [][]string, predColIdx map[string]int, cleanReasons func(string) string) error {
	if len(batch) == 0 {
		return nil
	}

	numCols := 19
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		INSERT INTO zone_predictions (
			zone_id, zone_lat, zone_lon, police_station, junction_name,
			prediction_time, hour, day_of_week, month, predicted_hotspot_risk,
			model_confidence, high_prob, prob_low, prob_medium, impact_score,
			priority_score, priority_level, recommended_action, reasons_json
		) VALUES `)

	args := make([]interface{}, 0, len(batch)*numCols)
	placeholderIdx := 1

	parseFloat := func(s string) float64 {
		v, _ := strconv.ParseFloat(s, 64)
		return v
	}

	parseInt := func(s string) int {
		v, _ := strconv.Atoi(s)
		return v
	}

	for i, record := range batch {
		if i > 0 {
			queryBuilder.WriteString(", ")
		}
		queryBuilder.WriteString("(")
		for j := 0; j < numCols; j++ {
			if j > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(fmt.Sprintf("$%d", placeholderIdx))
			placeholderIdx++
		}
		queryBuilder.WriteString(")")

		zoneID := record[predColIdx["zone_id"]]
		lat := parseFloat(record[predColIdx["zone_lat"]])
		lon := parseFloat(record[predColIdx["zone_lon"]])
		station := record[predColIdx["police_station"]]
		jName := record[predColIdx["junction_name"]]
		hour := parseInt(record[predColIdx["hour"]])
		dow := parseInt(record[predColIdx["day_of_week"]])
		month := parseInt(record[predColIdx["month"]])

		predictionTimeStr := record[predColIdx["hour_bin"]]
		predictionTime, err := time.Parse("2006-01-02 15:04:05-07:00", predictionTimeStr)
		if err != nil {
			predictionTime, _ = time.Parse(time.RFC3339, predictionTimeStr)
		}

		risk := record[predColIdx["hotspot_risk"]]
		highProb := parseFloat(record[predColIdx["high_prob"]])
		priorityScore := parseFloat(record[predColIdx["priority_score"]])
		priorityLevel := record[predColIdx["priority_level"]]
		recAction := record[predColIdx["recommended_action"]]
		reasons := cleanReasons(record[predColIdx["reasons"]])

		confidence := "MEDIUM"
		if highProb > 0.45 {
			confidence = "HIGH"
		} else if highProb < 0.15 {
			confidence = "LOW"
		}

		impactScore := math.Round(((priorityScore - 0.25*highProb*100)/0.75)*100) / 100
		if impactScore < 0 {
			impactScore = priorityScore
		}
		if impactScore > 100 {
			impactScore = 100
		}

		args = append(args,
			zoneID,
			lat,
			lon,
			station,
			jName,
			predictionTime,
			hour,
			dow,
			month,
			risk,
			confidence,
			highProb,
			1.0-highProb,
			0.0,
			impactScore,
			priorityScore,
			priorityLevel,
			recAction,
			reasons,
		)
	}

	_, err := tx.Exec(queryBuilder.String(), args...)
	return err
}
