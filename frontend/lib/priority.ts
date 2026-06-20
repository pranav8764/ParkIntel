// ============================================================================
// Priority & Risk Color Helpers
// Maps backend priority/risk levels to Aurelian design-system colors.
// ============================================================================

/** Severity threshold from the PRD — rows at or above this are "critical". */
export const CRITICAL_THRESHOLD = 72.0;

/** Check whether a priority score crosses the critical gate. */
export function isCritical(priorityScore: number): boolean {
  return priorityScore >= CRITICAL_THRESHOLD;
}

/**
 * Return a Tailwind text-color class for a given priority level string.
 */
export function priorityLevelColor(level: string): string {
  switch (level.toUpperCase()) {
    case "CRITICAL":
      return "text-gold";
    case "HIGH":
      return "text-gold/80";
    case "MEDIUM":
      return "text-synthetic-blue";
    case "LOW":
      return "text-emerald";
    default:
      return "text-white/60";
  }
}

/**
 * Return a Tailwind bg-color class for a given priority level string.
 */
export function priorityLevelBg(level: string): string {
  switch (level.toUpperCase()) {
    case "CRITICAL":
      return "bg-gold/20";
    case "HIGH":
      return "bg-gold/10";
    case "MEDIUM":
      return "bg-synthetic-blue/15";
    case "LOW":
      return "bg-emerald/15";
    default:
      return "bg-white/5";
  }
}

/**
 * Return a Tailwind border-color class for a given priority level string.
 */
export function priorityLevelBorder(level: string): string {
  switch (level.toUpperCase()) {
    case "CRITICAL":
      return "border-gold/40";
    case "HIGH":
      return "border-gold/25";
    case "MEDIUM":
      return "border-synthetic-blue/25";
    case "LOW":
      return "border-emerald/25";
    default:
      return "border-white/10";
  }
}

/**
 * Return a hex color for chart/map use (not a Tailwind class).
 */
export function priorityHex(level: string): string {
  switch (level.toUpperCase()) {
    case "CRITICAL":
      return "#d4af37";
    case "HIGH":
      return "#d4af37";
    case "MEDIUM":
      return "#3b82f6";
    case "LOW":
      return "#10b981";
    default:
      return "#6b7280";
  }
}

/**
 * Return a hex color for a risk level string (for map circles etc.).
 */
export function riskHex(risk: string): string {
  switch (risk.toUpperCase()) {
    case "HIGH":
      return "#d4af37";
    case "MEDIUM":
      return "#3b82f6";
    case "LOW":
      return "#10b981";
    default:
      return "#6b7280";
  }
}
