package handlers

import (
	"database/sql"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pranav8764/ParkIntel/backend/db"
	"github.com/pranav8764/ParkIntel/backend/inference"
)

type ZoneStats struct {
	TotalHistoricalViolations int      `json:"total_historical_violations"`
	RepeatHotspotScore        float64  `json:"repeat_hotspot_score"`
	TopViolationTypes         []string `json:"top_violation_types"`
	TopVehicleTypes           []string `json:"top_vehicle_types"`
}

type ZoneInsightsResponse struct {
	ZoneID                       string             `json:"zone_id"`
	PredictedHotspotRisk         string             `json:"predicted_hotspot_risk"`
	ModelConfidence              string             `json:"model_confidence"`
	ClassProbabilities           map[string]float64 `json:"class_probabilities"`
	ParkingCongestionImpactScore float64            `json:"parking_congestion_impact_score"`
	PriorityScore                float64            `json:"priority_score"`
	PriorityLevel                string             `json:"priority_level"`
	RecommendedAction            string             `json:"recommended_action"`
	Reasons                      []string           `json:"reasons"`
	Note                         string             `json:"note"`
	ZoneStats                    ZoneStats          `json:"zone_stats"`
}

// GetZoneInsights serves GET /api/zones/:zone_id/insights
func GetZoneInsights(c *gin.Context) {
	zoneID := c.Param("zone_id")
	hourStr := c.Query("hour")

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "database unavailable",
			"code":  503,
		})
		return
	}

	var hour int
	if hourStr != "" {
		var err error
		hour, err = strconv.Atoi(hourStr)
		if err != nil || hour < 0 || hour > 23 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid hour: must be 0-23",
				"code":  400,
			})
			return
		}
	} else {
		hour = time.Now().Hour()
	}

	// Look up the feature row from zone_time_features
	row := database.QueryRow(`
		SELECT 
			hour, day_of_week, police_station, junction_name, junction_flag,
			total_violations, wrong_parking_count, no_parking_count, main_road_count,
			double_parking_count, near_crossing_count, near_signal_count, footpath_count,
			heavy_vehicle_count, medium_vehicle_count, light_vehicle_count, two_wheel_count,
			avg_vio_severity, max_vio_severity, avg_veh_weight, violations_last_1h,
			violations_last_3h, violations_last_24h, violations_last_7d, repeat_hotspot_score,
			historical_zone_log_total, zone_hour_hist_mean, zone_dow_hist_mean, avg_confidence
		FROM zone_time_features
		WHERE zone_id = $1 AND hour = $2
	`, zoneID, hour)

	var f inference.FeatureInput
	var policeStation string
	var junctionName string
	var month int

	// Retrieve month from current time
	month = int(time.Now().Month())

	err := row.Scan(
		&f.Hour, &f.DayOfWeek, &policeStation, &junctionName, &f.JunctionFlag,
		&f.TotalViolations, &f.WrongParkingCount, &f.NoParkingCount, &f.MainRoadCount,
		&f.DoubleParkingCount, &f.NearCrossingCount, &f.NearSignalCount, &f.FootpathCount,
		&f.HeavyVehicleCount, &f.MediumVehicleCount, &f.LightVehicleCount, &f.TwoWheelCount,
		&f.AvgVioSeverity, &f.MaxVioSeverity, &f.AvgVehWeight, &f.ViolationsLast1h,
		&f.ViolationsLast3h, &f.ViolationsLast24h, &f.ViolationsLast7d, &f.RepeatHotspotScore,
		&f.HistoricalZoneLogTotal, &f.ZoneHourHistMean, &f.ZoneDowHistMean, &f.AvgConfidence,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Fallback: search for first available hour for this zone
			err = database.QueryRow(`
				SELECT 
					hour, day_of_week, police_station, junction_name, junction_flag,
					total_violations, wrong_parking_count, no_parking_count, main_road_count,
					double_parking_count, near_crossing_count, near_signal_count, footpath_count,
					heavy_vehicle_count, medium_vehicle_count, light_vehicle_count, two_wheel_count,
					avg_vio_severity, max_vio_severity, avg_veh_weight, violations_last_1h,
					violations_last_3h, violations_last_24h, violations_last_7d, repeat_hotspot_score,
					historical_zone_log_total, zone_hour_hist_mean, zone_dow_hist_mean, avg_confidence
				FROM zone_time_features
				WHERE zone_id = $1
				LIMIT 1
			`, zoneID).Scan(
				&f.Hour, &f.DayOfWeek, &policeStation, &junctionName, &f.JunctionFlag,
				&f.TotalViolations, &f.WrongParkingCount, &f.NoParkingCount, &f.MainRoadCount,
				&f.DoubleParkingCount, &f.NearCrossingCount, &f.NearSignalCount, &f.FootpathCount,
				&f.HeavyVehicleCount, &f.MediumVehicleCount, &f.LightVehicleCount, &f.TwoWheelCount,
				&f.AvgVioSeverity, &f.MaxVioSeverity, &f.AvgVehWeight, &f.ViolationsLast1h,
				&f.ViolationsLast3h, &f.ViolationsLast24h, &f.ViolationsLast7d, &f.RepeatHotspotScore,
				&f.HistoricalZoneLogTotal, &f.ZoneHourHistMean, &f.ZoneDowHistMean, &f.AvgConfidence,
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

	f.Month = float64(month)
	if f.DayOfWeek == 5 || f.DayOfWeek == 6 {
		f.IsWeekend = 1
	} else {
		f.IsWeekend = 0
	}

	// Look up active ONNX sessions singleton
	sessions := modelSessions
	if sessions == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "inference engine unavailable",
			"code":  500,
		})
		return
	}

	// Label encode the police station string
	f.PoliceStationEnc = float64(inference.EncodePoliceStation(sessions.Meta, policeStation))

	// Run ONNX inference
	res, err := inference.RunInference(sessions, f)
	if err != nil {
		log.Printf("Inference error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "inference failed",
			"code":  500,
		})
		return
	}

	// Recompute scores using scoring layer formulas
	impactScore := inference.ImpactScore(f)
	priorityScore := inference.PriorityScore(impactScore, float64(res.HighProb), f)
	priorityLevel := inference.PriorityLevel(priorityScore)
	recAction := inference.RecommendedAction(priorityLevel)
	confidence := inference.ModelConfidence(res.Probabilities)
	reasons := inference.Reasons(f)

	// Calculate historical violations from log value
	histViolations := int(math.Round(math.Exp(f.HistoricalZoneLogTotal) - 1))
	if histViolations < 0 {
		histViolations = 0
	}

	// Top violation types based on count
	type typeVal struct {
		name string
		val  float64
	}
	vios := []typeVal{
		{"WRONG PARKING", f.WrongParkingCount},
		{"NO PARKING", f.NoParkingCount},
		{"PARKING IN A MAIN ROAD", f.MainRoadCount},
		{"DOUBLE PARKING", f.DoubleParkingCount},
		{"PARKING NEAR ROAD CROSSING", f.NearCrossingCount},
		{"PARKING NEAR TRAFFIC LIGHT/ZEBRA CROSSING", f.NearSignalCount},
		{"PARKING ON FOOTPATH", f.FootpathCount},
	}
	sort.Slice(vios, func(i, j int) bool { return vios[i].val > vios[j].val })
	var topVios []string
	for _, v := range vios {
		if v.val > 0 {
			topVios = append(topVios, v.name)
		}
	}
	if len(topVios) == 0 {
		topVios = []string{"WRONG PARKING", "NO PARKING"}
	}

	// Top vehicle types based on count
	vehs := []typeVal{
		{"HEAVY", f.HeavyVehicleCount},
		{"MEDIUM", f.MediumVehicleCount},
		{"LIGHT", f.LightVehicleCount},
		{"TWO_WHEEL", f.TwoWheelCount},
	}
	sort.Slice(vehs, func(i, j int) bool { return vehs[i].val > vehs[j].val })
	var topVehs []string
	for _, v := range vehs {
		if v.val > 0 {
			topVehs = append(topVehs, v.name)
		}
	}
	if len(topVehs) == 0 {
		topVehs = []string{"LIGHT", "TWO_WHEEL", "HEAVY"}
	}

	response := ZoneInsightsResponse{
		ZoneID:               zoneID,
		PredictedHotspotRisk: res.PredictedClass,
		ModelConfidence:      confidence,
		ClassProbabilities: map[string]float64{
			"LOW":    math.Round(float64(res.Probabilities[0])*10000) / 10000,
			"MEDIUM": math.Round(float64(res.Probabilities[1])*10000) / 10000,
			"HIGH":   math.Round(float64(res.Probabilities[2])*10000) / 10000,
		},
		ParkingCongestionImpactScore: impactScore,
		PriorityScore:                priorityScore,
		PriorityLevel:                priorityLevel,
		RecommendedAction:            recAction,
		Reasons:                      reasons,
		Note:                         "Impact score is a proxy based on violation severity and hotspot recurrence, not measured traffic congestion.",
		ZoneStats: ZoneStats{
			TotalHistoricalViolations: histViolations,
			RepeatHotspotScore:        f.RepeatHotspotScore,
			TopViolationTypes:         topVios,
			TopVehicleTypes:           topVehs,
		},
	}

	c.JSON(http.StatusOK, response)
}
