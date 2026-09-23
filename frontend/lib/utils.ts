import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/** Formats an ISO timestamp as an absolute + human-relative string. */
export function formatDateTime(iso?: string | null): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleString(undefined, {
    year: "numeric",
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

export function formatRelativeTime(iso?: string | null): string {
  if (!iso) return "—";
  const d = new Date(iso).getTime();
  if (Number.isNaN(d)) return "—";
  const diffMs = d - Date.now();
  const diffSec = Math.round(diffMs / 1000);
  const abs = Math.abs(diffSec);

  const divisions: [Intl.RelativeTimeFormatUnit, number][] = [
    ["second", 60],
    ["minute", 60],
    ["hour", 24],
    ["day", 30],
    ["month", 12],
    ["year", Infinity],
  ];

  let unit: Intl.RelativeTimeFormatUnit = "second";
  let value = diffSec;
  let remaining = abs;
  for (const [u, size] of divisions) {
    if (remaining < size) {
      unit = u;
      break;
    }
    remaining = Math.floor(remaining / size);
    value = diffSec < 0 ? -remaining : remaining;
    unit = u;
  }

  const rtf = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });
  return rtf.format(value, unit);
}

export function truncateMiddle(str: string, max = 10): string {
  if (str.length <= max) return str;
  const half = Math.floor((max - 1) / 2);
  return `${str.slice(0, half)}…${str.slice(str.length - half)}`;
}

export function numberCompact(n: number): string {
  return new Intl.NumberFormat(undefined, { notation: "compact" }).format(n);
}

export function copyToClipboard(text: string) {
  if (typeof navigator !== "undefined" && navigator.clipboard) {
    return navigator.clipboard.writeText(text);
  }
  return Promise.reject(new Error("Clipboard API unavailable"));
}
