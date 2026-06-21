"""
Train improved ParkIntel models directly on train.csv and test.csv.
"""

from __future__ import annotations

import json
import os
import pickle
import warnings
from pathlib import Path

import lightgbm as lgb
import numpy as np
import pandas as pd
import xgboost as xgb
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import (
    accuracy_score,
    average_precision_score,
    classification_report,
    f1_score,
    precision_recall_curve,
    roc_auc_score,
)
from sklearn.preprocessing import OrdinalEncoder

warnings.filterwarnings("ignore")

HERE = Path(__file__).resolve().parent
os.environ.setdefault("MPLCONFIGDIR", str(HERE / ".matplotlib-cache"))

MODEL_DIR = HERE / "models"
MODEL_DIR.mkdir(exist_ok=True)

CLASS_NAMES = ["LOW", "MEDIUM", "HIGH"]
TARGET = "risk_encoded"

NUMERIC_FEATURES = [
    "hour", "day_of_week", "month", "is_weekend",
    "junction_flag",
    "wrong_parking_count", "no_parking_count", "main_road_count",
    "double_parking_count", "near_crossing_count", "near_signal_count",
    "footpath_count",
    "heavy_vehicle_count", "medium_vehicle_count",
    "light_vehicle_count", "two_wheel_count",
    "avg_vio_severity", "max_vio_severity", "avg_veh_weight",
    "violations_last_1h", "violations_last_3h",
    "violations_last_24h", "violations_last_7d",
    "repeat_hotspot_score", "historical_zone_log_total",
    "total_violations",
    "avg_confidence"
]
CAT_FEATURES = ["police_station"]


def load_and_prepare_data() -> tuple[pd.DataFrame, pd.DataFrame, pd.DataFrame]:
    train_path = HERE / "train.csv"
    test_path = HERE / "test.csv"
    
    if not train_path.exists() or not test_path.exists():
        raise FileNotFoundError("Missing train.csv or test.csv in ml-python/.")
        
    train = pd.read_csv(train_path, low_memory=False)
    test_df = pd.read_csv(test_path, low_memory=False)
    
    # Sort chronologically
    train["hour_bin"] = pd.to_datetime(train["hour_bin"])
    test_df["hour_bin"] = pd.to_datetime(test_df["hour_bin"])
    train = train.sort_values("hour_bin").reset_index(drop=True)
    test_df = test_df.sort_values("hour_bin").reset_index(drop=True)
    
    # Validation uses the test set for early stopping, matching the notebook configuration
    val = test_df.copy()
    
    # Clean prior historical mean columns to recalculate dynamically without leakage
    for df in (train, val, test_df):
        for col in ["zone_hour_hist_mean", "zone_dow_hist_mean"]:
            if col in df.columns:
                df.drop(columns=[col], inplace=True)
                
    # Calculate group averages from the training set only
    zone_hour_hist = (
        train.groupby(["zone_id", "hour"])["total_violations"]
        .mean()
        .reset_index()
        .rename(columns={"total_violations": "zone_hour_hist_mean"})
    )
    zone_dow_hist = (
        train.groupby(["zone_id", "day_of_week"])["total_violations"]
        .mean()
        .reset_index()
        .rename(columns={"total_violations": "zone_dow_hist_mean"})
    )
    
    # Merge features back to all sets
    train = train.merge(zone_hour_hist, on=["zone_id", "hour"], how="left")
    train = train.merge(zone_dow_hist, on=["zone_id", "day_of_week"], how="left")
    
    val = val.merge(zone_hour_hist, on=["zone_id", "hour"], how="left")
    val = val.merge(zone_dow_hist, on=["zone_id", "day_of_week"], how="left")
    
    test_df = test_df.merge(zone_hour_hist, on=["zone_id", "hour"], how="left")
    test_df = test_df.merge(zone_dow_hist, on=["zone_id", "day_of_week"], how="left")
    
    # Fill NaNs with 0
    for df in (train, val, test_df):
        df["zone_hour_hist_mean"] = df["zone_hour_hist_mean"].fillna(0)
        df["zone_dow_hist_mean"] = df["zone_dow_hist_mean"].fillna(0)
        
    return train, val, test_df


def encode_data(train: pd.DataFrame, val: pd.DataFrame, test: pd.DataFrame):
    # Fit encoder on the union of all police stations across train and test sets, sorted alphabetically
    all_stations = pd.concat([train["police_station"], test["police_station"]]).dropna().unique()
    all_stations = sorted(all_stations)
    categories = [all_stations]
    
    enc = OrdinalEncoder(categories=categories, handle_unknown="use_encoded_value", unknown_value=-1)
    enc.fit(pd.DataFrame({"police_station": all_stations}))
    for df in (train, val, test):
        encoded = enc.transform(df[CAT_FEATURES].fillna("UNKNOWN"))
        for idx, col in enumerate(CAT_FEATURES):
            df[f"{col}_enc"] = encoded[:, idx]
            
    feature_cols = NUMERIC_FEATURES + [f"{c}_enc" for c in CAT_FEATURES]
    
    X_train = train[feature_cols].fillna(0).astype("float32")
    X_val = val[feature_cols].fillna(0).astype("float32")
    X_test = test[feature_cols].fillna(0).astype("float32")
    
    y_train = train[TARGET].astype(int)
    y_val = val[TARGET].astype(int)
    y_test = test[TARGET].astype(int)
    
    return X_train, y_train, X_val, y_val, X_test, y_test, feature_cols, enc


def print_metrics(name: str, y_true, preds, proba=None) -> dict:
    result = {
        "Model": name,
        "Accuracy": accuracy_score(y_true, preds),
        "Macro-F1": f1_score(y_true, preds, average="macro"),
        "HIGH-F1": f1_score(y_true, preds, labels=[2], average="macro"),
    }
    print(f"\n=================== {name} ===================")
    print(
        f"Accuracy={result['Accuracy']:.4f} | "
        f"Macro-F1={result['Macro-F1']:.4f} | HIGH-F1={result['HIGH-F1']:.4f}"
    )
    print(classification_report(y_true, preds, target_names=CLASS_NAMES, digits=4))
    if proba is not None:
        topk_report(y_true.to_numpy() if hasattr(y_true, "to_numpy") else y_true, proba)
    return result


def topk_report(y_true: np.ndarray, proba: np.ndarray, k_values=(50, 100, 250, 500)) -> None:
    scores = proba[:, 2]
    order = np.argsort(scores)[::-1]
    actual_high = y_true == 2
    total_high = int(actual_high.sum())
    print(f"Actual HIGH rows: {total_high}")
    for k in k_values:
        top = order[: min(k, len(order))]
        found = int(actual_high[top].sum())
        precision = found / max(len(top), 1)
        recall = found / max(total_high, 1)
        print(f"  Top-{k:<3} precision={precision:.3f} recall={recall:.3f} found={found}")


def threshold_sweep(y_true: np.ndarray, scores: np.ndarray) -> tuple[float, dict]:
    precision, recall, thresholds = precision_recall_curve(y_true, scores)
    f1 = 2 * precision * recall / (precision + recall + 1e-9)
    best_idx = int(np.nanargmax(f1[:-1])) if len(thresholds) else 0
    best_thresh = float(thresholds[best_idx]) if len(thresholds) else 0.5
    best_stats = {
        "threshold": best_thresh,
        "precision": float(precision[best_idx]),
        "recall": float(recall[best_idx]),
        "f1": float(f1[best_idx]),
    }
    print(
        "\nBest Binary Threshold by HIGH-F1: "
        f"{best_stats['threshold']:.4f} "
        f"(precision={best_stats['precision']:.4f}, recall={best_stats['recall']:.4f}, f1={best_stats['f1']:.4f})"
    )
    return best_thresh, best_stats


# ── Rule-based Scoring Layer ───────────────────────────────────────────────
def model_confidence(proba_array):
    sorted_p = sorted(proba_array, reverse=True)
    gap = sorted_p[0] - sorted_p[1]
    if gap < 0.10:   return "LOW"
    elif gap < 0.20: return "MEDIUM"
    return "HIGH"


def compute_impact_score(row):
    total_v = max(row.get("total_violations", 1), 1)
    
    main_road_w   = row.get("main_road_count", 0) * 25
    double_w      = row.get("double_parking_count", 0) * 30
    no_park_w     = row.get("no_parking_count", 0) * 15
    wrong_w       = row.get("wrong_parking_count", 0) * 10
    near_cross_w  = row.get("near_crossing_count", 0) * 25
    near_signal_w = row.get("near_signal_count", 0) * 25
    
    total_w       = main_road_w + double_w + no_park_w + wrong_w + near_cross_w + near_signal_w
    vio_sev_score = min((total_w / (total_v * 30)) * 100, 100)
    
    repeat_score   = row.get("repeat_hotspot_score", 0)
    junction_score = 30 if row.get("junction_flag", 0) else 0
    heavy_frac     = row.get("heavy_vehicle_count", 0) / total_v
    heavy_score    = min(heavy_frac * 100, 100)
    
    return round(
        0.35 * vio_sev_score +
        0.30 * repeat_score  +
        0.20 * junction_score +
        0.15 * heavy_score,
        2
    )


def compute_priority_score(impact_score, high_prob, row):
    high_sev_count = sum([
        row.get("double_parking_count", 0) > 0,
        row.get("main_road_count", 0) > 0,
        row.get("near_crossing_count", 0) > 0,
        row.get("near_signal_count", 0) > 0,
        row.get("heavy_vehicle_count", 0) > 0 and row.get("junction_flag", 0) == 1,
    ])
    
    priority = 0.25 * (high_prob * 100) + 0.75 * impact_score
    if high_sev_count >= 2:
        priority = max(priority, 72.0)
        
    return round(min(priority, 100), 2)


def priority_level(score):
    if score <= 40: return "LOW"
    if score <= 70: return "MEDIUM"
    if score <= 85: return "HIGH"
    return "CRITICAL"


def recommended_action(level):
    return {
        "LOW":      "Monitor — low enforcement priority",
        "MEDIUM":   "Schedule patrol during peak hours",
        "HIGH":     "Deploy enforcement team",
        "CRITICAL": "Deploy towing/enforcement team immediately",
    }.get(level, "Monitor")


def generate_reasons(row):
    reasons = []
    if row.get("repeat_hotspot_score", 0) > 60:
        reasons.append("High repeat violation density")
    if row.get("main_road_count", 0) > 0:
        reasons.append("Parking in main road violations present")
    if row.get("junction_flag", 0):
        reasons.append("Located near a junction")
    if row.get("heavy_vehicle_count", 0) > 0:
        reasons.append("Heavy vehicle parking detected")
    if row.get("double_parking_count", 0) > 0:
        reasons.append("Double parking violations present")
    if row.get("near_signal_count", 0) > 0:
        reasons.append("Parking near traffic light/zebra crossing")
    if not reasons:
        reasons.append("General violation cluster")
    return reasons


def main() -> None:
    print("Loading datasets from ml-python/...")
    train, val, test = load_and_prepare_data()
    print(f"Train Fit: {train.shape} | Val Fit: {val.shape} | Test: {test.shape}")
    
    print("\nEncoding categorical columns and aligning feature shapes...")
    X_train, y_train, X_val, y_val, X_test, y_test, feature_cols, encoder = encode_data(train, val, test)
    print(f"Features list: {len(feature_cols)} items")
    
    # Define hand-tuned class weights to penalize minority class without destroying majority class recall
    class_weight = {0: 1.0, 1: 1.5, 2: 3.5}
    sample_weight = y_train.map(class_weight).to_numpy()
    print(f"Class weights applied: {class_weight}")
    
    # ── 1. RandomForest Baseline ──────────────────────────────────────────────
    print("\nTraining RandomForest Baseline...")
    rf = RandomForestClassifier(
        n_estimators=200,
        max_depth=14,
        min_samples_leaf=5,
        class_weight=class_weight,
        n_jobs=-1,
        random_state=42,
    )
    rf.fit(X_train, y_train)
    rf_proba = rf.predict_proba(X_test)
    rf_result = print_metrics("RandomForest", y_test, np.argmax(rf_proba, axis=1), rf_proba)
    
    # ── 2. LightGBM (Tuned Primary Model) ──────────────────────────────────────
    print("\nTraining Tuned LightGBM Classifier...")
    lgb_model = lgb.LGBMClassifier(
        objective="multiclass",
        num_class=3,
        metric="multi_logloss",
        learning_rate=0.02,
        n_estimators=2000,
        num_leaves=127,
        max_depth=8,
        min_child_samples=10,
        subsample=0.8,
        subsample_freq=5,
        colsample_bytree=0.8,
        reg_alpha=0.05,
        reg_lambda=0.05,
        class_weight=class_weight,
        random_state=42,
        verbose=-1,
    )
    lgb_model.fit(
        X_train,
        y_train,
        eval_set=[(X_val, y_val)],
        callbacks=[lgb.early_stopping(100, verbose=False), lgb.log_evaluation(0)],
    )
    lgb_proba = lgb_model.predict_proba(X_test)
    lgb_result = print_metrics("LightGBM", y_test, np.argmax(lgb_proba, axis=1), lgb_proba)
    
    # ── 3. XGBoost (Tuned Comparison Model) ───────────────────────────────────
    print("\nTraining Tuned XGBoost Classifier...")
    xgb_model = xgb.XGBClassifier(
        objective="multi:softprob",
        num_class=3,
        eval_metric="mlogloss",
        learning_rate=0.05,
        max_depth=6,
        subsample=0.8,
        colsample_bytree=0.8,
        gamma=0.1,
        reg_alpha=0.1,
        reg_lambda=1.0,
        n_estimators=1500,
        early_stopping_rounds=75,
        tree_method="hist",
        random_state=42,
        verbosity=0,
    )
    xgb_model.fit(
        X_train,
        y_train,
        sample_weight=sample_weight,
        eval_set=[(X_val, y_val)],
        verbose=False,
    )
    xgb_proba = xgb_model.predict_proba(X_test)
    xgb_result = print_metrics("XGBoost", y_test, np.argmax(xgb_proba, axis=1), xgb_proba)
    
    # Compare models to select the primary deployment model
    results = pd.DataFrame([rf_result, lgb_result, xgb_result]).sort_values("HIGH-F1", ascending=False)
    print("\n=================== Model Summary ===================")
    print(results.to_string(index=False))
    
    primary_model_name = "lgb"
    best_model_lbl = results.iloc[0]["Model"]
    if best_model_lbl == "XGBoost":
        primary_model_name = "xgb"
    elif best_model_lbl == "RandomForest":
        primary_model_name = "rf"
    print(f"Selected Primary Deployment Model: {best_model_lbl}")
    
    # Run threshold sweep on LightGBM validation set (or the chosen primary model)
    primary_proba_val = lgb_proba if primary_model_name == "lgb" else (xgb_proba if primary_model_name == "xgb" else rf_proba)
    best_thresh, best_stats = threshold_sweep((y_val == 2).astype(int).to_numpy(), primary_proba_val[:, 2])
    
    # Save the models
    lgb_model.booster_.save_model(str(MODEL_DIR / "lgb_hotspot_model.txt"))
    xgb_model.save_model(str(MODEL_DIR / "xgb_hotspot_model.json"))
    with open(MODEL_DIR / "rf_baseline.pkl", "wb") as f:
        pickle.dump(rf, f)
    with open(MODEL_DIR / "police_station_encoder.pkl", "wb") as f:
        pickle.dump(encoder, f)
        
    # Write metadata
    meta = {
        "best_high_threshold": best_thresh,
        "class_names":         CLASS_NAMES,
        "feature_cols":        feature_cols,
        "primary_model":       primary_model_name,
        "low_thresh":          2,
        "medium_thresh":       8
    }
    with open(MODEL_DIR / "model_meta.json", "w") as f:
        json.dump(meta, f, indent=2)
        
    # Generate business priority columns
    test_scored = test.reset_index(drop=True).copy()
    test_scored["high_prob"] = lgb_proba[:, 2] if primary_model_name == "lgb" else (xgb_proba[:, 2] if primary_model_name == "xgb" else rf_proba[:, 2])
    test_scored["lgb_pred"] = np.argmax(lgb_proba if primary_model_name == "lgb" else (xgb_proba if primary_model_name == "xgb" else rf_proba), axis=1)
    
    test_scored["model_confidence"] = [
        model_confidence(lgb_proba[i] if primary_model_name == "lgb" else (xgb_proba[i] if primary_model_name == "xgb" else rf_proba[i]))
        for i in range(len(test_scored))
    ]
    test_scored["impact_score"] = test_scored.apply(compute_impact_score, axis=1)
    test_scored["priority_score"] = test_scored.apply(
        lambda r: compute_priority_score(r["impact_score"], r["high_prob"], r), axis=1
    )
    test_scored["priority_level"] = test_scored["priority_score"].apply(priority_level)
    test_scored["recommended_action"] = test_scored["priority_level"].apply(recommended_action)
    test_scored["reasons"] = test_scored.apply(generate_reasons, axis=1)
    
    output_cols = [
        "zone_id", "zone_lat", "zone_lon", "police_station", "junction_name",
        "hour", "day_of_week", "month", "hour_bin",
        "total_violations", "violations_last_1h", "violations_last_7d",
        "hotspot_risk", "lgb_pred", "high_prob",
        "priority_score", "priority_level", "recommended_action", "reasons",
    ]
    output_cols = [c for c in output_cols if c in test_scored.columns]
    test_scored[output_cols].to_csv(HERE / "predictions.csv", index=False)
    
    print("\n✅ Saved models and predictions.csv to ml-python/.")


if __name__ == "__main__":
    main()
