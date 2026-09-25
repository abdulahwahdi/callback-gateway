"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { Loader2, LogIn, Webhook } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api, ApiError } from "@/lib/api";
import { useSettings } from "@/lib/settings";

export default function LoginPage() {
  const router = useRouter();
  const { settings, setSettings, ready } = useSettings();
  const [username, setUsername] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState("");

  // Already signed in, or login isn't required: go straight to the dashboard.
  React.useEffect(() => {
    if (!ready) return;
    if (settings.token) {
      router.replace("/");
      return;
    }
    api
      .authStatus(settings)
      .then((res) => {
        if (!res.data.login_enabled) router.replace("/");
      })
      .catch(() => {});
  }, [ready]); // eslint-disable-line react-hooks/exhaustive-deps

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await api.login(settings, username.trim(), password);
      setSettings({ token: res.data.token });
      router.replace("/");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Sign in failed");
      setLoading(false);
    }
  };

  return (
    <main className="bg-grid flex min-h-screen items-center justify-center px-4">
      <Card className="w-full max-w-sm animate-fade-in">
        <CardHeader className="space-y-3 text-center">
          <div className="mx-auto flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <Webhook className="h-5 w-5" />
          </div>
          <div className="space-y-1">
            <CardTitle>Sign in</CardTitle>
            <CardDescription>Webhook Middleware dashboard</CardDescription>
          </div>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                autoComplete="username"
                autoFocus
                required
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
            {error && (
              <p role="alert" className="text-sm text-destructive">
                {error}
              </p>
            )}
            <Button type="submit" className="w-full" disabled={loading || !ready}>
              {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <LogIn className="h-4 w-4" />}
              Sign in
            </Button>
          </form>
        </CardContent>
      </Card>
    </main>
  );
}
