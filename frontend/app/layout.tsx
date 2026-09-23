import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from "@/components/theme-provider";
import { QueryProvider } from "@/components/providers/query-provider";
import { SettingsProvider } from "@/lib/settings";
import { Toaster } from "@/components/ui/sonner";

export const metadata: Metadata = {
  title: "Webhook Middleware · Dashboard",
  description:
    "Analytics dashboard for the webhook-middleware payment gateway ingestion service.",
};

const fontSans = Inter({
  subsets: ["latin"],
  variable: "--font-geist-sans",
  display: "swap",
});

const fontMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-geist-mono",
  display: "swap",
});

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${fontSans.variable} ${fontMono.variable} font-sans`}>
        <ThemeProvider
          attribute="class"
          defaultTheme="dark"
          enableSystem
          disableTransitionOnChange
        >
          <SettingsProvider>
            <QueryProvider>
              {children}
              <Toaster richColors position="top-right" closeButton />
            </QueryProvider>
          </SettingsProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
