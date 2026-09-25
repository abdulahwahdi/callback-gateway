"use client";

import * as React from "react";
import { usePathname, useRouter } from "next/navigation";
import { LogOut, Menu, RefreshCw } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";

import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import { VisuallyHidden } from "@/components/ui/visually-hidden";
import { AppSidebarContent } from "@/components/app-sidebar";
import { useSettings } from "@/lib/settings";
import { ThemeToggle } from "@/components/theme-toggle";

const TITLES: Record<string, { title: string; description: string }> = {
  "/": {
    title: "Overview",
    description: "Real-time analytics across every payment gateway webhook.",
  },
  "/webhooks": {
    title: "Webhooks",
    description: "Browse, filter, inspect and replay every inbound event.",
  },
  "/sources": {
    title: "Sources",
    description: "Distinct payment gateways seen by this middleware.",
  },
  "/topics": {
    title: "Topic Manager",
    description: "Route each (env, source) pair to a specific Kafka topic.",
  },
  "/settings": {
    title: "Settings",
    description: "Point this dashboard at your webhook-middleware backend.",
  },
};

function resolveTitle(pathname: string) {
  if (TITLES[pathname]) return TITLES[pathname];
  const base = "/" + (pathname.split("/")[1] ?? "");
  return TITLES[base] ?? { title: "Dashboard", description: "" };
}

export function Topbar() {
  const pathname = usePathname();
  const router = useRouter();
  const qc = useQueryClient();
  const [mobileOpen, setMobileOpen] = React.useState(false);
  const [spinning, setSpinning] = React.useState(false);
  const { settings, setSettings } = useSettings();
  const { title, description } = resolveTitle(pathname);

  const handleRefresh = () => {
    setSpinning(true);
    qc.invalidateQueries();
    router.refresh();
    window.setTimeout(() => setSpinning(false), 600);
  };

  return (
    <header className="sticky top-0 z-20 flex items-center gap-3 border-b border-border/70 bg-background/80 px-4 py-3.5 backdrop-blur-md sm:px-6">
      <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
        <Button
          variant="outline"
          size="icon"
          className="lg:hidden"
          onClick={() => setMobileOpen(true)}
          aria-label="Open navigation"
        >
          <Menu className="h-4 w-4" />
        </Button>
        <SheetContent side="left" className="w-72 p-0 sm:max-w-72">
          <VisuallyHidden>
            <SheetTitle>Navigation</SheetTitle>
          </VisuallyHidden>
          <AppSidebarContent onNavigate={() => setMobileOpen(false)} />
        </SheetContent>
      </Sheet>

      <div className="min-w-0 flex-1">
        <h1 className="truncate text-lg font-semibold tracking-tight">{title}</h1>
        {description && (
          <p className="hidden truncate text-xs text-muted-foreground sm:block">
            {description}
          </p>
        )}
      </div>

      <Button variant="outline" size="icon" onClick={handleRefresh} aria-label="Refresh data">
        <RefreshCw className={spinning ? "h-4 w-4 animate-spin" : "h-4 w-4"} />
      </Button>
      <ThemeToggle />
      {settings.token && (
        <Button
          variant="outline"
          size="icon"
          aria-label="Sign out"
          onClick={() => {
            setSettings({ token: "" });
            qc.clear();
            router.replace("/login");
          }}
        >
          <LogOut className="h-4 w-4" />
        </Button>
      )}
    </header>
  );
}
