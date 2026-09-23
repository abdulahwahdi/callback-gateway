"use client";

import Link from "next/link";
import { ArrowUpRight, Radio } from "lucide-react";

import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useSources } from "@/hooks/use-webhooks";
import { numberCompact } from "@/lib/utils";

const GRADIENTS = [
  "from-violet-500 to-fuchsia-500",
  "from-sky-500 to-cyan-400",
  "from-emerald-500 to-lime-400",
  "from-amber-500 to-orange-500",
  "from-pink-500 to-rose-500",
  "from-indigo-500 to-blue-500",
];

export default function SourcesPage() {
  const { data, isLoading } = useSources();
  const sources = [...(data?.data ?? [])].sort((a, b) => b.total - a.total);

  if (!isLoading && sources.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center gap-3 py-16 text-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-muted">
            <Radio className="h-5 w-5 text-muted-foreground" />
          </div>
          <p className="text-sm font-medium">No gateways have sent anything yet</p>
          <p className="max-w-sm text-xs text-muted-foreground">
            Sources appear here automatically the moment a payment gateway calls{" "}
            <code className="rounded bg-muted px-1 py-0.5">POST /webhooks/:source</code>.
          </p>
        </CardContent>
      </Card>
    );
  }

  const total = sources.reduce((sum, s) => sum + s.total, 0);

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
      {isLoading
        ? Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-40 w-full rounded-xl" />)
        : sources.map((s, i) => {
            const share = total > 0 ? Math.round((s.total / total) * 100) : 0;
            return (
              <Link key={s.source} href={`/webhooks?source=${encodeURIComponent(s.source)}`}>
                <Card className="group h-full overflow-hidden transition-all hover:-translate-y-0.5 hover:shadow-lg">
                  <div
                    className={`h-1.5 w-full bg-gradient-to-r ${GRADIENTS[i % GRADIENTS.length]}`}
                  />
                  <CardContent className="space-y-4 p-5">
                    <div className="flex items-start justify-between">
                      <div>
                        <p className="text-base font-semibold capitalize">{s.source}</p>
                        <p className="text-xs text-muted-foreground">payment gateway</p>
                      </div>
                      <ArrowUpRight className="h-4 w-4 text-muted-foreground transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
                    </div>

                    <div>
                      <p className="text-3xl font-semibold tabular-nums tracking-tight">
                        {numberCompact(s.total)}
                      </p>
                      <p className="text-xs text-muted-foreground">events received</p>
                    </div>

                    <div className="space-y-1">
                      <div className="h-1.5 w-full overflow-hidden rounded-full bg-muted">
                        <div
                          className={`h-full rounded-full bg-gradient-to-r ${GRADIENTS[i % GRADIENTS.length]}`}
                          style={{ width: `${share}%` }}
                        />
                      </div>
                      <p className="text-[11px] text-muted-foreground">
                        {share}% of all traffic
                      </p>
                    </div>
                  </CardContent>
                </Card>
              </Link>
            );
          })}
    </div>
  );
}
