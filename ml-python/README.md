# ParkIntel ML Pipeline

This is the canonical model folder. The old `model-r2 (2)` folder has been removed.

## Layout

- `data/raw/grid-r2-t1-dataset.csv` - source violation dataset
- `data/processed/` - generated train/validation/test CSVs
- `models/` - trained model artifacts and ONNX exports
- `preprocess_fixed.py` - leak-resistant preprocessing
- `train_fixed.py` - chronological train/validation/test training
- `export_onnx.py` - model export for backend inference

## Setup

```bash
cd ml-python
python3 -m venv .venv
.venv/bin/pip install -r requirements.txt
```

On macOS, LightGBM and XGBoost need OpenMP:

```bash
brew install libomp
```

## Run

```bash
cd ml-python
.venv/bin/python preprocess_fixed.py
.venv/bin/python train_fixed.py
```

If `data/processed/*.csv` already exists, you can skip preprocessing and run only `train_fixed.py`.

## What To Read

The corrected data is extremely imbalanced because most zone-hours have no HIGH hotspot. Plain accuracy is therefore not the main success metric.

Use these outputs for enforcement usefulness:

- `models/lgb_high_ranker.txt` - primary HIGH-vs-rest ranking model
- `predictions_fixed.csv` - test predictions with `high_rank_score`
- `high_ranker_topk.csv` - global Precision@K/Recall@K
- `high_ranker_topk_by_hour.csv` - hourly Top-K results
- `high_ranker_topk_by_station.csv` - police-station Top-K results
