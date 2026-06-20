package inference

import (
	"math"
	"sort"
)

// ImpactScore calculates the absolute-count severity formula
func ImpactScore(f FeatureInput) float64 {
	score := 0.0
	score += f.DoubleParkingCount * 30
	score += f.MainRoadCount * 25
	score += f.NearCrossingCount * 25
	score += f.NearSignalCount * 25
	score += f.NoParkingCount * 8
	score += f.WrongParkingCount * 5
	score += f.HeavyVehicleCount * 10
	score += float64(f.JunctionFlag) * 15
	score += f.RepeatHotspotScore * 0.4
	if score > 100 {
		return 100
	}
	return math.Round(score*100) / 100
}

// PriorityScore computes the ML + rule blend with severity gate
func PriorityScore(impactScore float64, highProb float64, f FeatureInput) float64 {
	// Count how many high-severity violation types are present
	highSevCount := 0
	if f.DoubleParkingCount > 0 {
		highSevCount++
	}
	if f.MainRoadCount > 0 {
		highSevCount++
	}
	if f.NearCrossingCount > 0 {
		highSevCount++
	}
	if f.NearSignalCount > 0 {
		highSevCount++
	}
	if f.HeavyVehicleCount > 0 && f.JunctionFlag == 1 {
		highSevCount++
	}

	priority := 0.25*(highProb*100) + 0.75*impactScore

	// Severity gate: dangerous combinations get a floor of 72 (HIGH tier minimum)
	if highSevCount >= 2 && priority < 72.0 {
		priority = 72.0
	}
	if priority > 100 {
		priority = 100
	}
	return math.Round(priority*100) / 100
}

// ModelConfidence calculates the gap between the top two probabilities
func ModelConfidence(proba [3]float32) string {
	// Sort descending
	p := []float32{proba[0], proba[1], proba[2]}
	sort.Slice(p, func(i, j int) bool { return p[i] > p[j] })
	gap := float64(p[0] - p[1])
	if gap < 0.10 {
		return "LOW"
	} else if gap < 0.20 {
		return "MEDIUM"
	}
	return "HIGH"
}

// PriorityLevel maps the score to its text label
func PriorityLevel(score float64) string {
	switch {
	case score <= 40:
		return "LOW"
	case score <= 70:
		return "MEDIUM"
	case score <= 85:
		return "HIGH"
	default:
		return "CRITICAL"
	}
}

// RecommendedAction determines enforcement actions based on level
func RecommendedAction(level string) string {
	switch level {
	case "LOW":
		return "Monitor — low enforcement priority"
	case "MEDIUM":
		return "Schedule patrol during peak hours"
	case "HIGH":
		return "Deploy enforcement team"
	case "CRITICAL":
		return "Deploy towing/enforcement team immediately"
	default:
		return "Monitor"
	}
}

// Reasons lists rule-based explanation reasons
func Reasons(f FeatureInput) []string {
	var reasons []string
	if f.RepeatHotspotScore > 50 {
		reasons = append(reasons, "High repeat violation density")
	}
	if f.MainRoadCount > 0 {
		reasons = append(reasons, "Parking in main road violations present")
	}
	if f.JunctionFlag == 1 {
		reasons = append(reasons, "Located near a junction")
	}
	if f.HeavyVehicleCount > 0 {
		reasons = append(reasons, "Heavy vehicle parking detected")
	}
	if f.DoubleParkingCount > 0 {
		reasons = append(reasons, "Double parking violations present")
	}
	if f.NearSignalCount > 0 {
		reasons = append(reasons, "Parking near traffic light/zebra crossing")
	}
	if f.NearCrossingCount > 0 {
		reasons = append(reasons, "Parking near road crossing")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "General violation cluster")
	}
	return reasons
}
