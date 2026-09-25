"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { Loader2 } from "lucide-react";

import { api } from "@/lib/api";
import { useSettings } from "@/lib/settings";

/**
 * Sends visitors to /login when the backend has login enabled and this
 * browser has no session token (or the token was rejected).
 */
export function AuthGate({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const { settings, setSettings, ready } = useSettings();
  const [allowed, setAllowed] = React.useState(false);

  React.useEffect(() => {
    if (!ready) return;
    let cancelled = false;
    api
      .authStatus(settings)
      .then((res) => {
        if (cancelled) return;
        if (res.data.login_enabled && !settings.token) {
          router.replace("/login");
        } else {
          setAllowed(true);
        }
      })
      // Backend unreachable or older backend without /auth/status: show the
      // dashboard, which surfaces connection errors itself.
      .catch(() => !cancelled && setAllowed(true));
    return () => {
      cancelled = true;
    };
  }, [ready, settings.apiBaseUrl, settings.token]); // eslint-disable-line react-hooks/exhaustive-deps

  React.useEffect(() => {
    const onUnauthorized = () => {
      setSettings({ token: "" });
      setAllowed(false);
      router.replace("/login");
    };
    window.addEventListener("auth:unauthorized", onUnauthorized);
    return () => window.removeEventListener("auth:unauthorized", onUnauthorized);
  }, [router, setSettings]);

  if (!allowed) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
      </div>
    );
  }
  return <>{children}</>;
}
