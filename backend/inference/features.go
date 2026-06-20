package inference

import (
	"strings"
)

// OnnxMeta matches the onnx_meta.json structure
type OnnxMeta struct {
	FeatureCols          []string             `json:"feature_cols"`
	NFeatures            int                  `json:"n_features"`
	ClassNames           []string             `json:"class_names"`
	PrimaryModel         string               `json:"primary_model"`
	Models               map[string]string    `json:"models"`
	PoliceStationEncoder PoliceStationEncoder `json:"police_station_encoder"`
	PoliceStationClasses []string             `json:"-"` // Handled during startup
}

type PoliceStationEncoder struct {
	Classes []string `json:"classes"`
}

// FeatureInput matches the required 30 features
type FeatureInput struct {
	// Temporal
	Hour        float64
	DayOfWeek   float64
	Month       float64
	IsWeekend   float64
	// Spatial
	JunctionFlag float64
	// Violation counts (current window)
	WrongParkingCount  float64
	NoParkingCount     float64
	MainRoadCount      float64
	DoubleParkingCount float64
	NearCrossingCount  float64
	NearSignalCount    float64
	FootpathCount      float64
	// Vehicle counts
	HeavyVehicleCount  float64
	MediumVehicleCount float64
	LightVehicleCount  float64
	TwoWheelCount      float64
	// Severity
	AvgVioSeverity     float64
	MaxVioSeverity     float64
	AvgVehWeight       float64
	// Rolling features
	ViolationsLast1h   float64
	ViolationsLast3h   float64
	ViolationsLast24h  float64
	ViolationsLast7d   float64
	// Historical
	RepeatHotspotScore       float64
	HistoricalZoneLogTotal   float64
	TotalViolations          float64
	ZoneHourHistMean         float64
	ZoneDowHistMean          float64
	AvgConfidence            float64
	// Encoded categorical
	PoliceStationEnc         float64
}

// ToFloat32Slice returns the 30-feature vector exactly in the order specified by the prompt
func (f FeatureInput) ToFloat32Slice() []float32 {
	return []float32{
		float32(f.Hour), float32(f.DayOfWeek), float32(f.Month), float32(f.IsWeekend),
		float32(f.JunctionFlag),
		float32(f.WrongParkingCount), float32(f.NoParkingCount), float32(f.MainRoadCount),
		float32(f.DoubleParkingCount), float32(f.NearCrossingCount), float32(f.NearSignalCount),
		float32(f.FootpathCount),
		float32(f.HeavyVehicleCount), float32(f.MediumVehicleCount),
		float32(f.LightVehicleCount), float32(f.TwoWheelCount),
		float32(f.AvgVioSeverity), float32(f.MaxVioSeverity), float32(f.AvgVehWeight),
		float32(f.ViolationsLast1h), float32(f.ViolationsLast3h),
		float32(f.ViolationsLast24h), float32(f.ViolationsLast7d),
		float32(f.RepeatHotspotScore), float32(f.HistoricalZoneLogTotal),
		float32(f.TotalViolations), float32(f.ZoneHourHistMean), float32(f.ZoneDowHistMean),
		float32(f.AvgConfidence), float32(f.PoliceStationEnc),
	}
}

// ToSlice creates a custom feature slice containing only the columns specified by the model metadata
func (f FeatureInput) ToSlice(featureCols []string) []float32 {
	slice := make([]float32, len(featureCols))
	for i, col := range featureCols {
		var val float64
		switch col {
		case "hour":
			val = f.Hour
		case "day_of_week":
			val = f.DayOfWeek
		case "month":
			val = f.Month
		case "is_weekend":
			val = f.IsWeekend
		case "junction_flag":
			val = f.JunctionFlag
		case "wrong_parking_count":
			val = f.WrongParkingCount
		case "no_parking_count":
			val = f.NoParkingCount
		case "main_road_count":
			val = f.MainRoadCount
		case "double_parking_count":
			val = f.DoubleParkingCount
		case "near_crossing_count":
			val = f.NearCrossingCount
		case "near_signal_count":
			val = f.NearSignalCount
		case "footpath_count":
			val = f.FootpathCount
		case "heavy_vehicle_count":
			val = f.HeavyVehicleCount
		case "medium_vehicle_count":
			val = f.MediumVehicleCount
		case "light_vehicle_count":
			val = f.LightVehicleCount
		case "two_wheel_count":
			val = f.TwoWheelCount
		case "avg_vio_severity":
			val = f.AvgVioSeverity
		case "max_vio_severity":
			val = f.MaxVioSeverity
		case "avg_veh_weight":
			val = f.AvgVehWeight
		case "violations_last_1h":
			val = f.ViolationsLast1h
		case "violations_last_3h":
			val = f.ViolationsLast3h
		case "violations_last_24h":
			val = f.ViolationsLast24h
		case "violations_last_7d":
			val = f.ViolationsLast7d
		case "repeat_hotspot_score":
			val = f.RepeatHotspotScore
		case "historical_zone_log_total":
			val = f.HistoricalZoneLogTotal
		case "total_violations":
			val = f.TotalViolations
		case "zone_hour_hist_mean":
			val = f.ZoneHourHistMean
		case "zone_dow_hist_mean":
			val = f.ZoneDowHistMean
		case "avg_confidence":
			val = f.AvgConfidence
		case "police_station_enc":
			val = f.PoliceStationEnc
		}
		slice[i] = float32(val)
	}
	return slice
}

// EncodePoliceStation performs label encoding for police station names
func EncodePoliceStation(meta OnnxMeta, station string) float32 {
	stationUpper := strings.ToUpper(strings.TrimSpace(station))
	for i, s := range meta.PoliceStationClasses {
		if strings.ToUpper(s) == stationUpper {
			return float32(i)
		}
	}
	return 0 // fallback for unknown station
}
