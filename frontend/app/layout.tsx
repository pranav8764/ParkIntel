import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "ParkIntel — Aurelian Command Center",
  description:
    "Real-time illegal parking hotspot monitoring, ML-driven enforcement ranking, and policy simulation command center.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="h-full antialiased dark">
      <body className="min-h-full flex flex-col bg-background text-foreground">
        {children}
      </body>
    </html>
  );
}
