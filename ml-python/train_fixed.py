"""
Train corrected ParkIntel models on leak-resistant fixed data.
"""

from __future__ import annotations

import json
import os
import pickle
from pathlib import Path

HERE = Path(__file__).resolve().parent
os.environ.setdefault("MPLCONFIGDIR", str(HERE / ".matplotlib-cache"))

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
from sklearn.utils.class_weight import compute_class_weight


DATA_DIR = HERE / "data" / "processed"
MODEL_DIR = HERE / "models"
MODEL_DIR.mkdir(exist_ok=True)

CLASS_NAMES = ["LOW", "MEDIUM", "HIGH"]
TARGET = "risk_encoded"

NUMERIC_FEATURES = [
    "zone_lat",
    "zone_lon",
    "hour",
    "day_of_week",
    "month",
    "is_weekend",
    "hour_sin",
    "hour_cos",
    "dow_sin",
    "dow_cos",
    "junction_flag",
    "violations_last_1h",
    "violations_last_3h",
    "violations_last_24h",
    "violations_last_7d",
    "active_hours_last_7d",
    "repeat_hotspot_score",
    "zone_total_violations",
    "historical_zone_log_total",
    "same_hour_prev_day",
    "same_hour_prev_week",
    "prev_wrong_parking_count",
    "prev_no_parking_count",
    "prev_main_road_count",
    "prev_double_parking_count",
    "prev_near_crossing_count",
    "prev_near_signal_count",
    "prev_footpath_count",
    "prev_heavy_vehicle_count",
    "prev_medium_vehicle_count",
    "prev_light_vehicle_count",
    "prev_two_wheel_count",
    "prev_avg_vio_severity",
    "prev_max_vio_severity",
    "prev_avg_veh_weight",
    "prev_avg_confidence",
]
CAT_FEATURES = ["police_station"]


def load_data() -> tuple[pd.DataFrame, pd.DataFrame, pd.DataFrame]:
    paths = [DATA_DIR / "train_fixed.csv", DATA_DIR / "val_fixed.csv", DATA_DIR / "test_fixed.csv"]
    missing = [p for p in paths if not p.exists()]
    if missing:
        raise FileNotFoundError(f"Missing fixed data files: {missing}. Run preprocess_fixed.py first.")
    return (pd.read_csv(paths[0]), pd.read_csv(paths[1]), pd.read_csv(paths[2]))


def encode(train: pd.DataFrame, val: pd.DataFrame, test: pd.DataFrame):
    numeric = [c for c in NUMERIC_FEATURES if c in train.columns]
    categorical = [c for c in CAT_FEATURES if c in train.columns]

    enc = OrdinalEncoder(handle_unknown="use_encoded_value", unknown_value=-1)
    if categorical:
        enc.fit(train[categorical].fillna("UNKNOWN"))
        for frame in (train, val, test):
            encoded = enc.transform(frame[categorical].fillna("UNKNOWN"))
            for idx, col in enumerate(categorical):
                frame[f"{col}_enc"] = encoded[:, idx]

    feature_cols = numeric + [f"{c}_enc" for c in categorical]
    X_train = train[feature_cols].fillna(0).astype("float32")
    X_val = val[feature_cols].fillna(0).astype("float32")
    X_test = test[feature_cols].fillna(0).astype("float32")
    y_train = train[TARGET].astype(int)
    y_val = val[TARGET].astype(int)
    y_test = test[TARGET].astype(int)
    return X_train, y_train, X_val, y_val, X_test, y_test, feature_cols, enc


def metrics(name: str, y_true, preds, proba=None) -> dict:
    result = {
        "Model": name,
        "Accuracy": accuracy_score(y_true, preds),
        "Macro-F1": f1_score(y_true, preds, average="macro"),
        "HIGH-F1": f1_score(y_true, preds, labels=[2], average="macro"),
    }
    print(f"\n{name}")
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


def topk_binary_report(
    y_true: np.ndarray,
    scores: np.ndarray,
    k_values=(10, 25, 50, 100, 250, 500, 1000),
    label: str = "Global",
) -> list[dict]:
    order = np.argsort(scores)[::-1]
    total_high = int(y_true.sum())
    rows = []
    print(f"\n{label} HIGH ranking report")
    print(f"Actual HIGH rows: {total_high}")
    for k in k_values:
        top = order[: min(k, len(order))]
        found = int(y_true[top].sum())
        precision = found / max(len(top), 1)
        recall = found / max(total_high, 1)
        lift = precision / max(total_high / max(len(y_true), 1), 1e-12)
        rows.append(
            {
                "scope": label,
                "k": k,
                "precision_at_k": precision,
                "recall_at_k": recall,
                "lift": lift,
                "found_high": found,
                "total_high": total_high,
            }
        )
        print(
            f"  Top-{k:<4} precision={precision:.4f} "
            f"recall={recall:.4f} lift={lift:.1f} found={found}"
        )
    return rows


def grouped_topk_report(
    frame: pd.DataFrame,
    y_true: np.ndarray,
    scores: np.ndarray,
    group_col: str,
    k: int = 10,
    min_high: int = 5,
) -> pd.DataFrame:
    eval_df = frame[[group_col]].copy()
    eval_df["is_high"] = y_true
    eval_df["score"] = scores
    rows = []
    for group_value, group in eval_df.groupby(group_col):
        total_high = int(group["is_high"].sum())
        if total_high < min_high:
            continue
        top = group.nlargest(min(k, len(group)), "score")
        found = int(top["is_high"].sum())
        rows.append(
            {
                group_col: group_value,
                "rows": len(group),
                "total_high": total_high,
                f"precision_at_{k}": found / max(len(top), 1),
                f"recall_at_{k}": found / total_high,
                "found_high": found,
            }
        )
    result = pd.DataFrame(rows)
    if result.empty:
        return result
    return result.sort_values([f"precision_at_{k}", "total_high"], ascending=[False, False])


def threshold_report(y_true: np.ndarray, scores: np.ndarray) -> dict:
    precision, recall, thresholds = precision_recall_curve(y_true, scores)
    f1 = 2 * precision * recall / (precision + recall + 1e-12)
    best_idx = int(np.nanargmax(f1[:-1])) if len(thresholds) else 0
    best = {
        "threshold": float(thresholds[best_idx]) if len(thresholds) else 0.5,
        "precision": float(precision[best_idx]),
        "recall": float(recall[best_idx]),
        "f1": float(f1[best_idx]),
    }
    print(
        "\nBest binary threshold by HIGH-F1: "
        f"{best['threshold']:.4f} "
        f"(precision={best['precision']:.4f}, recall={best['recall']:.4f}, f1={best['f1']:.4f})"
    )
    return best


def main() -> None:
    train, val, test = load_data()
    print(f"Train {train.shape} | Val {val.shape} | Test {test.shape}")
    print("Train class distribution:")
    print(train["hotspot_risk"].value_counts().to_string())

    X_train, y_train, X_val, y_val, X_test, y_test, feature_cols, encoder = encode(train, val, test)
    classes = np.array([0, 1, 2])
    weights = compute_class_weight(class_weight="balanced", classes=classes, y=y_train)
    class_weight = {int(cls): float(weight) for cls, weight in zip(classes, weights)}
    sample_weight = y_train.map(class_weight).to_numpy()
    print(f"Feature count: {len(feature_cols)}")
    print(f"Class weights: {class_weight}")

    rf = RandomForestClassifier(
        n_estimators=250,
        max_depth=16,
        min_samples_leaf=5,
        class_weight=class_weight,
        n_jobs=-1,
        random_state=42,
    )
    rf.fit(X_train, y_train)
    rf_proba = rf.predict_proba(X_test)
    rf_result = metrics("RandomForest", y_test, np.argmax(rf_proba, axis=1), rf_proba)

    lgb_model = lgb.LGBMClassifier(
        objective="multiclass",
        num_class=3,
        metric="multi_logloss",
        learning_rate=0.04,
        n_estimators=2500,
        num_leaves=96,
        max_depth=10,
        min_child_samples=40,
        subsample=0.85,
        subsample_freq=1,
        colsample_bytree=0.85,
        reg_alpha=0.1,
        reg_lambda=0.2,
        class_weight=class_weight,
        random_state=42,
        verbose=-1,
    )
    lgb_model.fit(
        X_train,
        y_train,
        eval_set=[(X_val, y_val)],
        callbacks=[lgb.early_stopping(100, verbose=False), lgb.log_evaluation(100)],
    )
    lgb_proba = lgb_model.predict_proba(X_test)
    lgb_result = metrics("LightGBM", y_test, np.argmax(lgb_proba, axis=1), lgb_proba)

    xgb_model = xgb.XGBClassifier(
        objective="multi:softprob",
        num_class=3,
        eval_metric="mlogloss",
        learning_rate=0.05,
        max_depth=7,
        min_child_weight=3,
        subsample=0.85,
        colsample_bytree=0.85,
        gamma=0.1,
        reg_alpha=0.1,
        reg_lambda=1.0,
        n_estimators=1500,
        early_stopping_rounds=75,
        tree_method="hist",
        random_state=42,
        verbosity=1,
    )
    xgb_model.fit(
        X_train,
        y_train,
        sample_weight=sample_weight,
        eval_set=[(X_val, y_val)],
        verbose=100,
    )
    xgb_proba = xgb_model.predict_proba(X_test)
    xgb_result = metrics("XGBoost", y_test, np.argmax(xgb_proba, axis=1), xgb_proba)

    print("\nTraining binary HIGH-vs-rest ranker...")
    y_train_high = (y_train == 2).astype(int)
    y_val_high = (y_val == 2).astype(int)
    y_test_high = (y_test == 2).astype(int)
    pos = int(y_train_high.sum())
    neg = int(len(y_train_high) - pos)
    scale_pos_weight = neg / max(pos, 1)
    print(f"Binary train positives={pos:,}, negatives={neg:,}, scale_pos_weight={scale_pos_weight:.2f}")

    high_ranker = lgb.LGBMClassifier(
        objective="binary",
        metric="auc",
        learning_rate=0.03,
        n_estimators=3000,
        num_leaves=64,
        max_depth=8,
        min_child_samples=60,
        subsample=0.85,
        subsample_freq=1,
        colsample_bytree=0.85,
        reg_alpha=0.2,
        reg_lambda=0.5,
        scale_pos_weight=scale_pos_weight,
        random_state=42,
        verbose=-1,
    )
    high_ranker.fit(
        X_train,
        y_train_high,
        eval_set=[(X_val, y_val_high)],
        callbacks=[lgb.early_stopping(150, verbose=False), lgb.log_evaluation(100)],
    )
    high_scores = high_ranker.predict_proba(X_test)[:, 1]
    ap = average_precision_score(y_test_high, high_scores)
    auc = roc_auc_score(y_test_high, high_scores)
    best_threshold = threshold_report(y_test_high.to_numpy(), high_scores)
    ranking_rows = topk_binary_report(y_test_high.to_numpy(), high_scores)
    print(f"Binary HIGH ranker Average Precision={ap:.4f} | ROC-AUC={auc:.4f}")

    hourly_topk = grouped_topk_report(test, y_test_high.to_numpy(), high_scores, "hour", k=10)
    station_topk = grouped_topk_report(test, y_test_high.to_numpy(), high_scores, "police_station", k=10)
    print("\nBest hourly Precision@10 groups:")
    print(hourly_topk.head(12).to_string(index=False) if not hourly_topk.empty else "No groups with enough HIGH rows.")
    print("\nBest station Precision@10 groups:")
    print(station_topk.head(12).to_string(index=False) if not station_topk.empty else "No groups with enough HIGH rows.")

    results = pd.DataFrame([rf_result, lgb_result, xgb_result]).sort_values("HIGH-F1", ascending=False)
    print("\nSummary:")
    print(results.to_string(index=False))

    primary = "lgb"
    if results.iloc[0]["Model"] == "XGBoost":
        primary = "xgb"
    elif results.iloc[0]["Model"] == "RandomForest":
        primary = "rf"

    lgb_model.booster_.save_model(str(MODEL_DIR / "lgb_hotspot_model.txt"))
    high_ranker.booster_.save_model(str(MODEL_DIR / "lgb_high_ranker.txt"))
    xgb_model.save_model(str(MODEL_DIR / "xgb_hotspot_model.json"))
    with open(MODEL_DIR / "rf_baseline.pkl", "wb") as f:
        pickle.dump(rf, f)
    with open(MODEL_DIR / "police_station_encoder.pkl", "wb") as f:
        pickle.dump(encoder, f)

    meta = {
        "feature_cols": feature_cols,
        "class_names": CLASS_NAMES,
        "primary_model": "lgb_high_ranker",
        "multiclass_comparison_primary": primary,
        "target": {
            "LOW": "0 next-hour violations",
            "MEDIUM": "1-5 next-hour violations",
            "HIGH": ">=6 next-hour violations",
        },
        "validation": "chronological 70/15/15 split; early stopping uses validation, not test",
        "test_metrics": results.to_dict(orient="records"),
        "binary_high_ranker": {
            "average_precision": float(ap),
            "roc_auc": float(auc),
            "best_threshold": best_threshold,
            "topk": ranking_rows,
        },
    }
    (MODEL_DIR / "model_meta.json").write_text(json.dumps(meta, indent=2))

    scored = test[["zone_id", "zone_lat", "zone_lon", "hour_bin", "police_station", "hotspot_risk", "violations_next_1h"]].copy()
    scored["lgb_pred"] = np.argmax(lgb_proba, axis=1)
    scored["high_prob"] = lgb_proba[:, 2]
    scored["high_rank_score"] = high_scores
    scored["high_rank_pred"] = (high_scores >= best_threshold["threshold"]).astype(int)
    scored.to_csv(HERE / "predictions_fixed.csv", index=False)
    pd.DataFrame(ranking_rows).to_csv(HERE / "high_ranker_topk.csv", index=False)
    hourly_topk.to_csv(HERE / "high_ranker_topk_by_hour.csv", index=False)
    station_topk.to_csv(HERE / "high_ranker_topk_by_station.csv", index=False)
    print(f"\nSaved models and metadata to {MODEL_DIR}")
    print("Primary deployment model: lgb_high_ranker")
    print(f"Best multiclass comparison model by HIGH-F1: {primary}")


if __name__ == "__main__":
    main()
