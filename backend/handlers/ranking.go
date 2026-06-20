package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pranav8764/ParkIntel/backend/cache"
	"github.com/pranav8764/ParkIntel/backend/db"
)

type Ranking struct {
	Rank              int     `json:"rank"`
	ZoneID            string  `json:"zone_id"`
	PoliceStation     string  `json:"police_station"`
	JunctionName      string  `json:"junction_name"`
	ExpectedViolations int     `json:"expected_violations"`
	ImpactScore       float64 `json:"impact_score"`
	PriorityScore     float64 `json:"priority_score"`
	PriorityLevel     string  `json:"priority_level"`
	ModelConfidence      string  `json:"model_confidence"`
	RecommendedAction string  `json:"recommended_action"`
}

type RankingsResponse struct {
	Rankings []Ranking `json:"rankings"`
	Total    int       `json:"total"`
	Hour     int       `json:"hour"`
}

var rankingsCache = cache.NewMemoryCache(5 * time.Minute)

// GetRanking serves GET /api/enforcement/ranking
func GetRanking(c *gin.Context) {
	hourStr := c.Query("hour")
	dateStr := c.Query("date") // Format YYYY-MM-DD
	limitStr := c.Query("limit")

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

	limit := 20
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if limit > 100 {
		limit = 100
	}

	// Cache key includes hour, date, and limit
	cacheKey := fmt.Sprintf("%d:%s:%d", hour, dateStr, limit)
	if cachedVal, ok := rankingsCache.Get(cacheKey); ok {
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
		SELECT p.zone_id, p.police_station, p.junction_name, p.impact_score, p.priority_score,
		       p.priority_level, p.model_confidence, p.recommended_action,
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

	query += fmt.Sprintf(" ORDER BY p.priority_score DESC, p.impact_score DESC LIMIT $%d", argCount+1)
	args = append(args, limit)

	rows, err := database.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to query database",
			"code":  500,
		})
		return
	}
	defer rows.Close()

	rankings := []Ranking{}
	rankCount := 1
	for rows.Next() {
		var r Ranking
		var expViolations float64
		err := rows.Scan(
			&r.ZoneID, &r.PoliceStation, &r.JunctionName, &r.ImpactScore, &r.PriorityScore,
			&r.PriorityLevel, &r.ModelConfidence, &r.RecommendedAction, &expViolations,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to scan query results",
				"code":  500,
			})
			return
		}
		r.Rank = rankCount
		r.ExpectedViolations = int(expViolations)
		rankings = append(rankings, r)
		rankCount++
	}

	response := RankingsResponse{
		Rankings: rankings,
		Total:    len(rankings),
		Hour:     hour,
	}

	// Cache result
	rankingsCache.Set(cacheKey, response)

	c.JSON(http.StatusOK, response)
}
