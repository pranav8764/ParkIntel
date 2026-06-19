"""
convert_to_onnx.py — ParkIntel model → ONNX converter

Converts all three trained models to ONNX format for Go inference.
"""

import json
import joblib
from pathlib import Path

import numpy as np
import lightgbm as lgb
import xgboost as xgb
import onnxruntime as rt
import onnxmltools
from onnxmltools.convert.common.data_types import FloatTensorType as OnnxFloatType
from skl2onnx import convert_sklearn
from skl2onnx.common.data_types import FloatTensorType as SklFloatType

# ── Paths ─────────────────────────────────────────────────────────────────────
MODEL_DIR = Path("models")
OUT_DIR   = Path("models/onnx")
OUT_DIR.mkdir(parents=True, exist_ok=True)

# ── Load metadata ─────────────────────────────────────────────────────────────
with open(MODEL_DIR / "model_meta.json") as f:
    meta = json.load(f)

FEATURE_COLS = meta["feature_cols"]
N_FEATURES   = len(FEATURE_COLS)
N_CLASSES    = len(meta["class_names"])         # 3: LOW, MEDIUM, HIGH
TARGET_OPSET = 12                               # widely supported; works with ORT 1.8+

print(f"Feature count : {N_FEATURES}")
print(f"Classes       : {meta['class_names']}")
print(f"Primary model : {meta.get('primary_model', 'Not Specified')}")
print(f"ONNX opset    : {TARGET_OPSET}\n")

# Dummy batch used for shape validation after conversion
_dummy = np.random.rand(2, N_FEATURES).astype(np.float32)

# ── 1. LightGBM (lgb_hotspot_model.txt → .onnx via onnxmltools) ──────────────
print("── [1/3] Converting LightGBM...")

lgb_booster = lgb.Booster(model_file=str(MODEL_DIR / "lgb_hotspot_model.txt"))

lgb_onnx = onnxmltools.convert_lightgbm(
    lgb_booster,
    initial_types=[("float_input", OnnxFloatType([None, N_FEATURES]))],
    target_opset=TARGET_OPSET
)

lgb_out = OUT_DIR / "lgb_hotspot_model.onnx"
onnxmltools.utils.save_model(lgb_onnx, str(lgb_out))
print(f"   Saved → {lgb_out}")

# ── 2. XGBoost (xgb_hotspot_model.json → .onnx via onnxmltools) ──────────────
print("── [2/3] Converting XGBoost...")

xgb_model = xgb.XGBClassifier()
xgb_model.load_model(str(MODEL_DIR / "xgb_hotspot_model.json"))

# Strip string feature names so onnxmltools gets expected numeric indices
booster = xgb_model.get_booster()
booster.feature_names = None
booster.feature_types = None
xgb_model._Booster = booster  

xgb_onnx = onnxmltools.convert_xgboost(
    xgb_model,
    initial_types=[("float_input", OnnxFloatType([None, N_FEATURES]))],
    target_opset=TARGET_OPSET
)

xgb_out = OUT_DIR / "xgb_hotspot_model.onnx"
onnxmltools.utils.save_model(xgb_onnx, str(xgb_out))
print(f"   Saved → {xgb_out}")

# ── 3. RandomForest (rf_baseline.pkl → .onnx via skl2onnx) ───────────────────
print("── [3/3] Converting RandomForest...")

# FIX: Switch from pickle.load to joblib.load to resolve the UnpicklingError
rf_model = joblib.load(MODEL_DIR / "rf_baseline.pkl")

rf_onnx = convert_sklearn(
    rf_model,
    initial_types=[("float_input", SklFloatType([None, N_FEATURES]))],
    target_opset=TARGET_OPSET,
    options={"zipmap": False}, # Output shape: float32[N, 3] matrix
)

rf_out = OUT_DIR / "rf_baseline.onnx"
with open(rf_out, "wb") as f:
    f.write(rf_onnx.SerializeToString())
print(f"   Saved → {rf_out}")

# ── Validate all three with ONNX Runtime ──────────────────────────────────────
print("\n── Validating with ONNX Runtime...")

MODELS = [
    ("LightGBM",     lgb_out),
    ("XGBoost",      xgb_out),
    ("RandomForest", rf_out),
]

# ── Validate all three with ONNX Runtime ──────────────────────────────────────
print("\n── Validating with ONNX Runtime...")

MODELS = [
    ("LightGBM",     lgb_out),
    ("XGBoost",      xgb_out),
    ("RandomForest", rf_out),
]

for name, path in MODELS:
    sess        = rt.InferenceSession(str(path), providers=["CPUExecutionProvider"])
    input_name  = sess.get_inputs()[0].name
    input_shape = sess.get_inputs()[0].shape

    # Run inference
    outputs     = sess.run(None, {input_name: _dummy})

    label_out = outputs[0]
    proba_out = outputs[1] if len(outputs) > 1 else outputs[0]

    # Handle shape string cleanly whether proba_out is a numpy array or a list of dicts
    if hasattr(proba_out, "shape"):
        proba_str = f"proba={proba_out.shape}"
        sample_str = str(proba_out[0].round(4).tolist())
    else:
        proba_str = f"proba=[List of length {len(proba_out)}]"
        sample_str = str(proba_out[0]) # Print the first dict element

    print(f"   {name:<14} input={input_shape}  labels={label_out.shape}  {proba_str}  ✅")
    print(f"                  sample probabilities: {sample_str}")

# ── Write ONNX metadata sidecar ───────────────────────────────────────────────
onnx_meta = {
    "feature_cols":   FEATURE_COLS,
    "n_features":      N_FEATURES,
    "n_classes":       N_CLASSES,
    "class_names":    meta["class_names"],
    "input_tensor":   "float_input",
    "output_tensors": {
        "labels":        "output_label",       # int64[N]
        "probabilities": "output_probability", # float32[N, 3]
    },
    "opset":          TARGET_OPSET,
    "primary_model":  meta.get("primary_model", "lgb"),
    "models": {
        "lgb": "lgb_hotspot_model.onnx",
        "xgb": "xgb_hotspot_model.onnx",
        "rf":  "rf_baseline.onnx",
    },
}

onnx_meta_path = OUT_DIR / "onnx_meta.json"
with open(onnx_meta_path, "w") as f:
    json.dump(onnx_meta, f, indent=2)

print(f"\n   Metadata → {onnx_meta_path}")
print("\n✅ All conversions complete.")