"use client";

import * as React from "react";
import { Check, Copy } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { copyToClipboard, cn } from "@/lib/utils";

export function JsonViewer({
  value,
  className,
  emptyLabel = "No data",
}: {
  value: unknown;
  className?: string;
  emptyLabel?: string;
}) {
  const [copied, setCopied] = React.useState(false);

  const pretty = React.useMemo(() => {
    if (value === null || value === undefined) return "";
    try {
      return JSON.stringify(value, null, 2);
    } catch {
      return String(value);
    }
  }, [value]);

  const handleCopy = async () => {
    try {
      await copyToClipboard(pretty);
      setCopied(true);
      toast.success("Copied to clipboard");
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      toast.error("Couldn't copy to clipboard");
    }
  };

  if (!pretty || pretty === "{}" || pretty === "null") {
    return (
      <div className={cn("rounded-lg border border-dashed p-4 text-sm text-muted-foreground", className)}>
        {emptyLabel}
      </div>
    );
  }

  return (
    <div className={cn("group relative", className)}>
      <Button
        variant="outline"
        size="icon"
        className="absolute right-2 top-2 h-7 w-7 opacity-0 transition-opacity group-hover:opacity-100"
        onClick={handleCopy}
        aria-label="Copy JSON"
      >
        {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
      </Button>
      <pre className="scrollbar-thin max-h-96 overflow-auto rounded-lg border bg-muted/40 p-4 font-mono text-xs leading-relaxed">
        {pretty}
      </pre>
    </div>
  );
}
