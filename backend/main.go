package main

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	db := SetupDB()

	r := gin.Default()
	r.Use(cors.Default()) // Allow all origins for the hackathon

	r.GET("/api/hotspots", func(c *gin.Context) {
		dateStr := c.Query("date") // Format expected: YYYY-MM-DD
		hourStr := c.Query("hour")
		policeStation := c.Query("police_station")

		query := db.Model(&ZonePrediction{})

		// Example filtering logic
		if dateStr != "" {
			// Basic substring match since hourBin is a timestamp
			query = query.Where("date(hour_bin) = ?", dateStr)
		}
		if hourStr != "" {
			hour, err := strconv.Atoi(hourStr)
			if err == nil {
				query = query.Where("hour = ?", hour)
			}
		}
		if policeStation != "" && strings.ToUpper(policeStation) != "ALL" {
			query = query.Where("UPPER(police_station) = ?", strings.ToUpper(policeStation))
		}

		var results []ZonePrediction
		query.Find(&results)

		var hotspots []HotspotResponse
		for _, r := range results {
			hotspots = append(hotspots, HotspotResponse{
				ZoneID:             r.ZoneID,
				Lat:                r.Latitude,
				Lng:                r.Longitude,
				PriorityScore:      r.PriorityScore,
				PriorityLevel:      r.PriorityLevel,
				ImpactScore:        r.PriorityScore, // Using priority as impact
				ExpectedViolations: r.ViolationsLast1H, // Proxy
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"hotspots": hotspots,
		})
	})

	r.GET("/api/enforcement/ranking", func(c *gin.Context) {
		dateStr := c.Query("date")
		hourStr := c.Query("hour")

		query := db.Model(&ZonePrediction{})

		if dateStr != "" {
			query = query.Where("date(hour_bin) = ?", dateStr)
		}
		if hourStr != "" {
			hour, err := strconv.Atoi(hourStr)
			if err == nil {
				query = query.Where("hour = ?", hour)
			}
		}

		var results []ZonePrediction
		query.Order("priority_score DESC").Limit(100).Find(&results) // Limit to top 100 for ranking

		var rankings []RankingResponse
		for i, r := range results {
			rankings = append(rankings, RankingResponse{
				Rank:              i + 1,
				ZoneID:            r.ZoneID,
				PoliceStation:     r.PoliceStation,
				JunctionName:      r.JunctionName,
				PriorityScore:     r.PriorityScore,
				PriorityLevel:     r.PriorityLevel,
				RecommendedAction: r.RecommendedAction,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"rankings": rankings,
		})
	})

	r.GET("/api/zones/:zone_id/insights", func(c *gin.Context) {
		zoneID := c.Param("zone_id")

		var r ZonePrediction
		if err := db.Where("zone_id = ?", zoneID).First(&r).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Zone not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		finalOut := r.ToFinalOutput()

		response := ZoneInsightsResponse{
			ZoneID:             r.ZoneID,
			TotalViolations:    r.ViolationsLast7D, // Proxy for total
			RepeatHotspotScore: r.PriorityScore,    // Proxy
			TopViolationTypes:  []string{"WRONG PARKING", "NO PARKING"}, // Mock data
			TopVehicleTypes:    []string{"CAR", "SCOOTER"}, // Mock data
			ImpactScore:        finalOut.ParkingCongestionImpactScore,
			Reasons:            finalOut.Reasons,
		}

		c.JSON(http.StatusOK, response)
	})

	r.POST("/api/simulate", func(c *gin.Context) {
		var req SimulateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var r ZonePrediction
		if err := db.Where("zone_id = ?", req.ZoneID).First(&r).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Zone not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		reductionFactor := 1.0 - (req.ViolationReductionPercent / 100.0)
		simulatedScore := math.Max(0, r.PriorityScore * reductionFactor)

		newLevel := "LOW"
		if simulatedScore > 85 {
			newLevel = "CRITICAL"
		} else if simulatedScore > 70 {
			newLevel = "HIGH"
		} else if simulatedScore > 40 {
			newLevel = "MEDIUM"
		}

		c.JSON(http.StatusOK, SimulateResponse{
			CurrentPriorityScore:     r.PriorityScore,
			SimulatedPriorityScore:   math.Round(simulatedScore*100)/100,
			PriorityChange:           r.PriorityLevel + " → " + newLevel,
			EstimatedImpactReduction: math.Round((r.PriorityScore - simulatedScore)*100)/100,
		})
	})

	// Run the server
	r.Run(":8080")
}
