-- Pre-computed predictions (populated offline by Python)
CREATE TABLE IF NOT EXISTS zone_predictions (
    id               SERIAL PRIMARY KEY,
    zone_id          TEXT NOT NULL,
    zone_lat         DOUBLE PRECISION,
    zone_lon         DOUBLE PRECISION,
    police_station   TEXT,
    junction_name    TEXT,
    prediction_time  TIMESTAMP WITH TIME ZONE,
    hour             INT,
    day_of_week      INT,
    month            INT,
    predicted_hotspot_risk  TEXT,       -- LOW / MEDIUM / HIGH
    model_confidence        TEXT,       -- LOW / MEDIUM / HIGH
    high_prob               FLOAT,
    prob_low                FLOAT,
    prob_medium             FLOAT,
    impact_score            FLOAT,
    priority_score          FLOAT,
    priority_level          TEXT,       -- LOW / MEDIUM / HIGH / CRITICAL
    recommended_action      TEXT,
    reasons_json            JSONB,
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_zone_predictions_hour       ON zone_predictions(hour);
CREATE INDEX IF NOT EXISTS idx_zone_predictions_station    ON zone_predictions(police_station);
CREATE INDEX IF NOT EXISTS idx_zone_predictions_priority   ON zone_predictions(priority_score DESC);

-- Zone-level historical features (for live inference fallback)
CREATE TABLE IF NOT EXISTS zone_time_features (
    zone_id                  TEXT NOT NULL,
    hour                     INT NOT NULL,
    day_of_week              INT,
    police_station           TEXT,
    junction_name            TEXT,
    junction_flag            INT,
    total_violations         FLOAT,
    wrong_parking_count      FLOAT,
    no_parking_count         FLOAT,
    main_road_count          FLOAT,
    double_parking_count     FLOAT,
    near_crossing_count      FLOAT,
    near_signal_count        FLOAT,
    footpath_count           FLOAT,
    heavy_vehicle_count      FLOAT,
    medium_vehicle_count     FLOAT,
    light_vehicle_count      FLOAT,
    two_wheel_count          FLOAT,
    avg_vio_severity         FLOAT,
    max_vio_severity         FLOAT,
    avg_veh_weight           FLOAT,
    violations_last_1h       FLOAT,
    violations_last_3h       FLOAT,
    violations_last_24h      FLOAT,
    violations_last_7d       FLOAT,
    repeat_hotspot_score     FLOAT,
    historical_zone_log_total FLOAT,
    zone_hour_hist_mean      FLOAT,
    zone_dow_hist_mean       FLOAT,
    avg_confidence           FLOAT,
    PRIMARY KEY (zone_id, hour)
);
