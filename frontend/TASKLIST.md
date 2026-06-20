# ParkIntel Frontend Tasklist

Use this file to track frontend implementation progress.

Status legend:

- `[ ]` Not started
- `[~]` In progress
- `[x]` Complete
- `[!]` Blocked

## Phase 0: Decisions And Inputs

- [x] Confirm frontend should be scaffolded from scratch inside `frontend`. (Yes, scaffolded and built)
- [x] Confirm local backend base URL. (Yes, defaults to http://localhost:8080)
- [x] Provide or defer Mapbox token. (Yes, deferred with dark placeholder fallback)
- [x] Confirm default map center and zoom. (Yes, set to Bengaluru [77.5946, 12.9716], Zoom 12)
- [x] Confirm whether to use backend `priority_score` as the frontend priority field. (Yes, priority_score is utilized)
- [ ] Decide how to handle missing `gated` field for enforcement rows. (Pending user input)
- [ ] Decide how to handle missing ONNX latency, CPU, and memory telemetry. (Pending user input)
- [x] Confirm simulation v1 should match backend request shape. (Yes, simulated using violation_reduction_percent)
- [x] Choose single-page command center or multiple routes. (Yes, single-page command center page implemented)
- [x] Decide whether mock data fallback is allowed. (Yes, standard SWR fetch and loading/empty fallback styling is implemented)
- [ ] Decide whether filters should sync to the URL. (Pending user input)
- [ ] Confirm responsive target for v1. (Pending user input)

## Phase 1: Project Scaffold

- [x] Create Next.js 14 App Router app in `frontend`.
- [x] Add TypeScript configuration.
- [x] Add Tailwind CSS configuration.
- [x] Add PostCSS configuration.
- [x] Add ESLint configuration.
- [x] Add `package.json` scripts.
- [x] Add `.env.local.example`.
- [x] Install core dependencies.
- [x] Verify app starts locally.

## Phase 2: Design System

- [x] Add Aurelian color tokens to Tailwind.
- [x] Add Geist Sans and Geist Mono.
- [x] Create global CSS base styles.
- [x] Add radial background treatment.
- [x] Add reusable `cn()` utility.
- [x] Add priority color helper.
- [x] Add numeric/date formatting helpers.

## Phase 3: Base UI Components

- [x] Build `Button`.
- [x] Build `Badge`.
- [x] Build `Input`.
- [x] Build `Slider`.
- [x] Build `Panel`.
- [x] Build `MetricCard`.
- [x] Build `Skeleton`.
- [x] Build `EmptyState`.
- [x] Build `OfflineOverlay`.

## Phase 4: API Layer

- [x] Define API response TypeScript types.
- [x] Build API client with base URL support.
- [x] Add health endpoint client.
- [x] Add hotspots endpoint client.
- [x] Add ranking endpoint client.
- [x] Add zone insights endpoint client.
- [x] Add simulation endpoint client.
- [x] Add API error normalization.

## Phase 5: Data Hooks

- [x] Add `useHealth`.
- [x] Add `useHotspots`.
- [x] Add `useRanking`.
- [x] Add `useZoneInsights`.
- [x] Add `useSimulation`.
- [x] Configure 30 second SWR polling.
- [x] Add retry behavior.
- [x] Add loading and empty-state conventions.

## Phase 6: Global State

- [x] Create Zustand command store.
- [x] Store selected zone.
- [x] Store selected hour.
- [x] Store selected date.
- [x] Store police station filter.
- [x] Store risk level filter.
- [x] Store ranking limit.
- [x] Store simulation settings.
- [x] Store forecast mode state.

## Phase 7: Command Shell

- [x] Build root layout.
- [x] Build command-center page.
- [x] Build top status/navigation bar.
- [x] Add filter controls.
- [x] Add backend health indicator.
- [x] Add responsive panel layout.
- [x] Add global offline overlay.

## Phase 8: Map

- [x] Add Mapbox GL dependency and styles.
- [x] Initialize map with pitch `45`.
- [x] Initialize map with bearing `-17`.
- [x] Add missing-token fallback.
- [x] Convert hotspots to GeoJSON.
- [x] Render hotspot circle layer.
- [x] Scale radius by `priority_score`.
- [x] Color hotspots by priority/risk.
- [x] Add selected-zone visual state.
- [x] Add click-to-select behavior.
- [~] Add popup or compact detail preview.

## Phase 9: Telemetry Panel

- [x] Render backend status.
- [x] Render last refresh time.
- [x] Render selected filters.
- [x] Render hotspot count.
- [x] Render ranking count.
- [x] Add placeholders for unavailable ONNX latency, CPU, and memory metrics.
- [!] Replace placeholders if backend telemetry becomes available.

## Phase 10: Enforcement Ranking

- [x] Build ranking panel shell.
- [x] Build TanStack table columns.
- [x] Render rank.
- [x] Render zone ID.
- [x] Render police station.
- [x] Render junction.
- [x] Render expected violations.
- [x] Render impact score.
- [x] Render priority score.
- [x] Render priority level.
- [x] Render model confidence.
- [x] Render recommended action.
- [x] Highlight rows where `priority_score >= 72.0`.
- [x] Add row click to select zone.
- [x] Add loading state.
- [x] Add error state.
- [x] Add empty state.
- [!] Add Gate icon once backend exposes `gated`.

## Phase 11: Zone Insights

- [x] Build selected-zone insights panel.
- [x] Render no-zone-selected state.
- [x] Render priority score indicator.
- [x] Render impact score indicator.
- [x] Render predicted hotspot risk.
- [x] Render model confidence.
- [x] Render recommended action.
- [x] Render LOW/MEDIUM/HIGH probability chart.
- [x] Render zone stats.
- [x] Render top violation types.
- [x] Render top vehicle types.
- [x] Render Logic Transparency reasons box.
- [x] Add loading state.
- [x] Add error state.
- [!] Add ensemble-weight chart if backend exposes LGB/XGB/RF weights.

## Phase 12: Policy Simulation

- [x] Build simulation panel shell.
- [x] Require selected zone before simulation.
- [x] Add violation reduction slider.
- [x] Add submit action.
- [x] POST simulation request.
- [x] Render current priority score and level.
- [x] Render simulated priority score and level.
- [x] Render priority change.
- [x] Render estimated impact reduction.
- [x] Render backend note.
- [x] Add Forecast Mode badge.
- [x] Add Forecast Mode border styling.
- [x] Add reset simulation action.
- [!] Add broader client-side formula preview if backend exposes required feature inputs.

## Phase 13: Polish And Accessibility

- [ ] Add keyboard focus states.
- [ ] Add accessible labels for icon-only controls.
- [ ] Add reduced-motion-friendly transitions where appropriate.
- [ ] Verify text contrast.
- [ ] Verify no panel clipping or overlap.
- [ ] Verify desktop layout.
- [ ] Verify tablet layout if in v1 scope.
- [ ] Verify mobile layout if in v1 scope.

## Phase 14: Verification

- [x] Run lint. (Yes, runs with 0 errors)
- [x] Run typecheck. (Yes, passes successfully)
- [x] Run production build. (Yes, compiles and builds successfully)
- [ ] Test backend offline behavior.
- [ ] Test health polling.
- [ ] Test hotspots request with valid hour.
- [ ] Test ranking request with valid hour.
- [ ] Test zone insights request.
- [ ] Test simulation request.
- [ ] Test Mapbox rendering with token.
- [ ] Test Mapbox fallback without token.
- [ ] Test severity highlighting at `priority_score >= 72.0`.
- [ ] Test selected-zone flow from map.
- [ ] Test selected-zone flow from ranking table.

## Phase 15: Final Documentation

- [ ] Document local setup.
- [ ] Document required env vars.
- [ ] Document backend dependency.
- [ ] Document known API mismatches.
- [ ] Document future backend enhancements needed by PRD.
- [ ] Add screenshots or usage notes if requested.

