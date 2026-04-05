import { NextResponse } from "next/server";
import { Resend } from "resend";
import { createWebhookHeaders } from "@/lib/webhook-signature";

export async function POST(request: Request) {
  try {
    const { name, email, company, industry, volume, timeline } = await request.json();

    if (!name || !email || !company) {
      return NextResponse.json(
        { error: "Missing required fields" },
        { status: 400 }
      );
    }

    // Send to CRM webhook
    const crmWebhookUrl = process.env.CRM_WEBHOOK_URL;
    const crmWebhookSecret = process.env.CRM_WEBHOOK_SECRET;

    if (crmWebhookUrl && crmWebhookSecret) {
      try {
        const payload = JSON.stringify({
          name,
          email,
          company,
          service_interest: "kyc", // PassFlow is an identity verification product
          message: `Industry: ${industry || "Not specified"}\nMonthly Volume: ${volume || "Not specified"}\nTimeline: ${timeline || "Not specified"}`,
        });

        // Uses HMAC signature for security
        await fetch(`${crmWebhookUrl}/passflow/contact`, {
          method: "POST",
          headers: createWebhookHeaders(crmWebhookSecret, payload),
          body: payload,
        });
      } catch (crmError) {
        console.error("CRM webhook error:", crmError);
        // Don't fail the request if CRM fails
      }
    }

    // Send email notification
    const resend = new Resend(process.env.RESEND_API_KEY);
    await resend.emails.send({
      from: "Passflow <noreply@passflow.ai>",
      to: process.env.CONTACT_EMAIL || "sales@passflow.ai",
      subject: `New lead from Passflow.ai: ${company}`,
      html: `
        <h2>New Contact from Passflow.ai</h2>
        <p><strong>Name:</strong> ${name}</p>
        <p><strong>Email:</strong> ${email}</p>
        <p><strong>Company:</strong> ${company}</p>
        <p><strong>Industry:</strong> ${industry || "Not specified"}</p>
        <p><strong>Monthly Volume:</strong> ${volume || "Not specified"}</p>
        <p><strong>Timeline:</strong> ${timeline || "Not specified"}</p>
      `,
    });

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error("Contact form error:", error);
    return NextResponse.json(
      { error: "Failed to send message" },
      { status: 500 }
    );
  }
}
