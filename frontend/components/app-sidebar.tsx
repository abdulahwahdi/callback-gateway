"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Activity,
  BookOpen,
  ExternalLink,
  Gauge,
  Radio,
  Route,
  Settings as SettingsIcon,
  Webhook,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { useHealth } from "@/hooks/use-webhooks";
import { useSettings } from "@/lib/settings";

const NAV = [
  {
    href: "/",
    label: "Overview",
    icon: Gauge,
    description: "Analytics & trends",
  },
  {
    href: "/webhooks",
    label: "Webhooks",
    icon: Webhook,
    description: "Browse & replay events",
  },
  {
    href: "/sources",
    label: "Sources",
    icon: Radio,
    description: "Gateways seen so far",
  },
  {
    href: "/topics",
    label: "Topic Manager",
    icon: Route,
    description: "Kafka topic overrides",
  },
  {
    href: "/settings",
    label: "Settings",
    icon: SettingsIcon,
    description: "API endpoint & key",
  },
] as const;

export function AppSidebarContent({ onNavigate }: { onNavigate?: () => void }) {
  const pathname = usePathname();
  const { data: health, isError } = useHealth();
  const { settings } = useSettings();
  const online = !isError && Boolean(health);
  const docsUrl = `${settings.apiBaseUrl.replace(/\/+$/, "")}/docs`;

  return (
    <div className="flex h-full flex-col bg-sidebar text-sidebar-foreground">
      <div className="flex items-center gap-2.5 px-5 py-5">
        <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-violet-500 via-fuchsia-500 to-sky-500 shadow-lg shadow-fuchsia-500/20">
          <Webhook className="h-5 w-5 text-white" />
        </div>
        <div className="leading-tight">
          <p className="text-sm font-semibold text-white">webhook-middleware</p>
          <p className="text-[11px] text-sidebar-foreground/60">gateway ingestion</p>
        </div>
      </div>

      <div className="mx-5 mb-4 flex items-center gap-2 rounded-lg border border-sidebar-border bg-white/5 px-3 py-2">
        <span className="relative flex h-2 w-2">
          <span
            className={cn(
              "absolute inline-flex h-full w-full animate-ping rounded-full opacity-75",
              online ? "bg-success" : "bg-destructive"
            )}
          />
          <span
            className={cn(
              "relative inline-flex h-2 w-2 rounded-full",
              online ? "bg-success" : "bg-destructive"
            )}
          />
        </span>
        <span className="text-xs font-medium text-sidebar-foreground/80">
          {online ? "Backend online" : "Backend unreachable"}
        </span>
      </div>

      <nav className="flex-1 space-y-1 px-3">
        {NAV.map((item) => {
          const active =
            item.href === "/" ? pathname === "/" : pathname.startsWith(item.href);
          const Icon = item.icon;
          return (
            <Link
              key={item.href}
              href={item.href}
              onClick={onNavigate}
              className={cn(
                "group relative flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors",
                active
                  ? "bg-white/10 text-white"
                  : "text-sidebar-foreground/70 hover:bg-white/5 hover:text-white"
              )}
            >
              {active && (
                <span className="absolute left-0 top-1/2 h-5 w-1 -translate-y-1/2 rounded-r-full bg-gradient-to-b from-violet-400 to-sky-400" />
              )}
              <Icon className="h-4 w-4 shrink-0" />
              <span className="flex-1">{item.label}</span>
            </Link>
          );
        })}

        <a
          href={docsUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="group relative flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-sidebar-foreground/70 transition-colors hover:bg-white/5 hover:text-white"
        >
          <BookOpen className="h-4 w-4 shrink-0" />
          <span className="flex-1">API Docs</span>
          <ExternalLink className="h-3.5 w-3.5 shrink-0 opacity-0 transition-opacity group-hover:opacity-60" />
        </a>
      </nav>

      <div className="mx-3 mb-4 mt-2 rounded-xl border border-sidebar-border bg-white/5 p-3.5">
        <div className="mb-1.5 flex items-center gap-2 text-xs font-semibold text-white">
          <Activity className="h-3.5 w-3.5 text-fuchsia-400" />
          Retry worker
        </div>
        <p className="text-[11px] leading-relaxed text-sidebar-foreground/60">
          Failed publishes to Kafka self-heal automatically in the background.
        </p>
      </div>
    </div>
  );
}

export function AppSidebar() {
  return (
    <aside className="fixed inset-y-0 left-0 z-30 hidden w-64 border-r border-sidebar-border lg:block">
      <AppSidebarContent />
    </aside>
  );
}
