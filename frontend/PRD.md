# Comprehensive PRD: ParkIntel Aurelian Command Center (Frontend)

## 1. Project Overview
A high-fidelity, luxurious, and minimalist real-time command-and-control interface for the **ParkIntel Go Inference Service**. Built with **Next.js 14 (App Router)** and **Tailwind CSS**, this frontend visualizes illegal parking hotspots, ML-driven zone insights, and policy simulation results using the "Aurelian Command" (Deep Slate & Gold) design language.

---

## 2. Design Foundation: Aurelian Command System

### 2.1 Core Palette
- **Primary Background**: `#101415` (Deep Slate)
- **Primary Accent/Gold**: `#d4af37` (Aurelian Gold) - Used for critical data, active states, and focus elements.
- **Secondary Surface**: `#191c1e` (Muted Slate)
- **Border/Outline**: `rgba(212, 175, 55, 0.15)` (Translucent Gold)
- **Glass Effect**: `rgba(16, 20, 21, 0.7)` with `backdrop-blur-xl`
- **Semantic Colors**:
  - **Critical (High Priority)**: `#d4af37` (Gold)
  - **Moderate**: `#3b82f6` (Synthetic Blue)
  - **Success/Safe**: `#10b981` (Emerald)

### 2.2 Typography (Geist)
- **Display**: `font-display`, `tracking-widest`, `uppercase`, `font-bold`
- **Headline**: `font-headline`, `tracking-tight`, `font-semibold`
- **Body/Monospace**: `font-mono` for all telemetry data (latency, inference scores, coordinates).
- **Scale**:
  - Hero Numbers: `72px / 1`
  - Section Titles: `14px / 1.5` with `0.1em` tracking.

### 2.3 Visual Signatures
- **Glassmorphism**: Floating panels must use `border border-white/5` and `shadow-2xl shadow-black/40`.
- **Gradients**: Subtle radial gradients behind central map components: `radial-gradient(circle, rgba(214,175,55,0.05) 0%, rgba(16,20,21,1) 70%)`.
- **Interactions**: Smooth `duration-300` transitions for all hover states. Buttons should use a slight `scale-95` on active.

---

## 3. Technical Architecture

### 3.1 Stack Requirements
- **Framework**: Next.js 14.2+ (App Router).
- **Styling**: Tailwind CSS + `tailwind-merge` for dynamic classes.
- **Data Fetching**: `SWR` for real-time polling (30s intervals) with optimistic UI updates.
- **State Management**: `Zustand` for global state (active zone, simulation parameters, time-range filters).
- **Charts**: `Recharts` or `Chart.js` styled with custom Gold/Slate theme.
- **Mapping**: `Mapbox GL JS` with a custom style (monochrome dark base, gold road networks).

### 3.2 Component Strategy
- **Base Components**: Foundational UI (Buttons, Badges, Inputs) utilizing Tailwind design tokens.
- **Composite Layouts**:
  - `Shell`: Global navigation + background map layer.
  - `Panel`: Reusable glassmorphic container with optional blur and header.
- **Real-time Telemetry**: Custom hooks (`useInferenceMetrics`) for polling backend endpoints.

---

## 4. Feature & Logic Specifications

### 4.1 Immersive Command Center (Main Map)
- **Interactive Layers**: 
  - `Hotspots`: GeoJSON points with dynamic radius based on `PriorityScore`.
  - `Clusters`: Grouped violations for macro-level monitoring.
- **Telemetry Panel**: Displays real-time data from the Go backend:
  - `ONNX Latency`: MS values.
  - `System Load`: CPU/Memory usage of the Go service.
- **Dynamic Pricing Slider**: Client-side state that previews potential revenue impact based on the `/api/simulate` formula before submission.

### 4.2 Enforcement Ranking (Dashboard)
- **Severity Gating Logic**: 
  - UI must highlight any row where `PriorityScore >= 72.0`.
  - Add a "Gate" icon for rows where the backend applied the manual floor-gate (detected via a `gated: true` flag in API response).
- **Table Virtualization**: Use `tanstack-table` for high-performance rendering of the enforcement queue.

### 4.3 Zone Insights Drill-down
- **Inference Input Visualization**: 
  - Use circular progress indicators for `Spatial Convergence`.
  - Bar charts for `Model Probabilities` (LGB, XGB, RF ensemble weights).
- **Logic Transparency Box**: A dedicated UI section explaining the *Reason* for priority (e.g., "High-severity violations at Junction detected").

### 4.4 Policy Simulation Engine
- **Simulated State Overlay**: When simulation is active, UI should transition to a "Forecast Mode" with a distinct border color or badge.
- **Formula Mirroring**:
  - `Impact = 30*DP + 25*MR + 25*NC + 25*NS + 8*NP + 5*WP + 10*HV + 15*JF + 0.4*RH`
  - `Priority = 0.25*(P(High)*100) + 0.75*Impact`

---

## 5. API Mapping (ParkIntel Go Service)

| Feature | Endpoint | Method | Key Fields |
| :--- | :--- | :--- | :--- |
| **Live Map** | `/api/hotspots` | `GET` | `lat`, `lng`, `priority`, `type` |
| **Ranking** | `/api/enforcement/ranking` | `GET` | `rank`, `zone_id`, `priority`, `impact` |
| **Insights** | `/api/zones/:id/insights` | `GET` | `features`, `probabilities`, `recommendation` |
| **Simulation**| `/api/simulate` | `POST` | `reduction_percent`, `patrol_scaling` |

---

## 6. Implementation Notes for Codex
1. **Glassmorphism Implementation**: Use `bg-surface/30 backdrop-blur-2xl border border-white/10` for all floating panels.
2. **Typography Setup**: Import Geist Sans and Geist Mono. Use Mono specifically for any value that refreshes frequently to prevent layout shift.
3. **Map Initialization**: Ensure Mapbox is initialized with `pitch: 45` and `bearing: -17` for an immersive perspective view.
4. **Error Handling**: Implement a "System Offline" overlay if the Go backend is unreachable.

---
**End of Specification**
