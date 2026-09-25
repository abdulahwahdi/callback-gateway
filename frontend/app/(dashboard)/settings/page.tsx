"use client";

import * as React from "react";
import { CheckCircle2, KeyRound, Loader2, Server, XCircle } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { useSettings } from "@/lib/settings";
import { api, ApiError } from "@/lib/api";

export default function SettingsPage() {
  const { settings, setSettings, reset } = useSettings();
  const [apiBaseUrl, setApiBaseUrl] = React.useState(settings.apiBaseUrl);
  const [apiKey, setApiKey] = React.useState(settings.apiKey);
  const [testState, setTestState] = React.useState<"idle" | "loading" | "ok" | "error">("idle");
  const [testMessage, setTestMessage] = React.useState<string>("");

  React.useEffect(() => {
    setApiBaseUrl(settings.apiBaseUrl);
    setApiKey(settings.apiKey);
  }, [settings.apiBaseUrl, settings.apiKey]);

  const dirty = apiBaseUrl !== settings.apiBaseUrl || apiKey !== settings.apiKey;

  const handleSave = () => {
    setSettings({ apiBaseUrl: apiBaseUrl.trim().replace(/\/+$/, ""), apiKey: apiKey.trim() });
    toast.success("Settings saved", { description: "Stored in this browser's localStorage." });
  };

  const handleTest = async () => {
    setTestState("loading");
    try {
      await api.health({ apiBaseUrl: apiBaseUrl.trim().replace(/\/+$/, ""), apiKey: apiKey.trim(), token: settings.token });
      setTestState("ok");
      setTestMessage("Connected successfully.");
    } catch (err) {
      setTestState("error");
      setTestMessage(err instanceof ApiError ? err.message : "Unknown error");
    }
  };

  return (
    <div className="max-w-2xl space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Server className="h-4 w-4 text-primary" />
            Backend connection
          </CardTitle>
          <CardDescription>
            Points this dashboard at your webhook-middleware Go service. Stored locally in this
            browser — nothing is sent anywhere else.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-5">
          <div className="space-y-2">
            <Label htmlFor="api-base-url">API base URL</Label>
            <Input
              id="api-base-url"
              placeholder="http://localhost:8090"
              value={apiBaseUrl}
              onChange={(e) => setApiBaseUrl(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              The root address of the webhook-middleware service (where <code>/healthz</code> and{" "}
              <code>/dashboard/*</code> are mounted).
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="api-key" className="flex items-center gap-1.5">
              <KeyRound className="h-3.5 w-3.5" /> X-API-Key
              <span className="font-normal text-muted-foreground">(optional)</span>
            </Label>
            <Input
              id="api-key"
              type="password"
              placeholder="Only needed if DASHBOARD_API_KEY is set on the backend"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
            />
          </div>

          <Separator />

          <div className="flex flex-wrap items-center gap-3">
            <Button onClick={handleSave} disabled={!dirty}>
              Save changes
            </Button>
            <Button variant="outline" onClick={handleTest} disabled={testState === "loading"}>
              {testState === "loading" && <Loader2 className="h-4 w-4 animate-spin" />}
              Test connection
            </Button>
            <Button
              variant="ghost"
              onClick={() => {
                reset();
                toast.info("Settings reset to defaults");
              }}
            >
              Reset
            </Button>
          </div>

          {testState === "ok" && (
            <p className="flex items-center gap-1.5 text-sm text-success">
              <CheckCircle2 className="h-4 w-4" /> {testMessage}
            </p>
          )}
          {testState === "error" && (
            <p className="flex items-start gap-1.5 text-sm text-destructive">
              <XCircle className="mt-0.5 h-4 w-4 shrink-0" /> {testMessage}
            </p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>About this dashboard</CardTitle>
          <CardDescription>What each menu talks to on the backend.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-2 text-sm text-muted-foreground">
          <Row label="Overview" endpoint="GET /dashboard/webhooks/stats" />
          <Row label="Webhooks" endpoint="GET /dashboard/webhooks, /:id, /:id/replay, /retry-failed" />
          <Row label="Sources" endpoint="GET /dashboard/webhooks/sources" />
        </CardContent>
      </Card>
    </div>
  );
}

function Row({ label, endpoint }: { label: string; endpoint: string }) {
  return (
    <div className="flex flex-col gap-0.5 border-b border-border/60 py-2 last:border-0 sm:flex-row sm:items-center sm:justify-between">
      <span className="font-medium text-foreground">{label}</span>
      <code className="text-xs">{endpoint}</code>
    </div>
  );
}
