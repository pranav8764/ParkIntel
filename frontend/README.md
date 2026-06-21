# 🛡️ ParkIntel Aurelian Command Center

The interactive, high-fidelity, and real-time command-and-control frontend for the **ParkIntel Go Inference Service**. Built with **Next.js 14+ (App Router)** and **Tailwind CSS**, it visualizes illegal parking hotspots, ML-driven zone insights, and policy simulation results using the premium **Aurelian Command** (Deep Slate & Gold) design system.

---

## 📋 Table of Contents
1. [Product Overview & Goals](#-product-overview--goals)
2. [Design System & Aesthetics](#-design-system--aesthetics)
3. [Technical Architecture & Stack](#-technical-architecture--stack)
4. [Project Structure](#-project-structure)
5. [State Management & Data Flow](#-state-management--data-flow)
6. [API Mapping & PRD Mismatches](#-api-mapping--prd-mismatches)
7. [Environment Variables](#-environment-variables)
8. [Getting Started & Local Setup](#-getting-started--local-setup)
9. [Development Progress & Verification](#-development-progress--verification)

---

## 🎯 Product Overview & Goals

ParkIntel helps traffic police transition from reactive, patrol-based enforcement to proactive, data-driven deployment by:
* **Visualizing city-level hotspots** using an immersive interactive map layer.
* **Providing real-time enforcement rankings** that highlight critical regions (`PriorityScore >= 72.0`).
* **Delivering zone-level insights** including historical trends, top violation types, vehicle profiles, and explainable decision logic.
* **Simulating policy impacts** (e.g., assessing priority shifts when illegal parking is reduced by a target percentage).

### Target Users
* **Traffic Control Room**: Monitors city-wide status and live telemetry.
* **Enforcement Lead / Field Teams**: Receives ranked patrol deployments.
* **Police Station Officers**: Analyzes hot zones within specific jurisdictions.

---

## ✨ Design System & Aesthetics

The application adheres to the **Aurelian Command System** to feel immersive, premium, and responsive:

* **Color Palette**:
  * **Primary Background**: `#101415` (Deep Slate)
  * **Secondary Surface**: `#191c1e` (Muted Slate)
  * **Accent/Gold**: `#d4af37` (Aurelian Gold) — for active elements, highlights, and critical states.
  * **Borders**: Translucent Gold (`rgba(212, 175, 55, 0.15)`) or subtle white (`rgba(255, 255, 255, 0.1)`).
* **Glassmorphism**: Panels utilize a translucent, blurred appearance (`bg-surface/30 backdrop-blur-2xl border border-white/10 shadow-2xl`).
* **Typography**: Integrated **Geist Sans** and **Geist Mono**. Monospace fonts are used for all telemetry and coordinates to prevent layout shifts during updates.
* **Transitions**: Smooth micro-animations (`duration-300` hover transitions and scale adjustments on interaction).

---

## 💻 Technical Architecture & Stack

* **Core Framework**: [Next.js 14+](https://nextjs.org/) (App Router, TypeScript)
* **Styling**: [Tailwind CSS v4](https://tailwindcss.com/) & `tailwind-merge`
* **State Management**: [Zustand](https://github.com/pmndrs/zustand) (single-store model for UI states and filter selections)
* **Data Fetching & Polling**: [SWR](https://swr.vercel.app/) (30-second interval polling with auto-revalidation and optimistic UI support)
* **Charts**: [Recharts](https://recharts.org/) (styled with gold and slate themes)
* **Interactive Mapping**: [Mapbox GL JS v3](https://docs.mapbox.com/mapbox-gl-js/) (rendered at pitch `45` and bearing `-17` for an immersive 3D perspective, with dark monochrome base and gold road networks)
* **Table Performance**: [@tanstack/react-table](https://tanstack.com/table/latest) (handling cell selection, sorting, and conditional styling)

---

## 📂 Project Structure

```
frontend/
├── app/
│   ├── layout.tsx                 # Root layout, injects fonts and global styles
│   ├── page.tsx                   # Main Command Center Dashboard route
│   └── globals.css                # Custom CSS variables, background styling & scrollbars
├── components/
│   ├── command-center/            # Feature-specific dashboard panels
│   │   ├── command-map.tsx        # Mapbox Map with circle layers & selection logic
│   │   ├── command-shell.tsx      # Main layout grid for command center
│   │   ├── filter-controls.tsx    # Station, risk, hour, and date filters
│   │   ├── offline-overlay.tsx    # Blocks UI when Go Backend is unreachable
│   │   ├── ranking-panel.tsx      # Interactive enforcement ranking table
│   │   ├── simulation-panel.tsx   # Policy simulation controls & forecast mode
│   │   ├── telemetry-panel.tsx    # Live counts, latency, and refresh stats
│   │   └── zone-insights-panel.tsx# Model class probabilities and zone attributes
│   └── ui/                        # Reusable component library
│       ├── badge.tsx
│       ├── button.tsx
│       ├── empty-state.tsx
│       ├── input.tsx
│       ├── metric-card.tsx
│       ├── panel.tsx
│       ├── skeleton.tsx
│       └── slider.tsx
├── hooks/                         # Typed SWR fetching wrappers
│   ├── use-health.ts              # Monitors backend connectivity
│   ├── use-hotspots.ts            # Fetches map coordinates & priorities
│   ├── use-ranking.ts             # Fetches ranked zone listings
│   ├── use-simulation.ts          # Sends POST requests to simulation endpoint
│   └── use-zone-insights.ts       # Fetches zone details & insights
├── lib/                           # Utilities & formatting functions
│   ├── api-client.ts              # Base fetcher configuration
│   ├── format.ts                  # Date/time formatting helpers
│   ├── map.ts                     # Mapbox constants & token validators
│   ├── priority.ts                # Priority styling & badge generators
│   └── utils.ts                   # Tailwind merge utility (`cn`)
├── store/
│   └── command-store.ts           # Zustand global state (filters, selected zone, forecast mode)
├── types/
│   └── api.ts                     # TypeScript definitions matching backend contracts
├── public/                        # Static assets
└── package.json                   # Project dependencies and script runner
```

---

## 🔄 State Management & Data Flow

### Global Store
Zustand manages parameters shared across panels:
* `selectedZoneId`: Currently inspected zone.
* `selectedHour`: Hour filter (0-23) for temporal forecasting.
* `selectedDate`: Target date (YYYY-MM-DD).
* `selectedPoliceStation`: Jurisdiction filtering.
* `selectedRiskLevel`: High/Medium/Low filtering.
* `activeSimulation`: Resulting simulated payload.
* `forecastMode`: Boolean indicating that the user is previewing a simulated policy outcome.

### Data Synchronicity
* **Health Check**: Every 30 seconds, `useHealth` pings the backend `/health` endpoint. If this fails, the application is locked behind a glassmorphic **Offline Overlay**.
* **Hotspots & Rankings**: Polls every 30 seconds. Selecting a row in the **Ranking Panel** or clicking a hotspot on the **Command Map** updates the `selectedZoneId`, triggering SWR to load fresh **Zone Insights** immediately.
* **Simulation Execution**: Submitting a reduction percentage sends a POST request. The returned simulated scores are stored, triggering **Forecast Mode** (adding gold borders and overlay indicators).

---

## 🔗 API Mapping & PRD Mismatches

The frontend integrates directly with the Go REST API. The actual backend response contracts slightly differ from the initial [PRD.md](file:///c:/Projects/ParkIntel/frontend/PRD.md) specifications. The frontend uses the actual implementation contracts below:

### 1. Health (`GET /health`)
* **Endpoint**: `/health`
* **Response**: `{"status": "healthy"}`

### 2. Hotspots (`GET /api/hotspots`)
* **Params**: `hour` (required, 0-23), `date` (YYYY-MM-DD), `police_station`, `risk_level`
* **PRD vs Code**: PRD designated the core score key as `priority`. The backend returns `priority_score`, which is used in the frontend.

### 3. Enforcement Ranking (`GET /api/enforcement/ranking`)
* **Params**: `hour` (required, 0-23), `date` (YYYY-MM-DD), `limit` (max 100)
* **PRD vs Code**: The PRD requested a `gated: boolean` flag indicating manual floor-gate application (for `priority_score >= 72.0`). Because the backend database schema doesn't yet include this field, the flag is derived programmatically or deferred.

### 4. Zone Insights (`GET /api/zones/:id/insights`)
* **Params**: `hour` (required, 0-23)
* **PRD vs Code**: The PRD specified LGB, XGB, and RF ensemble model weights. The backend instead returns `class_probabilities` (`LOW`, `MEDIUM`, `HIGH`). The frontend renders these directly as custom bar charts.

### 5. Policy Simulation (`POST /api/simulate`)
* **Payload**: `{"zone_id": string, "violation_reduction_percent": number}`
* **PRD vs Code**: The PRD requested a client-side formula preview and multiple inputs (like patrol scaling). The backend is limited to `violation_reduction_percent` and returns a simplified impact comparison. The simulation panel adheres to this Go service interface.

---

## 🔑 Environment Variables

To configure local development, copy `.env.local.example` to `.env.local`:

```bash
cp .env.local.example .env.local
```

Modify the variables as necessary:
* `NEXT_PUBLIC_API_BASE_URL`: The server URL of your running Go service (e.g., `http://localhost:8080`).
* `NEXT_PUBLIC_MAPBOX_TOKEN`: Your Mapbox GL JS public token. If left blank, the frontend maps will automatically fall back to an elegant, dark placeholder view showing detailed zone coordinates and priority markers.

---

## 🚀 Getting Started & Setup

You can run the frontend either inside a standalone Docker container, as part of the multi-container Docker Compose stack, or locally on your host machine.

### Option 1: Running standalone via Docker

The frontend application can be built and run using Node.js Alpine base image:

1. **Build the frontend image**:
   From the project root directory, run:
   ```bash
   docker build -t parkintel-frontend -f frontend/Dockerfile ./frontend
   ```

2. **Run the container**:
   ```bash
   docker run --name parkintel-frontend-container \
     -p 3000:3000 \
     -e NEXT_PUBLIC_API_BASE_URL=http://localhost:8080 \
     -d parkintel-frontend
   ```
   Now access the frontend at [http://localhost:3000](http://localhost:3000) on your host machine.

### Option 2: Running via Docker Compose (Recommended)

From the project root directory, simply run:
```bash
docker compose up --build -d
```
Refer to the root [README.md](file:///home/gagan-ahlawat/Documents/ParkIntel/README.md) for more details.

### Option 3: Local Host Running

#### Prerequisites
* **Node.js**: Version 18+ or 20+
* **Package Manager**: npm

1. **Install Dependencies**:
   Navigate to the `frontend/` directory and run:
   ```bash
   npm install
   ```

2. **Run the Development Server**:
   ```bash
   npm run dev
   ```
   Open [http://localhost:3000](http://localhost:3000) in your web browser.

3. **Production Build & Validation**:
   ```bash
   # Run lint checks
   npm run lint

   # Build the Next.js bundle
   npm run build
   ```

---

## 🛠️ Development Progress & Verification

The current features correspond directly to the stages in [TASKLIST.md](file:///c:/Projects/ParkIntel/frontend/TASKLIST.md):
* **Design & Tokens**: Fully configured. All components adapt dynamically.
* **Telemetry**: Integrated with status overlays and countdown metrics.
* **Interactive Ranking Table**: Implemented with row click highlights and high-severity row flags (`>= 72.0`).
* **Mapbox integration**: Renders hotspots based on coordinate outputs, falling back gracefully without a token.
* **Error States**: Offline overlays and endpoint failure panels are wired dynamically.

For detailed development logs or to mark off finished items, see [TASKLIST.md](file:///c:/Projects/ParkIntel/frontend/TASKLIST.md) and [FRONTEND_IMPLEMENTATION_PLAN.md](file:///c:/Projects/ParkIntel/frontend/FRONTEND_IMPLEMENTATION_PLAN.md).

---
