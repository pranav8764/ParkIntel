package handlers

import (
	"database/sql"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pranav8764/ParkIntel/backend/db"
	"github.com/pranav8764/ParkIntel/backend/inference"
)

type SimulateRequest struct {
	ZoneID                    string  `json:"zone_id"`
	ViolationReductionPercent float64 `json:"violation_reduction_percent"`
}

type SimulateResponse struct {
	ZoneID                   string  `json:"zone_id"`
	ViolationReductionPercent float64 `json:"violation_reduction_percent"`
	CurrentPriorityScore     float64 `json:"current_priority_score"`
	CurrentPriorityLevel     string  `json:"current_priority_level"`
	SimulatedPriorityScore   float64 `json:"simulated_priority_score"`
	SimulatedPriorityLevel   string  `json:"simulated_priority_level"`
	PriorityChange           string  `json:"priority_change"`
	EstimatedImpactReduction float64 `json:"estimated_impact_reduction"`
	Note                     string  `json:"note"`
}

// PostSimulate serves POST /api/simulate
func PostSimulate(c *gin.Context) {
	var req SimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"code":  400,
		})
		return
	}

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "database unavailable",
			"code":  503,
		})
		return
	}

	// 1. Fetch current priority score and risk details from zone_predictions
	var currentPriority float64
	var currentLevel string
	var highProb float64
	var hour int

	err := database.QueryRow(`
		SELECT priority_score, priority_level, high_prob, hour
		FROM zone_predictions
		WHERE zone_id = $1
		ORDER BY prediction_time DESC LIMIT 1
	`, req.ZoneID).Scan(&currentPriority, &currentLevel, &highProb, &hour)

	if err != nil {
		if err == sql.ErrNoRows {
			// Fallback: check if the zone exists in time features at least
			var exists bool
			_ = database.QueryRow("SELECT EXISTS(SELECT 1 FROM zone_time_features WHERE zone_id = $1)", req.ZoneID).Scan(&exists)
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "zone not found",
					"code":  404,
				})
				return
			}
			// If features exist but no pre-computed prediction, use defaults
			currentPriority = 40.0
			currentLevel = "LOW"
			highProb = 0.1
			hour = time.Now().Hour()
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database error",
				"code":  500,
			})
			return
		}
	}

	// 2. Fetch the zone's time features to perform violation reduction scaling
	var f inference.FeatureInput
	err = database.QueryRow(`
		SELECT 
			junction_flag, wrong_parking_count, no_parking_count, main_road_count,
			double_parking_count, near_crossing_count, near_signal_count, footpath_count,
			heavy_vehicle_count, repeat_hotspot_score
		FROM zone_time_features
		WHERE zone_id = $1 AND hour = $2
	`, req.ZoneID, hour).Scan(
		&f.JunctionFlag, &f.WrongParkingCount, &f.NoParkingCount, &f.MainRoadCount,
		&f.DoubleParkingCount, &f.NearCrossingCount, &f.NearSignalCount, &f.FootpathCount,
		&f.HeavyVehicleCount, &f.RepeatHotspotScore,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Fallback: search for first available hour's features
			err = database.QueryRow(`
				SELECT 
					junction_flag, wrong_parking_count, no_parking_count, main_road_count,
					double_parking_count, near_crossing_count, near_signal_count, footpath_count,
					heavy_vehicle_count, repeat_hotspot_score
				FROM zone_time_features
				WHERE zone_id = $1
				LIMIT 1
			`, req.ZoneID).Scan(
				&f.JunctionFlag, &f.WrongParkingCount, &f.NoParkingCount, &f.MainRoadCount,
				&f.DoubleParkingCount, &f.NearCrossingCount, &f.NearSignalCount, &f.FootpathCount,
				&f.HeavyVehicleCount, &f.RepeatHotspotScore,
			)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "zone not found",
					"code":  404,
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database error",
				"code":  500,
			})
			return
		}
	}

	// Recompute original impact score first
	originalImpact := inference.ImpactScore(f)

	// Scale violation counts by (1 - reduction/100)
	factor := 1.0 - (req.ViolationReductionPercent / 100.0)
	f.DoubleParkingCount *= factor
	f.MainRoadCount *= factor
	f.NearCrossingCount *= factor
	f.NearSignalCount *= factor
	f.NoParkingCount *= factor
	f.WrongParkingCount *= factor
	f.HeavyVehicleCount *= factor
	f.FootpathCount *= factor

	// Recompute impact score and priority score with scaled features
	simulatedImpact := inference.ImpactScore(f)
	simulatedPriority := inference.PriorityScore(simulatedImpact, highProb, f)
	simulatedLevel := inference.PriorityLevel(simulatedPriority)

	estimatedImpactReduction := originalImpact - simulatedImpact
	if estimatedImpactReduction < 0 {
		estimatedImpactReduction = 0
	}

	response := SimulateResponse{
		ZoneID:                    req.ZoneID,
		ViolationReductionPercent: req.ViolationReductionPercent,
		CurrentPriorityScore:      currentPriority,
		CurrentPriorityLevel:      currentLevel,
		SimulatedPriorityScore:    simulatedPriority,
		SimulatedPriorityLevel:    simulatedLevel,
		PriorityChange:            currentLevel + " → " + simulatedLevel,
		EstimatedImpactReduction:  math.Round(estimatedImpactReduction*100) / 100,
		Note:                     "This is a priority-score simulation, not actual traffic-speed simulation.",
	}

	c.JSON(http.StatusOK, response)
}
