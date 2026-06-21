# ParkIntel Go Inference Service Developer Guide

This directory contains the source code for the **ParkIntel Go REST API Inference Service** (`backend/`), which serves pre-computed and live-inferred illegal parking hotspot predictions. The service loads three pre-trained ONNX models once at startup, replicates the Python scoring layer, and integrates with a PostgreSQL database.

---

## 🏗️ Architecture & Component Layout

The backend service is structured into modular layers, separating configuration, database access, model inference, and HTTP routing:

```
backend/
├── main.go               # Service entry point; orchestrates lifecycle & routing
├── go.mod / go.sum       # Go module and dependency declarations
├── schema.sql            # PostgreSQL DDL table and index declarations
├── config/
│   └── config.go         # Environment variable loader using godotenv
├── db/
│   └── db.go             # PostgreSQL connection pool and CSV data-ingestion
├── inference/
│   ├── session.go        # Thread-safe ONNX runtime session manager
│   ├── features.go       # Feature input vector builder and police station encoder
│   └── scoring.go        # Mathematical scoring layer (Impact, Priority, Confidence)
├── handlers/
│   ├── init.go           # Handler package initialization
│   ├── hotspots.go       # GET /api/hotspots (Pre-computed hotspots)
│   ├── ranking.go        # GET /api/enforcement/ranking (Ranked list)
│   ├── zones.go          # GET /api/zones/:zone_id/insights (Live ONNX Inference)
│   └── simulate.go       # POST /api/simulate (Priority score simulation)
├── middleware/
│   └── cors.go           # Gin CORS configuration middleware
└── cache/
    └── cache.go          # Thread-safe in-memory cache with TTL verification
```

---

## 🔄 Detailed Data Flows

### 1. Startup & Initialization Sequence
When the application starts, it performs a strict sequence of checks and setups before booting the HTTP server:

```
[Main Entry (main.go)]
  │
  ├──► 1. Load Configurations (.env)
  │
  ├──► 2. Initialize DB Connection Pool (db.go)
  │
  ├──► 3. Initialize DB Schema (schema.sql)
  │
  ├──► 4. Initialize ONNX Runtime & Load Sessions (session.go)
  │      ├── Loads lgb, xgb, and rf sessions once into memory
  │      └── Pre-allocates fixed-shape input/output tensors
  │
  ├──► 5. Auto-Ingest CSV Data (if database is empty)
  │      ├── Step A: Parse train.csv -> Compute historical zone/hour/day averages
  │      ├── Step B: Parse test.csv -> Insert into zone_time_features
  │      └── Step C: Parse predictions.csv -> Insert into zone_predictions
  │
  └──► 6. Boot HTTP Engine & Register Routes (Gin Router)
```

### 2. Live Inference (`GET /api/zones/:zone_id/insights`)
This endpoint runs live model inference utilizing the pre-allocated ONNX sessions:

```
[Request: GET /api/zones/:zone_id/insights]
  │
  ├──► 1. Fetch feature records from 'zone_time_features' by zone_id + hour
  │
  ├──► 2. Construct 'FeatureInput' struct
  │
  ├──► 3. Map features dynamically to 'FeatureCols' list in onnx_meta.json [1, 28]
  │
  ├──► 4. Acquire Mutex Lock for the target Model Session
  │      ├── Write feature slice directly into input tensor memory (GetData())
  │      ├── Execute session.Run()
  │      └── Extract predictions from label & probability tensors (GetData())
  │
  ├──► 5. Release Mutex Lock
  │
  ├──► 6. Feed outputs to the Scoring Layer
  │      ├── Compute ImpactScore()
  │      ├── Compute PriorityScore()
  │      ├── Compute ModelConfidence()
  │      └── Generate Recommendations & Reasons
  │
  └──► 7. Return JSON response
```

### 3. Simulation Flow (`POST /api/simulate`)
This endpoint performs rule-based priority score simulations without running the ONNX models:

```
[Request: POST /api/simulate]
  │
  ├──► 1. Fetch current priority score from 'zone_predictions'
  │
  ├──► 2. Fetch original features from 'zone_time_features'
  │
  ├──► 3. Scale all violation counts by (1 - reduction_percent/100)
  │
  ├──► 4. Recompute ImpactScore() with scaled values
  │
  ├──► 5. Recompute PriorityScore() keeping ML probabilities unchanged
  │
  └──► 6. Return Simulated JSON response
```

---

## 🧮 Scoring Layer Formulas

The service replicates the Python scoring layer exactly using these formulas in `inference/scoring.go`:

### Impact Score
An absolute-count severity formula representing illegal parking congestion impact (capped at 100):
$$\text{Impact} = 30 \times \text{DoubleParking} + 25 \times \text{MainRoad} + 25 \times \text{NearCrossing} + 25 \times \text{NearSignal} + 8 \times \text{NoParking} + 5 \times \text{WrongParking} + 10 \times \text{HeavyVehicle} + 15 \times \text{JunctionFlag} + 0.4 \times \text{RepeatHotspot}$$

### Priority Score
A blend of ML model high-risk probability and impact score:
$$\text{Priority} = 0.25 \times (P(\text{HIGH}) \times 100) + 0.75 \times \text{Impact}$$
* **Severity Gate**: If the zone contains 2 or more high-severity violations (Double Parking, Main Road, Near Crossing, Near Signal, or Heavy Vehicle at a Junction) and the priority score is under 72.0, it is automatically floor-gated to a minimum of **72.0** (putting it in the HIGH priority tier).

### Model Confidence
Calculated based on the margin between the top two class probabilities:
- **LOW**: $\text{Margin} < 0.10$
- **MEDIUM**: $0.10 \le \text{Margin} < 0.20$
- **HIGH**: $\text{Margin} \ge 0.20$

---

## ⚙️ Setup & Execution

You can run the backend service either inside a standalone Docker container, as part of the multi-container Docker Compose stack, or locally on your host machine.

### Option 1: Running standalone via Docker

The backend service is containerized using a multi-stage Docker build that downloads the correct C++ ONNX Runtime library (`libonnxruntime.so`) dynamically.

1. **Build the backend image**:
   From the project root directory, run:
   ```bash
   docker build -t parkintel-backend -f backend/Dockerfile ./backend
   ```

2. **Run the container**:
   Ensure your PostgreSQL database is running, then start the backend container. You must mount the `models/onnx` folder and the `ml-python` datasets folder into the container for model inference and auto-ingestion to work:
   ```bash
   docker run --name parkintel-backend-container \
     -p 8080:8080 \
     -e DATABASE_URL=postgres://postgres:postgres@host.docker.internal:5432/parkintel?sslmode=disable \
     -e ONNX_MODEL_DIR=/app/models/onnx \
     -e ONNX_RUNTIME_LIB_PATH=/usr/lib/libonnxruntime.so.1.27.0 \
     -v $(pwd)/models/onnx:/app/models/onnx \
     -v $(pwd)/ml-python:/app/ml-python \
     -d parkintel-backend
   ```
   *(Note: Use `host.docker.internal` on macOS/Windows, or your host IP address on Linux, to refer to a PostgreSQL instance running locally on your host machine.)*

### Option 2: Running via Docker Compose (Recommended)

From the project root directory, simply boot the entire stack (Database, Backend, and Frontend):
```bash
docker compose up --build -d
```
The Docker Compose setup mounts all necessary folders and variables automatically. Refer to the root [README.md](file:///home/gagan-ahlawat/Documents/ParkIntel/README.md) for details.

### Option 3: Traditional Local Execution

#### Prerequisites
- **Go**: Version 1.25+
- **PostgreSQL**: Running instance on port `5432`
- **ONNX Runtime Shared Library**: The Go library uses CGO to load `libonnxruntime.so`. Production containers bake ONNX Runtime into `/usr/lib`; local development can either install the system library or point `ONNX_RUNTIME_LIB_PATH` to the actual library installed by the Python `onnxruntime` package.

1. **Run PostgreSQL Database**:
   ```bash
   docker run --name parkintel-postgres -e POSTGRES_DB=parkintel -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:16-alpine
   ```

2. **Configure Environment Variables**:
   Create a `.env` file in the `backend/` directory:
   ```env
   DATABASE_URL=postgres://postgres:postgres@localhost:5432/parkintel?sslmode=disable
   ONNX_MODEL_DIR=../models/onnx
   PORT=8080
   GIN_MODE=release
   ONNX_RUNTIME_LIB_PATH=../ml-python/.venv/lib/python3.12/site-packages/onnxruntime/capi/libonnxruntime.so.1.27.0
   ```

3. **Build & Run**:
   Run the Go service from the `backend/` directory:
   ```bash
   # Fetch and tidy packages
   go mod tidy

   # Run tests
   go test -v ./...

   # Start service
   go run main.go
   ```

The service will automatically:
- Load the ONNX models.
- Run schema migrations.
- Ingest data from CSV files.
- Bind the server to `:8080`.

---

## 🛠️ Troubleshooting Large ONNX Model HTTP Pushes
If pushing the large ONNX binary models (~82MB uncompressed) over HTTPS results in connection drops (`HTTP 408` timeouts or `send-pack: unexpected disconnect`), optimize your local Git parameters:

```bash
# Increase HTTP post buffer to 500MB
git config http.postBuffer 524288000

# Remove low-speed limits and timeout limits
git config http.lowSpeedLimit 0
git config http.lowSpeedTime 999999

# Force HTTP/1.1 to prevent HTTP/2 framing disconnections
git config http.version HTTP/1.1

# Disable pack compression to prevent CPU bottlenecks during serialization
git config core.compression 0
```
