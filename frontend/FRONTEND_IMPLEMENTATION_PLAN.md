# ParkIntel Frontend Implementation Plan

This document captures what is needed to build the ParkIntel Aurelian Command Center frontend described in `PRD.md`.

The current `frontend` folder is a greenfield workspace. At the time this plan was created, it only contained the PRD and no existing Next.js application files.

## Product Goal

Build a high-fidelity, minimalist, real-time command-and-control frontend for the ParkIntel Go inference service.

The frontend should let an operator:

- View illegal parking hotspots on an immersive dark Mapbox map.
- Monitor enforcement priority rankings.
- Drill into zone-level model insights.
- Run policy simulations for selected zones.
- Understand when the backend is offline or data is unavailable.

## Required Stack

The PRD specifies:

- Next.js 14.2+ with App Router.
- TypeScript.
- Tailwind CSS.
- `tailwind-merge` for safe dynamic class composition.
- SWR for polling and data fetching.
- Zustand for global UI state.
- Recharts or Chart.js for charts.
- Mapbox GL JS for map rendering.
- TanStack Table for enforcement ranking.

Recommended additional packages:

- `clsx` for conditional class names.
- `lucide-react` for command-center icons.
- `@tanstack/react-table` for table logic.
- A virtualization package only if the ranking list grows beyond the backend limit of 100 rows.

## Environment Variables

Create `frontend/.env.local.example` with:

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_MAPBOX_TOKEN=
```

Required decisions:

- Confirm the backend local URL and port.
- Provide a Mapbox token, or approve a non-map fallback for local development.

## Backend API Contracts

The PRD lists the intended API surface, and the backend currently exposes matching routes. Some field names differ from the PRD, so the frontend should use the actual backend contracts unless the backend is updated.

### Health

Endpoint:

```http
GET /health
```

Purpose:

- Detect backend availability.
- Drive the "System Offline" overlay.

### Hotspots

Endpoint:

```http
GET /api/hotspots?hour={0-23}&date={YYYY-MM-DD}&police_station={station}&risk_level={risk}
```

Important requirement:

- `hour` is required by the backend.

Current response shape:

```ts
type HotspotsResponse = {
  hotspots: Hotspot[];
  count: number;
  filters_applied: Record<string, unknown>;
};

type Hotspot = {
  zone_id: string;
  lat: number;
  lng: number;
  police_station: string;
  priority_score: number;
  priority_level: string;
  impact_score: number;
  expected_violations: number;
  predicted_hotspot_risk: string;
  model_confidence: string;
};
```

PRD mismatch:

- PRD says the hotspot key field is `priority`.
- Backend returns `priority_score`.
- Frontend should either use `priority_score` or backend should be changed to add `priority`.

### Enforcement Ranking

Endpoint:

```http
GET /api/enforcement/ranking?hour={0-23}&date={YYYY-MM-DD}&limit={number}
```

Important requirements:

- `hour` is required by the backend.
- Backend caps `limit` at 100.

Current response shape:

```ts
type RankingsResponse = {
  rankings: Ranking[];
  total: number;
  hour: number;
};

type Ranking = {
  rank: number;
  zone_id: string;
  police_station: string;
  junction_name: string;
  expected_violations: number;
  impact_score: number;
  priority_score: number;
  priority_level: string;
  model_confidence: string;
  recommended_action: string;
};
```

PRD mismatch:

- PRD requires a `gated: true` flag for manual floor-gate rows.
- Backend does not currently return `gated`.
- Frontend can only show the Gate icon after the backend exposes this field or after we define a reliable derived rule.

### Zone Insights

Endpoint:

```http
GET /api/zones/:zone_id/insights?hour={0-23}
```

Current response shape:

```ts
type ZoneInsightsResponse = {
  zone_id: string;
  predicted_hotspot_risk: string;
  model_confidence: string;
  class_probabilities: {
    LOW: number;
    MEDIUM: number;
    HIGH: number;
  };
  parking_congestion_impact_score: number;
  priority_score: number;
  priority_level: string;
  recommended_action: string;
  reasons: string[];
  note: string;
  zone_stats: {
    total_historical_violations: number;
    repeat_hotspot_score: number;
    top_violation_types: string[];
    top_vehicle_types: string[];
  };
};
```

PRD mismatch:

- PRD mentions `features`, `probabilities`, and `recommendation`.
- Backend returns `class_probabilities`, `recommended_action`, `reasons`, and `zone_stats`.
- PRD mentions ensemble model weights for LGB, XGB, and RF, but backend returns LOW/MEDIUM/HIGH class probabilities.

### Simulation

Endpoint:

```http
POST /api/simulate
```

Current request:

```ts
type SimulateRequest = {
  zone_id: string;
  violation_reduction_percent: number;
};
```

Current response:

```ts
type SimulateResponse = {
  zone_id: string;
  violation_reduction_percent: number;
  current_priority_score: number;
  current_priority_level: string;
  simulated_priority_score: number;
  simulated_priority_level: string;
  priority_change: string;
  estimated_impact_reduction: number;
  note: string;
};
```

PRD mismatch:

- PRD table mentions `reduction_percent` and `patrol_scaling`.
- Backend accepts `violation_reduction_percent` only.
- PRD asks for client-side formula mirroring with multiple feature inputs, but those inputs are not currently exposed to the frontend.

## Design System

Core colors from PRD:

```ts
const colors = {
  background: "#101415",
  surface: "#191c1e",
  gold: "#d4af37",
  borderGold: "rgba(212, 175, 55, 0.15)",
  blue: "#3b82f6",
  emerald: "#10b981",
};
```

Visual rules:

- Use deep slate as the app background.
- Use gold for critical data, active state, selected zone, high priority, and focus rings.
- Use translucent glass panels:
  - `bg-surface/30`
  - `backdrop-blur-2xl`
  - `border border-white/10`
  - `shadow-2xl shadow-black/40`
- Use smooth `duration-300` hover/active transitions.
- Use `scale-95` on active buttons.
- Use mono font for live values to reduce visual jitter.

## Proposed Frontend Structure

```txt
frontend/
  app/
    layout.tsx
    page.tsx
    globals.css
  components/
    command-center/
      command-shell.tsx
      command-map.tsx
      telemetry-panel.tsx
      ranking-panel.tsx
      zone-insights-panel.tsx
      simulation-panel.tsx
      offline-overlay.tsx
    ui/
      badge.tsx
      button.tsx
      input.tsx
      panel.tsx
      skeleton.tsx
      slider.tsx
      metric-card.tsx
  lib/
    api-client.ts
    format.ts
    map.ts
    priority.ts
    utils.ts
  hooks/
    use-health.ts
    use-hotspots.ts
    use-ranking.ts
    use-zone-insights.ts
    use-simulation.ts
  store/
    command-store.ts
  types/
    api.ts
  public/
  .env.local.example
  package.json
  tailwind.config.ts
  tsconfig.json
```

## State Model

Use Zustand for:

- `selectedZoneId`
- `selectedHour`
- `selectedDate`
- `selectedPoliceStation`
- `selectedRiskLevel`
- `rankingLimit`
- `simulationReductionPercent`
- `activeSimulation`
- `forecastMode`

Default values that need confirmation:

- Initial hour.
- Initial date behavior.
- Default ranking limit.
- Whether forecast mode starts disabled.

## Data Fetching

Use SWR with 30 second polling for:

- Backend health.
- Hotspots.
- Ranking.
- Selected zone insights.

Simulation should use an imperative mutation because it is a POST action.

Recommended behavior:

- If health fails, show "System Offline".
- If a panel-specific request fails, show an inline error card and retry action.
- If data is empty, show an empty state instead of an error.

## Main UI Areas

### Command Shell

Responsibilities:

- Own page background and radial gradient.
- Render global nav/status bar.
- Render map and floating panels.
- Coordinate selected zone state.

### Map

Responsibilities:

- Initialize Mapbox GL.
- Use pitch `45` and bearing `-17`.
- Render hotspots as GeoJSON circles.
- Scale circle radius by `priority_score`.
- Color by priority/risk.
- Select zone on click.
- Show active selected-zone visual state.

Blocked by:

- Mapbox token.
- Desired map center and zoom.

### Telemetry Panel

Responsibilities:

- Show backend health.
- Show data refresh state.
- Show selected filters.
- Show ONNX latency, CPU, and memory if backend exposes them.

Blocked by:

- Backend does not currently expose ONNX latency, CPU, or memory metrics.

### Ranking Panel

Responsibilities:

- Render enforcement ranking table.
- Highlight `priority_score >= 72.0`.
- Click a row to select the zone.
- Show rank, zone, station, junction, expected violations, impact, priority, confidence, and action.

Blocked by:

- Gate icon requires backend `gated` field or a confirmed derivation rule.

### Zone Insights Panel

Responsibilities:

- Show selected zone details.
- Render LOW/MEDIUM/HIGH class probabilities.
- Render circular score indicators.
- Render top violation and vehicle types.
- Render reasons in a Logic Transparency box.

Blocked by:

- PRD mentions LGB/XGB/RF ensemble weights, but backend does not expose them.

### Simulation Panel

Responsibilities:

- Let user adjust `violation_reduction_percent`.
- Submit simulation for selected zone.
- Show current vs simulated priority.
- Enable Forecast Mode styling when result exists.

Blocked by:

- PRD asks for broader policy simulation controls than backend currently supports.

## Open Questions

These must be answered before implementation starts.

1. Should the frontend be scaffolded from scratch inside the existing `frontend` folder?
2. What backend base URL should local development use?
3. Do you have a Mapbox token available?
4. What should the default map center and zoom be?
5. Should the app use the backend field `priority_score` as-is, or should the backend be updated to add `priority`?
6. Should the Gate icon be omitted until the backend returns `gated`, or should the backend be updated?
7. Should ONNX latency, CPU, and memory telemetry be added to the backend before the frontend implements that panel?
8. Should simulation v1 use only `violation_reduction_percent`, matching the backend?
9. Should the app be a single command-center route or split into multiple pages?
10. Should the frontend include mock data fallback for development when the backend is offline?
11. Should filter state be shareable in the URL?
12. Should the UI target desktop-first only, or should mobile/tablet be fully supported in v1?

## Recommended Build Order

1. Resolve open questions.
2. Scaffold Next.js app.
3. Configure Tailwind and theme tokens.
4. Build base UI components.
5. Implement typed API client.
6. Implement SWR hooks.
7. Implement Zustand store.
8. Build static command shell.
9. Add backend health and offline overlay.
10. Add ranking panel.
11. Add zone insights panel.
12. Add simulation panel.
13. Add Mapbox map.
14. Wire all panels together through selected zone state.
15. Add polish, loading states, errors, and responsive layout.
16. Run lint, typecheck, and production build.

## Verification Checklist

Before considering the frontend complete:

- `npm run lint` passes.
- `npm run typecheck` passes, if configured separately.
- `npm run build` passes.
- App handles backend offline state.
- Hotspots load for a valid hour.
- Ranking loads for a valid hour.
- Ranking row click selects a zone.
- Zone insights load for a selected zone.
- Simulation works for a selected zone.
- Map renders when token is provided.
- Map fallback renders when token is missing.
- No layout overlap on desktop.
- Critical rows are visually distinct at `priority_score >= 72.0`.

