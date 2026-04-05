import crypto from 'crypto';

/**
 * Generates an HMAC-SHA256 signature for webhook payloads
 * Format: HMAC-SHA256(timestamp + "." + body)
 */
export function generateWebhookSignature(
  secret: string,
  timestamp: number,
  body: string
): string {
  const signedPayload = `${timestamp}.${body}`;
  return crypto
    .createHmac('sha256', secret)
    .update(signedPayload)
    .digest('hex');
}

/**
 * Creates headers for a signed webhook request
 */
export function createWebhookHeaders(
  secret: string,
  body: string
): Record<string, string> {
  const timestamp = Math.floor(Date.now() / 1000);
  const signature = generateWebhookSignature(secret, timestamp, body);

  return {
    'Content-Type': 'application/json',
    'X-Webhook-Timestamp': timestamp.toString(),
    'X-Webhook-Signature': signature,
  };
}

/**
 * Creates headers for simple secret authentication (fallback)
 */
export function createSimpleWebhookHeaders(
  secret: string
): Record<string, string> {
  return {
    'Content-Type': 'application/json',
    'X-Webhook-Secret': secret,
  };
}
