package inference

import (
	"testing"
)

func TestImpactScore(t *testing.T) {
	// Setup a feature input that matches specific severe conditions
	f := FeatureInput{
		DoubleParkingCount: 1,  // 1 * 30 = 30
		MainRoadCount:      1,  // 1 * 25 = 25
		NearCrossingCount:  0,
		NearSignalCount:    0,
		NoParkingCount:     1,  // 1 * 8 = 8
		WrongParkingCount:  2,  // 2 * 5 = 10
		HeavyVehicleCount:  0,
		JunctionFlag:       1,  // 1 * 15 = 15
		RepeatHotspotScore: 50, // 50 * 0.4 = 20
	}

	// Expected: 30 + 25 + 8 + 10 + 15 + 20 = 108 -> capped at 100
	score := ImpactScore(f)
	if score != 100 {
		t.Errorf("Expected ImpactScore to be 100, got %f", score)
	}

	// Setup a lower score input
	f2 := FeatureInput{
		WrongParkingCount:  1,   // 5
		NoParkingCount:     2,   // 16
		RepeatHotspotScore: 10,  // 4
	}
	// Expected: 5 + 16 + 4 = 25
	score2 := ImpactScore(f2)
	if score2 != 25 {
		t.Errorf("Expected ImpactScore to be 25, got %f", score2)
	}
}

func TestPriorityScore(t *testing.T) {
	// Severe conditions, but low priority score, should hit the floor of 72.0
	f := FeatureInput{
		DoubleParkingCount: 1,
		MainRoadCount:      1,
	}
	impact := 55.0
	highProb := 0.10 // 0.25 * 10 = 2.5 + 0.75 * 55 = 41.25 + 2.5 = 43.75
	priority := PriorityScore(impact, highProb, f)
	if priority != 72.0 {
		t.Errorf("Expected PriorityScore to hit severity gate of 72.0, got %f", priority)
	}
}

func TestModelConfidence(t *testing.T) {
	// Gap < 0.10 -> LOW
	res1 := ModelConfidence([3]float32{0.4, 0.35, 0.25})
	if res1 != "LOW" {
		t.Errorf("Expected LOW, got %s", res1)
	}

	// Gap 0.10 - 0.20 -> MEDIUM
	res2 := ModelConfidence([3]float32{0.45, 0.3, 0.25})
	if res2 != "MEDIUM" {
		t.Errorf("Expected MEDIUM, got %s", res2)
	}

	// Gap >= 0.20 -> HIGH
	res3 := ModelConfidence([3]float32{0.6, 0.2, 0.2})
	if res3 != "HIGH" {
		t.Errorf("Expected HIGH, got %s", res3)
	}
}
