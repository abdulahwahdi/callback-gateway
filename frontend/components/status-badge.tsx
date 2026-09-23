import { CheckCircle2, Clock3, XCircle } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { PublishStatus } from "@/lib/types";

const CONFIG: Record<
  string,
  { label: string; className: string; icon: React.ComponentType<{ className?: string }> }
> = {
  published: {
    label: "Published",
    className: "bg-success/15 text-success border-success/30",
    icon: CheckCircle2,
  },
  pending: {
    label: "Pending",
    className: "bg-warning/15 text-warning border-warning/30",
    icon: Clock3,
  },
  failed: {
    label: "Failed",
    className: "bg-destructive/15 text-destructive border-destructive/30",
    icon: XCircle,
  },
};

export function StatusBadge({
  status,
  className,
}: {
  status: PublishStatus | string;
  className?: string;
}) {
  const cfg = CONFIG[status] ?? {
    label: status,
    className: "bg-muted text-muted-foreground border-border",
    icon: Clock3,
  };
  const Icon = cfg.icon;

  return (
    <Badge
      variant="outline"
      className={cn("font-medium capitalize", cfg.className, className)}
    >
      <Icon className="h-3 w-3" />
      {cfg.label}
    </Badge>
  );
}
