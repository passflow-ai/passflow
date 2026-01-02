import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { Analytics } from "@vercel/analytics/next";
import { SpeedInsights } from "@vercel/speed-insights/next";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Passflow.ai - Identity that accelerates revenue and blocks fraud",
  description: "Real-time identity verification for fintech, lending, and marketplaces. Up to 60% reduction in fraud. 20-40% increase in onboarding conversion.",
  keywords: "identity verification, fraud prevention, KYC, fintech, onboarding, real-time verification, synthetic identity fraud",
  openGraph: {
    title: "Passflow.ai - Identity that accelerates revenue and blocks fraud",
    description: "Real-time identity verification for fintech, lending, and marketplaces.",
    type: "website",
    url: "https://passflow.ai",
  },
  twitter: {
    card: "summary_large_image",
    title: "Passflow.ai - Identity that accelerates revenue and blocks fraud",
    description: "Real-time identity verification for fintech, lending, and marketplaces.",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        {children}
        <Analytics />
        <SpeedInsights />
      </body>
    </html>
  );
}
