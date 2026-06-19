package main

import (
	"encoding/json"
	"time"
)

// ZonePrediction maps to the database table zone_predictions
type ZonePrediction struct {
	ID                     uint   `gorm:"primaryKey"`
	ZoneID                 string `gorm:"index"`
	Latitude               float64
	Longitude              float64
	PoliceStation          string `gorm:"index"`
	JunctionName           string
	Hour                   int `gorm:"index"`
	DayOfWeek              int
	Month                  int
	HourBin                time.Time
	ViolationsLast1H       int
	ViolationsLast24H      int
	ViolationsLast7D       int
	PredictedHotspotRisk   string
	HighProb               float64
	PriorityScore          float64 `gorm:"index"`
	PriorityLevel          string
	RecommendedAction      string
	ReasonsJSON            string // store as json string
}

// Ensure table name is zone_predictions
func (ZonePrediction) TableName() string {
	return "zone_predictions"
}

// Output struct for Hotspots API (GET /api/hotspots)
type HotspotResponse struct {
	ZoneID            string  `json:"zone_id"`
	Lat               float64 `json:"lat"`
	Lng               float64 `json:"lng"`
	PriorityScore     float64 `json:"priority_score"`
	PriorityLevel     string  `json:"priority_level"`
	ImpactScore       float64 `json:"impact_score"`
	ExpectedViolations int    `json:"expected_violations"`
}

// Output struct for Enforcement Ranking API (GET /api/enforcement/ranking)
type RankingResponse struct {
	Rank              int     `json:"rank"`
	ZoneID            string  `json:"zone_id"`
	PoliceStation     string  `json:"police_station"`
	JunctionName      string  `json:"junction_name"`
	PriorityScore     float64 `json:"priority_score"`
	PriorityLevel     string  `json:"priority_level"`
	RecommendedAction string  `json:"recommended_action"`
}

// Output struct for Zone Insights API (GET /api/zones/:zone_id/insights)
type ZoneInsightsResponse struct {
	ZoneID             string   `json:"zone_id"`
	TotalViolations    int      `json:"total_violations"`
	RepeatHotspotScore float64  `json:"repeat_hotspot_score"`
	TopViolationTypes  []string `json:"top_violation_types"`
	TopVehicleTypes    []string `json:"top_vehicle_types"`
	ImpactScore        float64  `json:"impact_score"`
	Reasons            []string `json:"reasons"`
}

// Output struct for Simulation API (POST /api/simulate)
type SimulateRequest struct {
	ZoneID                    string  `json:"zone_id"`
	ViolationReductionPercent float64 `json:"violation_reduction_percent"`
}

type SimulateResponse struct {
	CurrentPriorityScore      float64 `json:"current_priority_score"`
	SimulatedPriorityScore    float64 `json:"simulated_priority_score"`
	PriorityChange            string  `json:"priority_change"`
	EstimatedImpactReduction  float64 `json:"estimated_impact_reduction"`
}

// Model Input Output matches instructions
type FinalOutputExample struct {
	ZoneID                       string   `json:"zone_id"`
	Latitude                     float64  `json:"latitude"`
	Longitude                    float64  `json:"longitude"`
	PoliceStation                string   `json:"police_station"`
	JunctionName                 string   `json:"junction_name"`
	PredictedHotspotRisk         string   `json:"predicted_hotspot_risk"`
	ExpectedViolationsNextHour   int      `json:"expected_violations_next_hour"`
	ParkingCongestionImpactScore float64  `json:"parking_congestion_impact_score"`
	PriorityScore                float64  `json:"priority_score"`
	PriorityLevel                string   `json:"priority_level"`
	RecommendedAction            string   `json:"recommended_action"`
	Reasons                      []string `json:"reasons"`
}

func (z *ZonePrediction) ToFinalOutput() FinalOutputExample {
	var reasons []string
	json.Unmarshal([]byte(z.ReasonsJSON), &reasons)
	if len(reasons) == 0 {
		reasons = []string{"General violation cluster"}
	}

	return FinalOutputExample{
		ZoneID:                       z.ZoneID,
		Latitude:                     z.Latitude,
		Longitude:                    z.Longitude,
		PoliceStation:                z.PoliceStation,
		JunctionName:                 z.JunctionName,
		PredictedHotspotRisk:         z.PredictedHotspotRisk,
		ExpectedViolationsNextHour:   z.ViolationsLast1H, // Proxy
		ParkingCongestionImpactScore: z.PriorityScore,    // Proxy
		PriorityScore:                z.PriorityScore,
		PriorityLevel:                z.PriorityLevel,
		RecommendedAction:            z.RecommendedAction,
		Reasons:                      reasons,
	}
}
