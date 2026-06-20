# ParkIntel Frontend Tasklist

Use this file to track frontend implementation progress.

Status legend:

- `[ ]` Not started
- `[~]` In progress
- `[x]` Complete
- `[!]` Blocked

## Phase 0: Decisions And Inputs

- [ ] Confirm frontend should be scaffolded from scratch inside `frontend`.
- [ ] Confirm local backend base URL.
- [ ] Provide or defer Mapbox token.
- [ ] Confirm default map center and zoom.
- [ ] Confirm whether to use backend `priority_score` as the frontend priority field.
- [ ] Decide how to handle missing `gated` field for enforcement rows.
- [ ] Decide how to handle missing ONNX latency, CPU, and memory telemetry.
- [ ] Confirm simulation v1 should match backend request shape.
- [ ] Choose single-page command center or multiple routes.
- [ ] Decide whether mock data fallback is allowed.
- [ ] Decide whether filters should sync to the URL.
- [ ] Confirm responsive target for v1.

## Phase 1: Project Scaffold

- [ ] Create Next.js 14 App Router app in `frontend`.
- [ ] Add TypeScript configuration.
- [ ] Add Tailwind CSS configuration.
- [ ] Add PostCSS configuration.
- [ ] Add ESLint configuration.
- [ ] Add `package.json` scripts.
- [ ] Add `.env.local.example`.
- [ ] Install core dependencies.
- [ ] Verify app starts locally.

## Phase 2: Design System

- [ ] Add Aurelian color tokens to Tailwind.
- [ ] Add Geist Sans and Geist Mono.
- [ ] Create global CSS base styles.
- [ ] Add radial background treatment.
- [ ] Add reusable `cn()` utility.
- [ ] Add priority color helper.
- [ ] Add numeric/date formatting helpers.

## Phase 3: Base UI Components

- [ ] Build `Button`.
- [ ] Build `Badge`.
- [ ] Build `Input`.
- [ ] Build `Slider`.
- [ ] Build `Panel`.
- [ ] Build `MetricCard`.
- [ ] Build `Skeleton`.
- [ ] Build `EmptyState`.
- [ ] Build `OfflineOverlay`.

## Phase 4: API Layer

- [ ] Define API response TypeScript types.
- [ ] Build API client with base URL support.
- [ ] Add health endpoint client.
- [ ] Add hotspots endpoint client.
- [ ] Add ranking endpoint client.
- [ ] Add zone insights endpoint client.
- [ ] Add simulation endpoint client.
- [ ] Add API error normalization.

## Phase 5: Data Hooks

- [ ] Add `useHealth`.
- [ ] Add `useHotspots`.
- [ ] Add `useRanking`.
- [ ] Add `useZoneInsights`.
- [ ] Add `useSimulation`.
- [ ] Configure 30 second SWR polling.
- [ ] Add retry behavior.
- [ ] Add loading and empty-state conventions.

## Phase 6: Global State

- [ ] Create Zustand command store.
- [ ] Store selected zone.
- [ ] Store selected hour.
- [ ] Store selected date.
- [ ] Store police station filter.
- [ ] Store risk level filter.
- [ ] Store ranking limit.
- [ ] Store simulation settings.
- [ ] Store forecast mode state.

## Phase 7: Command Shell

- [ ] Build root layout.
- [ ] Build command-center page.
- [ ] Build top status/navigation bar.
- [ ] Add filter controls.
- [ ] Add backend health indicator.
- [ ] Add responsive panel layout.
- [ ] Add global offline overlay.

## Phase 8: Map

- [ ] Add Mapbox GL dependency and styles.
- [ ] Initialize map with pitch `45`.
- [ ] Initialize map with bearing `-17`.
- [ ] Add missing-token fallback.
- [ ] Convert hotspots to GeoJSON.
- [ ] Render hotspot circle layer.
- [ ] Scale radius by `priority_score`.
- [ ] Color hotspots by priority/risk.
- [ ] Add selected-zone visual state.
- [ ] Add click-to-select behavior.
- [ ] Add popup or compact detail preview.

## Phase 9: Telemetry Panel

- [ ] Render backend status.
- [ ] Render last refresh time.
- [ ] Render selected filters.
- [ ] Render hotspot count.
- [ ] Render ranking count.
- [ ] Add placeholders for unavailable ONNX latency, CPU, and memory metrics.
- [ ] Replace placeholders if backend telemetry becomes available.

## Phase 10: Enforcement Ranking

- [ ] Build ranking panel shell.
- [ ] Build TanStack table columns.
- [ ] Render rank.
- [ ] Render zone ID.
- [ ] Render police station.
- [ ] Render junction.
- [ ] Render expected violations.
- [ ] Render impact score.
- [ ] Render priority score.
- [ ] Render priority level.
- [ ] Render model confidence.
- [ ] Render recommended action.
- [ ] Highlight rows where `priority_score >= 72.0`.
- [ ] Add row click to select zone.
- [ ] Add loading state.
- [ ] Add error state.
- [ ] Add empty state.
- [!] Add Gate icon once backend exposes `gated`.

## Phase 11: Zone Insights

- [ ] Build selected-zone insights panel.
- [ ] Render no-zone-selected state.
- [ ] Render priority score indicator.
- [ ] Render impact score indicator.
- [ ] Render predicted hotspot risk.
- [ ] Render model confidence.
- [ ] Render recommended action.
- [ ] Render LOW/MEDIUM/HIGH probability chart.
- [ ] Render zone stats.
- [ ] Render top violation types.
- [ ] Render top vehicle types.
- [ ] Render Logic Transparency reasons box.
- [ ] Add loading state.
- [ ] Add error state.
- [!] Add ensemble-weight chart if backend exposes LGB/XGB/RF weights.

## Phase 12: Policy Simulation

- [ ] Build simulation panel shell.
- [ ] Require selected zone before simulation.
- [ ] Add violation reduction slider.
- [ ] Add submit action.
- [ ] POST simulation request.
- [ ] Render current priority score and level.
- [ ] Render simulated priority score and level.
- [ ] Render priority change.
- [ ] Render estimated impact reduction.
- [ ] Render backend note.
- [ ] Add Forecast Mode badge.
- [ ] Add Forecast Mode border styling.
- [ ] Add reset simulation action.
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

- [ ] Run lint.
- [ ] Run typecheck.
- [ ] Run production build.
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

