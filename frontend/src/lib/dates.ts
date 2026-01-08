// Brisbane timezone utilities
const BRISBANE_TZ = "Australia/Brisbane";
const LOCALE = "en-AU";

/**
 * Format a date string or Date object for display in Brisbane timezone
 */
export function formatDate(
  date: string | Date,
  options?: Intl.DateTimeFormatOptions
): string {
  const d = typeof date === "string" ? new Date(date) : date;
  return d.toLocaleDateString(LOCALE, {
    timeZone: BRISBANE_TZ,
    ...options,
  });
}

/**
 * Format a datetime string or Date object for display in Brisbane timezone
 */
export function formatDateTime(
  date: string | Date,
  options?: Intl.DateTimeFormatOptions
): string {
  const d = typeof date === "string" ? new Date(date) : date;
  return d.toLocaleString(LOCALE, {
    timeZone: BRISBANE_TZ,
    dateStyle: "medium",
    timeStyle: "short",
    ...options,
  });
}

/**
 * Format a time string or Date object for display in Brisbane timezone
 */
export function formatTime(
  date: string | Date,
  options?: Intl.DateTimeFormatOptions
): string {
  const d = typeof date === "string" ? new Date(date) : date;
  return d.toLocaleTimeString(LOCALE, {
    timeZone: BRISBANE_TZ,
    timeStyle: "short",
    ...options,
  });
}

/**
 * Get today's date in Brisbane timezone as YYYY-MM-DD string
 */
export function getTodayBrisbane(): string {
  const now = new Date();
  return now.toLocaleDateString("en-CA", { timeZone: BRISBANE_TZ }); // en-CA gives YYYY-MM-DD format
}

/**
 * Get date N days ago in Brisbane timezone as YYYY-MM-DD string
 */
export function getDaysAgoBrisbane(days: number): string {
  const date = new Date();
  date.setDate(date.getDate() - days);
  return date.toLocaleDateString("en-CA", { timeZone: BRISBANE_TZ });
}

/**
 * Get start of year in Brisbane timezone as YYYY-MM-DD string
 */
export function getYearStartBrisbane(): string {
  const now = new Date();
  const year = parseInt(
    now.toLocaleDateString("en-CA", { timeZone: BRISBANE_TZ }).split("-")[0]
  );
  return `${year}-01-01`;
}

/**
 * Format currency in AUD
 */
export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat(LOCALE, {
    style: "currency",
    currency: "AUD",
  }).format(amount);
}

/**
 * Format percentage
 */
export function formatPercent(value: number, decimals = 1): string {
  return `${value.toFixed(decimals)}%`;
}

/**
 * Format a freshness timestamp with relative time
 */
export function formatFreshness(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  let relative: string;
  if (diffMins < 1) {
    relative = "just now";
  } else if (diffMins < 60) {
    relative = `${diffMins}m ago`;
  } else if (diffHours < 24) {
    relative = `${diffHours}h ago`;
  } else {
    relative = `${diffDays}d ago`;
  }

  return `${formatDateTime(date)} (${relative})`;
}

/**
 * Parse a date range string into start and end dates
 */
export function parseDateRange(range: string): { startDate: string; endDate: string } {
  const today = getTodayBrisbane();
  
  switch (range) {
    case "30d":
      return { startDate: getDaysAgoBrisbane(30), endDate: today };
    case "ytd":
      return { startDate: getYearStartBrisbane(), endDate: today };
    case "trailing12m":
      return { startDate: getDaysAgoBrisbane(365), endDate: today };
    default:
      return { startDate: getDaysAgoBrisbane(30), endDate: today };
  }
}
