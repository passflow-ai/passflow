import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { Analytics } from "@vercel/analytics/next";
import { SpeedInsights } from "@vercel/speed-insights/next";
import "./globals.css";

const jsonLd = {
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "Organization",
      "@id": "https://passflow.ai/#organization",
      name: "PassFlow, Inc",
      url: "https://passflow.ai",
      logo: "https://passflow.ai/logo.png",
      description: "AI agent orchestration platform. Build, deploy, and manage autonomous AI agents.",
      address: {
        "@type": "PostalAddress",
        streetAddress: "3723 Greenville Ave STE 57075",
        addressLocality: "Dallas",
        addressRegion: "TX",
        postalCode: "75206",
        addressCountry: "US"
      },
      telephone: "+1-469-249-3858",
      email: "hello@passflow.ai",
      sameAs: []
    },
    {
      "@type": "WebSite",
      "@id": "https://passflow.ai/#website",
      url: "https://passflow.ai",
      name: "Passflow.ai",
      publisher: { "@id": "https://passflow.ai/#organization" }
    },
    {
      "@type": "SoftwareApplication",
      "@id": "https://passflow.ai/#product",
      name: "Passflow AI Agent Platform",
      applicationCategory: "BusinessApplication",
      operatingSystem: "Web",
      description: "AI-powered agent orchestration platform for workflow automation with multiple LLM providers.",
      offers: {
        "@type": "AggregateOffer",
        priceCurrency: "USD",
        lowPrice: "0",
        highPrice: "799",
        offerCount: "4"
      },
      provider: { "@id": "https://passflow.ai/#organization" },
      featureList: [
        "Visual workflow builder for AI agents",
        "Multi-LLM support (Anthropic, OpenAI, Gemini, Azure)",
        "MCP integrations and custom tools",
        "Automated triggers (cron, Slack, email, webhooks)",
        "BYOK - Bring Your Own API Key",
        "Real-time execution monitoring and analytics"
      ]
    }
  ]
};

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  metadataBase: new URL("https://passflow.ai"),
  title: {
    default: "Passflow.ai - AI Agent Orchestration Platform",
    template: "%s | Passflow.ai",
  },
  description: "Build, deploy, and manage AI agents that automate your workflows 24/7. No code required. Free to start. Connect Anthropic, OpenAI, and more.",
  keywords: [
    "AI agents",
    "workflow automation",
    "AI orchestration",
    "LLM automation",
    "agent builder",
    "no-code AI",
    "MCP protocol",
    "AI workflow",
    "autonomous agents",
    "business automation",
    "AI platform",
    "BYOK AI"
  ],
  authors: [{ name: "PassFlow, Inc" }],
  creator: "PassFlow, Inc",
  publisher: "PassFlow, Inc",
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
  openGraph: {
    type: "website",
    locale: "en_US",
    url: "https://passflow.ai",
    siteName: "Passflow.ai",
    title: "Passflow.ai - AI Agent Orchestration Platform",
    description: "Build, deploy, and manage AI agents that automate your workflows 24/7. No code required. Free to start. Connect Anthropic, OpenAI, and more.",
    images: [
      {
        url: "/og-image.png",
        width: 1200,
        height: 630,
        alt: "Passflow.ai - AI Agent Orchestration Platform",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Passflow.ai - AI Agent Orchestration Platform",
    description: "Build, deploy, and manage AI agents that automate your workflows 24/7. No code required. Free to start. Connect Anthropic, OpenAI, and more.",
    images: ["/og-image.png"],
  },
  alternates: {
    canonical: "https://passflow.ai",
  },
  icons: {
    icon: [
      { url: "/favicon.svg", type: "image/svg+xml" },
    ],
    apple: "/favicon.svg",
  },
  manifest: "/site.webmanifest",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <head>
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
        />
      </head>
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        {children}
        <Analytics />
        <SpeedInsights />
      </body>
    </html>
  );
}
