"use client";

import * as React from "react";

const STORAGE_KEY = "webhook-middleware:settings:v1";

export interface DashboardSettings {
  apiBaseUrl: string;
  apiKey: string;
  /** Session token from POST /auth/login (empty when logged out). */
  token: string;
}

function defaults(): DashboardSettings {
  return {
    apiBaseUrl:
      process.env.NEXT_PUBLIC_API_BASE_URL?.replace(/\/+$/, "") ||
      "http://localhost:8090",
    apiKey: process.env.NEXT_PUBLIC_API_KEY || "",
    token: "",
  };
}

function load(): DashboardSettings {
  if (typeof window === "undefined") return defaults();
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return defaults();
    const parsed = JSON.parse(raw);
    return { ...defaults(), ...parsed };
  } catch {
    return defaults();
  }
}

interface SettingsContextValue {
  settings: DashboardSettings;
  setSettings: (next: Partial<DashboardSettings>) => void;
  reset: () => void;
  /** True once localStorage has been read (avoids a logged-out flash on load). */
  ready: boolean;
}

const SettingsContext = React.createContext<SettingsContextValue | null>(null);

export function SettingsProvider({ children }: { children: React.ReactNode }) {
  const [settings, setSettingsState] = React.useState<DashboardSettings>(defaults);

  const [ready, setReady] = React.useState(false);

  React.useEffect(() => {
    setSettingsState(load());
    setReady(true);
  }, []);

  const setSettings = React.useCallback((next: Partial<DashboardSettings>) => {
    setSettingsState((prev) => {
      const merged = { ...prev, ...next };
      if (typeof window !== "undefined") {
        window.localStorage.setItem(STORAGE_KEY, JSON.stringify(merged));
      }
      return merged;
    });
  }, []);

  const reset = React.useCallback(() => {
    const d = defaults();
    setSettingsState(d);
    if (typeof window !== "undefined") {
      window.localStorage.removeItem(STORAGE_KEY);
    }
  }, []);

  const value = React.useMemo(
    () => ({ settings, setSettings, reset, ready }),
    [settings, setSettings, reset, ready]
  );

  return React.createElement(SettingsContext.Provider, { value }, children);
}

export function useSettings() {
  const ctx = React.useContext(SettingsContext);
  if (!ctx) throw new Error("useSettings must be used within SettingsProvider");
  return ctx;
}
