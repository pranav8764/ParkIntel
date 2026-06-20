package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pranav8764/ParkIntel/backend/cache"
	"github.com/pranav8764/ParkIntel/backend/db"
)

type Hotspot struct {
	ZoneID               string  `json:"zone_id"`
	Lat                  float64 `json:"lat"`
	Lng                  float64 `json:"lng"`
	PoliceStation        string  `json:"police_station"`
	PriorityScore        float64 `json:"priority_score"`
	PriorityLevel        string  `json:"priority_level"`
	ImpactScore          float64 `json:"impact_score"`
	ExpectedViolations   int     `json:"expected_violations"`
	PredictedHotspotRisk string  `json:"predicted_hotspot_risk"`
	ModelConfidence      string  `json:"model_confidence"`
}

type HotspotsResponse struct {
	Hotspots       []Hotspot         `json:"hotspots"`
	Count          int               `json:"count"`
	FiltersApplied map[string]interface{} `json:"filters_applied"`
}

var hotspotsCache = cache.NewMemoryCache(5 * time.Minute)

// GetHotspots serves GET /api/hotspots
func GetHotspots(c *gin.Context) {
	hourStr := c.Query("hour")
	dateStr := c.Query("date") // Format YYYY-MM-DD
	policeStation := c.Query("police_station")
	riskLevel := c.Query("risk_level")

	if hourStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid hour: must be 0-23",
			"code":  400,
		})
		return
	}

	hour, err := strconv.Atoi(hourStr)
	if err != nil || hour < 0 || hour > 23 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid hour: must be 0-23",
			"code":  400,
		})
		return
	}

	// Read from cache if eligible (we cache by (hour, police_station))
	// Wait, we only use cache if no other params are specified, or we include all filters in cache key
	cacheKey := fmt.Sprintf("%d:%s:%s:%s", hour, dateStr, policeStation, riskLevel)
	if cachedVal, ok := hotspotsCache.Get(cacheKey); ok {
		c.JSON(http.StatusOK, cachedVal)
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

	// Build query
	query := `
		SELECT p.zone_id, p.zone_lat, p.zone_lon, p.police_station, p.priority_score,
		       p.priority_level, p.impact_score, p.predicted_hotspot_risk, p.model_confidence,
		       COALESCE(f.violations_last_1h, 0) as expected_violations
		FROM zone_predictions p
		LEFT JOIN zone_time_features f ON p.zone_id = f.zone_id AND p.hour = f.hour
		WHERE p.hour = $1
	`
	args := []interface{}{hour}
	argCount := 1

	if dateStr != "" {
		argCount++
		query += fmt.Sprintf(" AND p.prediction_time::date = $%d", argCount)
		args = append(args, dateStr)
	}

	if policeStation != "" && strings.ToUpper(policeStation) != "ALL" {
		argCount++
		query += fmt.Sprintf(" AND UPPER(p.police_station) = UPPER($%d)", argCount)
		args = append(args, policeStation)
	}

	if riskLevel != "" {
		argCount++
		query += fmt.Sprintf(" AND UPPER(p.predicted_hotspot_risk) = UPPER($%d)", argCount)
		args = append(args, riskLevel)
	}

	query += " ORDER BY p.priority_score DESC LIMIT 100"

	rows, err := database.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to query database",
			"code":  500,
		})
		return
	}
	defer rows.Close()

	hotspots := []Hotspot{}
	for rows.Next() {
		var h Hotspot
		var expViolations float64
		err := rows.Scan(
			&h.ZoneID, &h.Lat, &h.Lng, &h.PoliceStation, &h.PriorityScore,
			&h.PriorityLevel, &h.ImpactScore, &h.PredictedHotspotRisk, &h.ModelConfidence,
			&expViolations,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to scan query results",
				"code":  500,
			})
			return
		}
		h.ExpectedViolations = int(expViolations)
		hotspots = append(hotspots, h)
	}

	filters := map[string]interface{}{
		"hour": hour,
	}
	if policeStation != "" {
		filters["police_station"] = policeStation
	}
	if dateStr != "" {
		filters["date"] = dateStr
	}
	if riskLevel != "" {
		filters["risk_level"] = riskLevel
	}

	response := HotspotsResponse{
		Hotspots:       hotspots,
		Count:          len(hotspots),
		FiltersApplied: filters,
	}

	// Cache result
	hotspotsCache.Set(cacheKey, response)

	c.JSON(http.StatusOK, response)
}
