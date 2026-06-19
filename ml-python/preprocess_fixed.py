"""
Leak-resistant preprocessing for ParkIntel hotspot prediction.

This script fixes the main issues in the notebook pipeline:
- keeps inactive zone-hours so the model can learn true LOW/no-hotspot cases
- computes historical zone totals with expanding history only, not full-data totals
- builds labels from next-hour violations
- creates chronological train/validation/test CSVs
"""

from __future__ import annotations

import ast
import json
from pathlib import Path

import numpy as np
import pandas as pd


HERE = Path(__file__).resolve().parent
DATA_DIR = HERE / "data"
RAW_DIR = DATA_DIR / "raw"
OUT_DIR = DATA_DIR / "processed"
RAW_CANDIDATES = [
    RAW_DIR / "grid-r2-t1-dataset.csv",
    RAW_DIR / "dataset.csv",
    HERE / "dataset.csv",
]
OUT_DIR.mkdir(exist_ok=True)

DATE_MIN = pd.Timestamp("2024-01-01", tz="UTC")
DATE_MAX = pd.Timestamp("2024-06-01", tz="UTC")

BLR_LAT_MIN, BLR_LAT_MAX = 12.70, 13.20
BLR_LON_MIN, BLR_LON_MAX = 77.30, 77.85

CONF_MAP = {
    "approved": 1.0,
    "created": 0.7,
    "processing": 0.6,
    "pending": 0.7,
    "rejected": None,
    "duplicate": None,
}

VIO_SEVERITY = {
    "DOUBLE PARKING": 30,
    "PARKING IN A MAIN ROAD": 25,
    "PARKING NEAR ROAD CROSSING": 25,
    "PARKING NEAR TRAFFIC LIGHT/ZEBRA CROSSING": 25,
    "PARKING NEAR BUS STOP": 20,
    "PARKING NEAR SCHOOL": 20,
    "PARKING NEAR HOSPITAL": 20,
    "PARKING OPPOSITE ANOTHER PARKED VEHICLE": 20,
    "NO PARKING": 15,
    "PARKING ON FOOTPATH": 12,
    "WRONG PARKING": 10,
    "DEFECTIVE NUMBER PLATE": 3,
}
VIO_DEFAULT_WEIGHT = 5

VEHICLE_GROUPS = {
    "BUS": "HEAVY",
    "PRIVATE BUS": "HEAVY",
    "TOURIST BUS": "HEAVY",
    "SCHOOL BUS": "HEAVY",
    "HGV": "HEAVY",
    "LORRY": "HEAVY",
    "TANKER": "HEAVY",
    "LGV": "MEDIUM",
    "VAN": "MEDIUM",
    "TEMPO": "MEDIUM",
    "MAXI-CAB": "MEDIUM",
    "MAXI CAB": "MEDIUM",
    "CAR": "LIGHT",
    "JEEP": "LIGHT",
    "SUV": "LIGHT",
    "SCOOTER": "TWO_WHEEL",
    "MOTORCYCLE": "TWO_WHEEL",
    "MOPED": "TWO_WHEEL",
    "BIKE": "TWO_WHEEL",
    "AUTO": "AUTO",
    "AUTORICKSHAW": "AUTO",
    "GOODS AUTO": "AUTO",
    "PASSENGER AUTO": "AUTO",
}
VEH_WEIGHT = {
    "HEAVY": 25,
    "MEDIUM": 18,
    "LIGHT": 12,
    "AUTO": 10,
    "TWO_WHEEL": 5,
    "OTHER": 5,
    "UNKNOWN": 5,
}

NON_JUNCTION_VALUES = {
    "NO JUNCTION",
    "NO_JUNCTION",
    "NONE",
    "NA",
    "N/A",
    "NOT APPLICABLE",
    "NIL",
    "-",
    "",
    "NO JUNCTION NAME",
}

COUNT_COLS = [
    "wrong_parking_count",
    "no_parking_count",
    "main_road_count",
    "double_parking_count",
    "near_crossing_count",
    "near_signal_count",
    "footpath_count",
    "heavy_vehicle_count",
    "medium_vehicle_count",
    "light_vehicle_count",
    "two_wheel_count",
]


def raw_path() -> Path:
    for path in RAW_CANDIDATES:
        if path.exists():
            return path
    raise FileNotFoundError(
        "Raw dataset not found. Expected one of:\n"
        + "\n".join(str(p) for p in RAW_CANDIDATES)
    )


def clean_str(value, fallback: str = "UNKNOWN") -> str:
    if pd.isna(value):
        return fallback
    return str(value).strip().upper()


def normalize_junction(value) -> str:
    cleaned = clean_str(value, fallback="NO_JUNCTION")
    if cleaned in NON_JUNCTION_VALUES:
        return "NO_JUNCTION"
    return cleaned


def parse_violation_type(raw) -> list[str]:
    if pd.isna(raw):
        return ["UNKNOWN"]
    text = str(raw).strip()
    if text.startswith("["):
        for parser in (lambda s: json.loads(s.replace("'", '"')), ast.literal_eval):
            try:
                return [str(item).strip().upper() for item in parser(text) if item]
            except Exception:
                pass
    return [item.strip().upper() for item in text.split(",") if item.strip()]


def map_vehicle(raw) -> str:
    if pd.isna(raw):
        return "UNKNOWN"
    return VEHICLE_GROUPS.get(str(raw).strip().upper(), "OTHER")


def reduce_mem(df: pd.DataFrame) -> pd.DataFrame:
    int_cols = df.select_dtypes(include=["int64"]).columns
    float_cols = df.select_dtypes(include=["float64"]).columns
    for col in int_cols:
        df[col] = pd.to_numeric(df[col], downcast="integer")
    for col in float_cols:
        df[col] = pd.to_numeric(df[col], downcast="float")
    return df


def build_active_hourly(df: pd.DataFrame) -> pd.DataFrame:
    df["vio_severity_val"] = df["primary_violation"].map(VIO_SEVERITY).fillna(VIO_DEFAULT_WEIGHT)
    df["veh_weight_val"] = df["vehicle_normalized"].map(VEH_WEIGHT).fillna(5)
    df["junction_flag"] = (df["junction_name"] != "NO_JUNCTION").astype("int8")

    group_cols = ["zone_id", "hour_bin"]
    agg = df.groupby(group_cols).agg(
        total_violations=("id", "count"),
        police_station=("police_station", lambda x: x.mode().iat[0] if not x.mode().empty else "UNKNOWN_STATION"),
        junction_name=("junction_name", lambda x: x.mode().iat[0] if not x.mode().empty else "NO_JUNCTION"),
        junction_flag=("junction_flag", "max"),
        wrong_parking_count=("primary_violation", lambda x: (x == "WRONG PARKING").sum()),
        no_parking_count=("primary_violation", lambda x: (x == "NO PARKING").sum()),
        main_road_count=("primary_violation", lambda x: (x == "PARKING IN A MAIN ROAD").sum()),
        double_parking_count=("primary_violation", lambda x: (x == "DOUBLE PARKING").sum()),
        near_crossing_count=("primary_violation", lambda x: (x == "PARKING NEAR ROAD CROSSING").sum()),
        near_signal_count=("primary_violation", lambda x: (x == "PARKING NEAR TRAFFIC LIGHT/ZEBRA CROSSING").sum()),
        footpath_count=("primary_violation", lambda x: (x == "PARKING ON FOOTPATH").sum()),
        heavy_vehicle_count=("vehicle_normalized", lambda x: (x == "HEAVY").sum()),
        medium_vehicle_count=("vehicle_normalized", lambda x: (x == "MEDIUM").sum()),
        light_vehicle_count=("vehicle_normalized", lambda x: (x == "LIGHT").sum()),
        two_wheel_count=("vehicle_normalized", lambda x: (x == "TWO_WHEEL").sum()),
        avg_vio_severity=("vio_severity_val", "mean"),
        max_vio_severity=("vio_severity_val", "max"),
        avg_veh_weight=("veh_weight_val", "mean"),
        avg_confidence=("record_confidence", "mean"),
    )
    return agg.reset_index()


def complete_zone_panel(df: pd.DataFrame, active: pd.DataFrame) -> pd.DataFrame:
    zone_meta = df.groupby("zone_id").agg(
        zone_lat=("zone_lat", "first"),
        zone_lon=("zone_lon", "first"),
        police_station=("police_station", lambda x: x.mode().iat[0] if not x.mode().empty else "UNKNOWN_STATION"),
        junction_name=("junction_name", lambda x: x.mode().iat[0] if not x.mode().empty else "NO_JUNCTION"),
        junction_flag=("junction_flag", "max"),
    )

    hours = pd.date_range(
        df["hour_bin"].min().floor("h"),
        df["hour_bin"].max().floor("h"),
        freq="h",
        tz="UTC",
    )
    full_index = pd.MultiIndex.from_product(
        [zone_meta.index, hours], names=["zone_id", "hour_bin"]
    )

    panel = active.set_index(["zone_id", "hour_bin"]).reindex(full_index)
    panel = panel.reset_index()
    panel = panel.merge(zone_meta.reset_index(), on="zone_id", how="left", suffixes=("", "_meta"))

    for col in ["police_station", "junction_name", "junction_flag"]:
        meta_col = f"{col}_meta"
        panel[col] = panel[col].fillna(panel[meta_col])
        panel = panel.drop(columns=[meta_col])

    panel["total_violations"] = panel["total_violations"].fillna(0).astype("int16")
    for col in COUNT_COLS:
        panel[col] = panel[col].fillna(0).astype("int16")

    panel["avg_vio_severity"] = panel["avg_vio_severity"].fillna(0)
    panel["max_vio_severity"] = panel["max_vio_severity"].fillna(0)
    panel["avg_veh_weight"] = panel["avg_veh_weight"].fillna(0)
    panel["avg_confidence"] = panel["avg_confidence"].fillna(0)
    panel["junction_flag"] = panel["junction_flag"].fillna(0).astype("int8")
    return panel.sort_values(["zone_id", "hour_bin"]).reset_index(drop=True)


def add_time_and_history(panel: pd.DataFrame) -> pd.DataFrame:
    panel["hour"] = panel["hour_bin"].dt.hour.astype("int8")
    panel["day_of_week"] = panel["hour_bin"].dt.dayofweek.astype("int8")
    panel["month"] = panel["hour_bin"].dt.month.astype("int8")
    panel["is_weekend"] = (panel["day_of_week"] >= 5).astype("int8")
    panel["hour_sin"] = np.sin(2 * np.pi * panel["hour"] / 24)
    panel["hour_cos"] = np.cos(2 * np.pi * panel["hour"] / 24)
    panel["dow_sin"] = np.sin(2 * np.pi * panel["day_of_week"] / 7)
    panel["dow_cos"] = np.cos(2 * np.pi * panel["day_of_week"] / 7)

    g = panel.groupby("zone_id", group_keys=False)
    shifted = g["total_violations"].shift(1).fillna(0)
    panel["violations_last_1h"] = shifted.astype("int16")
    panel["violations_last_3h"] = g["total_violations"].transform(
        lambda s: s.shift(1).rolling(3, min_periods=1).sum()
    ).fillna(0).astype("int16")
    panel["violations_last_24h"] = g["total_violations"].transform(
        lambda s: s.shift(1).rolling(24, min_periods=1).sum()
    ).fillna(0).astype("int16")
    panel["violations_last_7d"] = g["total_violations"].transform(
        lambda s: s.shift(1).rolling(168, min_periods=1).sum()
    ).fillna(0).astype("int16")
    panel["active_hours_last_7d"] = g["total_violations"].transform(
        lambda s: s.shift(1).gt(0).rolling(168, min_periods=1).sum()
    ).fillna(0).astype("int16")
    panel["repeat_hotspot_score"] = (panel["active_hours_last_7d"] / 168 * 100).clip(0, 100)
    panel["zone_total_violations"] = (
        g["total_violations"].cumsum() - panel["total_violations"]
    ).clip(lower=0).astype("int32")
    panel["historical_zone_log_total"] = np.log1p(panel["zone_total_violations"])

    for source in COUNT_COLS + ["avg_vio_severity", "max_vio_severity", "avg_veh_weight", "avg_confidence"]:
        panel[f"prev_{source}"] = g[source].shift(1).fillna(0)

    panel["same_hour_prev_day"] = g["total_violations"].shift(24).fillna(0).astype("int16")
    panel["same_hour_prev_week"] = g["total_violations"].shift(168).fillna(0).astype("int16")
    panel["violations_next_1h"] = g["total_violations"].shift(-1)
    panel = panel.dropna(subset=["violations_next_1h"]).copy()
    panel["violations_next_1h"] = panel["violations_next_1h"].astype("int16")
    return reduce_mem(panel)


def label_risk(v: int) -> str:
    if v == 0:
        return "LOW"
    if v < 6:
        return "MEDIUM"
    return "HIGH"


def downsample_train_low(train: pd.DataFrame, random_state: int = 42) -> pd.DataFrame:
    high_med = train[train["hotspot_risk"] != "LOW"]
    low = train[train["hotspot_risk"] == "LOW"]
    max_low = min(len(low), max(len(high_med) * 2, 50_000))
    low_sample = low.sample(n=max_low, random_state=random_state) if len(low) > max_low else low
    sampled = pd.concat([high_med, low_sample], ignore_index=True)
    return sampled.sample(frac=1, random_state=random_state).reset_index(drop=True)


def main() -> None:
    source = raw_path()
    print(f"Reading raw dataset: {source}")
    df = pd.read_csv(source, low_memory=False)

    df["created_datetime"] = pd.to_datetime(df["created_datetime"], utc=True, errors="coerce")
    df = df.dropna(subset=["created_datetime"])
    df = df[(df["created_datetime"] >= DATE_MIN) & (df["created_datetime"] < DATE_MAX)].copy()
    df["latitude"] = pd.to_numeric(df["latitude"], errors="coerce")
    df["longitude"] = pd.to_numeric(df["longitude"], errors="coerce")
    df = df.dropna(subset=["latitude", "longitude"])
    df = df[
        df["latitude"].between(BLR_LAT_MIN, BLR_LAT_MAX)
        & df["longitude"].between(BLR_LON_MIN, BLR_LON_MAX)
    ].copy()

    df["validation_status"] = df["validation_status"].fillna("created").str.lower().str.strip()
    df["record_confidence"] = df["validation_status"].map(CONF_MAP)
    df = df[df["record_confidence"].notnull()].copy()
    df["record_confidence"] = df["record_confidence"].astype("float32")

    df["police_station"] = df["police_station"].apply(lambda x: clean_str(x, "UNKNOWN_STATION"))
    df["junction_name"] = df["junction_name"].apply(normalize_junction)
    df["violation_list"] = df["violation_type"].apply(parse_violation_type)
    df["primary_violation"] = df["violation_list"].apply(lambda x: x[0] if x else "UNKNOWN")
    df["vehicle_normalized"] = df["vehicle_type"].apply(map_vehicle)
    df["zone_lat"] = df["latitude"].round(3)
    df["zone_lon"] = df["longitude"].round(3)
    df["zone_id"] = "Z_" + df["zone_lat"].astype(str) + "_" + df["zone_lon"].astype(str)
    df["hour_bin"] = df["created_datetime"].dt.floor("h")

    print(f"Clean rows: {len(df):,}; zones: {df['zone_id'].nunique():,}")
    active = build_active_hourly(df)
    print(f"Active zone-hours: {len(active):,}")
    panel = complete_zone_panel(df, active)
    print(f"Complete zone-hour panel: {panel.shape}")
    panel = add_time_and_history(panel)

    panel["hotspot_risk"] = panel["violations_next_1h"].apply(label_risk)
    panel["risk_encoded"] = panel["hotspot_risk"].map({"LOW": 0, "MEDIUM": 1, "HIGH": 2}).astype("int8")

    hours = panel["hour_bin"].sort_values().unique()
    train_cut = hours[int(len(hours) * 0.70)]
    val_cut = hours[int(len(hours) * 0.85)]

    train_full = panel[panel["hour_bin"] <= train_cut].copy()
    val = panel[(panel["hour_bin"] > train_cut) & (panel["hour_bin"] <= val_cut)].copy()
    test = panel[panel["hour_bin"] > val_cut].copy()
    train = downsample_train_low(train_full)

    print("Class distributions:")
    for name, data in [("train_full", train_full), ("train_sampled", train), ("val", val), ("test", test)]:
        print(f"\n{name}: {len(data):,}")
        print(data["hotspot_risk"].value_counts().to_string())

    train.to_csv(OUT_DIR / "train_fixed.csv", index=False)
    val.to_csv(OUT_DIR / "val_fixed.csv", index=False)
    test.to_csv(OUT_DIR / "test_fixed.csv", index=False)

    meta = {
        "source_csv": str(source),
        "train_cutoff": str(train_cut),
        "val_cutoff": str(val_cut),
        "low_definition": "0 next-hour violations",
        "medium_definition": "1-5 next-hour violations",
        "high_definition": ">=6 next-hour violations",
        "negative_sampling": "train LOW rows capped at 2x non-LOW, minimum cap 50k",
    }
    (OUT_DIR / "preprocess_fixed_meta.json").write_text(json.dumps(meta, indent=2))
    print(f"\nSaved fixed data to {OUT_DIR}")


if __name__ == "__main__":
    main()
