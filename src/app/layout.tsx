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
      description: "Real-time identity verification for fintech, lending, and marketplaces.",
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
      name: "Passflow Identity Verification",
      applicationCategory: "BusinessApplication",
      operatingSystem: "Web",
      description: "AI-powered identity verification with document verification, biometric matching, and liveness detection.",
      offers: {
        "@type": "AggregateOffer",
        priceCurrency: "USD",
        lowPrice: "0.45",
        highPrice: "1.50",
        offerCount: "3"
      },
      provider: { "@id": "https://passflow.ai/#organization" },
      featureList: [
        "Document verification for 190+ countries",
        "iBeta Level 1 certified liveness detection",
        "Biometric matching with 99.9% accuracy",
        "Real-time verification in under 10 seconds",
        "ISO 27001 and ISO 9001 certified",
        "NIST compliant"
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
    default: "Passflow.ai - Identity Verification that Accelerates Revenue",
    template: "%s | Passflow.ai",
  },
  description: "Real-time identity verification for fintech, lending, and marketplaces. Up to 60% reduction in fraud. 20-40% increase in onboarding conversion. ISO 27001 certified.",
  keywords: [
    "identity verification",
    "KYC solution",
    "fraud prevention",
    "document verification",
    "biometric verification",
    "liveness detection",
    "fintech onboarding",
    "AML compliance",
    "synthetic identity fraud",
    "real-time verification",
    "eKYC",
    "identity proofing"
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
    title: "Passflow.ai - Identity Verification that Accelerates Revenue",
    description: "Real-time identity verification for fintech, lending, and marketplaces. Reduce fraud by 60% while increasing conversion by 20-40%.",
    images: [
      {
        url: "/og-image.png",
        width: 1200,
        height: 630,
        alt: "Passflow.ai - Identity Verification",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Passflow.ai - Identity Verification that Accelerates Revenue",
    description: "Real-time identity verification for fintech, lending, and marketplaces. Reduce fraud by 60%.",
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
