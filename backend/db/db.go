package db

import (
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

// IngestData performs data ingestion if the tables are empty
func IngestData(trainCSVPath, testCSVPath, predictionsCSVPath string) error {
	var count int
	err := dbPool.QueryRow("SELECT COUNT(*) FROM zone_predictions").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check zone_predictions count: %w", err)
	}

	if count > 0 {
		log.Printf("Database already contains %d records. Skipping ingestion.\n", count)
		return nil
	}

	log.Println("Database is empty. Starting CSV ingestion...")

	// 1. Load train.csv to calculate historical baselines: zone_hour_hist_mean and zone_dow_hist_mean
	log.Println("Step 1/3: Reading train.csv to compute historical means...")
	hourHist, dowHist, err := computeHistoricalMeans(trainCSVPath)
	if err != nil {
		return fmt.Errorf("failed to compute historical means: %w", err)
	}

	// 2. Ingest zone_time_features from test.csv
	log.Println("Step 2/3: Ingesting zone_time_features from test.csv...")
	err = ingestTimeFeatures(testCSVPath, hourHist, dowHist)
	if err != nil {
		return fmt.Errorf("failed to ingest time features: %w", err)
	}

	// 3. Ingest zone_predictions from predictions.csv
	log.Println("Step 3/3: Ingesting zone_predictions from predictions.csv...")
	err = ingestPredictions(predictionsCSVPath)
	if err != nil {
		return fmt.Errorf("failed to ingest predictions: %w", err)
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

func ingestTimeFeatures(testCSVPath string, hourHist, dowHist map[histKey]float64) error {
	file, err := os.Open(testCSVPath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return err
	}

	colIdx := make(map[string]int)
	for i, name := range header {
		colIdx[name] = i
	}

	tx, err := dbPool.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`
		INSERT INTO zone_time_features (
			zone_id, hour, day_of_week, police_station, junction_name, junction_flag,
			total_violations, wrong_parking_count, no_parking_count, main_road_count,
			double_parking_count, near_crossing_count, near_signal_count, footpath_count,
			heavy_vehicle_count, medium_vehicle_count, light_vehicle_count, two_wheel_count,
			avg_vio_severity, max_vio_severity, avg_veh_weight, violations_last_1h,
			violations_last_3h, violations_last_24h, violations_last_7d, repeat_hotspot_score,
			historical_zone_log_total, zone_hour_hist_mean, zone_dow_hist_mean, avg_confidence
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30
		) ON CONFLICT (zone_id, hour) DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	parseFloat := func(s string) float64 {
		v, _ := strconv.ParseFloat(s, 64)
		return v
	}

	parseInt := func(s string) int {
		v, _ := strconv.Atoi(s)
		return v
	}

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		zoneID := record[colIdx["zone_id"]]
		hour := parseInt(record[colIdx["hour"]])
		dow := parseInt(record[colIdx["day_of_week"]])

		// Look up historical means computed from train set
		hk := histKey{zoneID: zoneID, keyVal: hour}
		dk := histKey{zoneID: zoneID, keyVal: dow}
		hourHistMean := hourHist[hk]
		dowHistMean := dowHist[dk]

		_, err = stmt.Exec(
			zoneID,
			hour,
			dow,
			record[colIdx["police_station"]],
			record[colIdx["junction_name"]],
			parseInt(record[colIdx["junction_flag"]]),
			parseFloat(record[colIdx["total_violations"]]),
			parseFloat(record[colIdx["wrong_parking_count"]]),
			parseFloat(record[colIdx["no_parking_count"]]),
			parseFloat(record[colIdx["main_road_count"]]),
			parseFloat(record[colIdx["double_parking_count"]]),
			parseFloat(record[colIdx["near_crossing_count"]]),
			parseFloat(record[colIdx["near_signal_count"]]),
			parseFloat(record[colIdx["footpath_count"]]),
			parseFloat(record[colIdx["heavy_vehicle_count"]]),
			parseFloat(record[colIdx["medium_vehicle_count"]]),
			parseFloat(record[colIdx["light_vehicle_count"]]),
			parseFloat(record[colIdx["two_wheel_count"]]),
			parseFloat(record[colIdx["avg_vio_severity"]]),
			parseFloat(record[colIdx["max_vio_severity"]]),
			parseFloat(record[colIdx["avg_veh_weight"]]),
			parseFloat(record[colIdx["violations_last_1h"]]),
			parseFloat(record[colIdx["violations_last_3h"]]),
			parseFloat(record[colIdx["violations_last_24h"]]),
			parseFloat(record[colIdx["violations_last_7d"]]),
			parseFloat(record[colIdx["repeat_hotspot_score"]]),
			parseFloat(record[colIdx["historical_zone_log_total"]]),
			hourHistMean,
			dowHistMean,
			parseFloat(record[colIdx["avg_confidence"]]),
		)
		if err != nil {
			return fmt.Errorf("failed to insert time feature: %w", err)
		}
		count++
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Printf("Successfully ingested %d zone_time_features.\n", count)
	return nil
}

func ingestPredictions(predictionsCSVPath string) error {
	file, err := os.Open(predictionsCSVPath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return err
	}

	colIdx := make(map[string]int)
	for i, name := range header {
		colIdx[name] = i
	}

	tx, err := dbPool.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`
		INSERT INTO zone_predictions (
			zone_id, zone_lat, zone_lon, police_station, junction_name,
			prediction_time, hour, day_of_week, month, predicted_hotspot_risk,
			model_confidence, high_prob, prob_low, prob_medium, impact_score,
			priority_score, priority_level, recommended_action, reasons_json
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	parseFloat := func(s string) float64 {
		v, _ := strconv.ParseFloat(s, 64)
		return v
	}

	parseInt := func(s string) int {
		v, _ := strconv.Atoi(s)
		return v
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
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		zoneID := record[colIdx["zone_id"]]
		lat := parseFloat(record[colIdx["zone_lat"]])
		lon := parseFloat(record[colIdx["zone_lon"]])
		station := record[colIdx["police_station"]]
		jName := record[colIdx["junction_name"]]
		hour := parseInt(record[colIdx["hour"]])
		dow := parseInt(record[colIdx["day_of_week"]])
		month := parseInt(record[colIdx["month"]])

		predictionTimeStr := record[colIdx["hour_bin"]]
		predictionTime, err := time.Parse("2006-01-02 15:04:05-07:00", predictionTimeStr)
		if err != nil {
			// Fallback parsing format if needed
			predictionTime, _ = time.Parse(time.RFC3339, predictionTimeStr)
		}

		risk := record[colIdx["hotspot_risk"]]
		highProb := parseFloat(record[colIdx["high_prob"]])
		priorityScore := parseFloat(record[colIdx["priority_score"]])
		priorityLevel := record[colIdx["priority_level"]]
		recAction := record[colIdx["recommended_action"]]
		reasons := cleanReasons(record[colIdx["reasons"]])

		// Estimate confidence and impact score to store precomputed predictions
		confidence := "MEDIUM"
		if highProb > 0.45 {
			confidence = "HIGH"
		} else if highProb < 0.15 {
			confidence = "LOW"
		}

		// Pre-computed predictions impact score can be approximated from priority score during CSV ingestion
		// (Will be computed properly during live inference).
		impactScore := math.Round(((priorityScore - 0.25*highProb*100)/0.75)*100) / 100
		if impactScore < 0 {
			impactScore = priorityScore
		}
		if impactScore > 100 {
			impactScore = 100
		}

		_, err = stmt.Exec(
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
			1.0-highProb, // dummy prob_low
			0.0,          // dummy prob_medium
			impactScore,
			priorityScore,
			priorityLevel,
			recAction,
			reasons,
		)
		if err != nil {
			return fmt.Errorf("failed to insert prediction: %w", err)
		}
		count++
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Printf("Successfully ingested %d zone_predictions.\n", count)
	return nil
}
