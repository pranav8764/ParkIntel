// ============================================================================
// Formatting Helpers
// Consistent number / date / label formatting for the command center.
// ============================================================================

/**
 * Format a number to a fixed number of decimal places.
 * Falls back to "—" for nullish values.
 */
export function formatScore(value: number | null | undefined, decimals = 1): string {
  if (value == null || isNaN(value)) return "—";
  return value.toFixed(decimals);
}

/**
 * Format a large integer with locale grouping (e.g. 1,234).
 */
export function formatCount(value: number | null | undefined): string {
  if (value == null || isNaN(value)) return "—";
  return value.toLocaleString("en-IN");
}

/**
 * Format a percentage value (0–100) for display.
 */
export function formatPercent(value: number | null | undefined, decimals = 1): string {
  if (value == null || isNaN(value)) return "—";
  return `${value.toFixed(decimals)}%`;
}

/**
 * Format a probability (0–1) as a percentage string.
 */
export function formatProbability(value: number | null | undefined, decimals = 1): string {
  if (value == null || isNaN(value)) return "—";
  return `${(value * 100).toFixed(decimals)}%`;
}

/**
 * Format an hour (0–23) as a readable time label.
 */
export function formatHour(hour: number): string {
  if (hour < 0 || hour > 23) return "—";
  const suffix = hour >= 12 ? "PM" : "AM";
  const display = hour === 0 ? 12 : hour > 12 ? hour - 12 : hour;
  return `${display}:00 ${suffix}`;
}

/**
 * Format a date string (YYYY-MM-DD) for display.
 */
export function formatDate(dateStr: string | null | undefined): string {
  if (!dateStr) return "—";
  try {
    const date = new Date(dateStr + "T00:00:00");
    return date.toLocaleDateString("en-IN", {
      day: "numeric",
      month: "short",
      year: "numeric",
    });
  } catch {
    return dateStr;
  }
}

/**
 * Format a timestamp for display (last-refreshed indicator).
 */
export function formatTimestamp(date: Date): string {
  return date.toLocaleTimeString("en-IN", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: true,
  });
}

/**
 * Truncate a string with ellipsis.
 */
export function truncate(str: string, maxLength: number): string {
  if (str.length <= maxLength) return str;
  return str.slice(0, maxLength - 1) + "…";
}
