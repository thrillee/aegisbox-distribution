# HTTP Provider Webhook Setup Guide

This document explains how to configure webhook URLs for each supported HTTP SMS provider to receive Delivery Receipt (DLR) callbacks.

## Overview

When aegisbox sends SMS through HTTP providers, it needs to receive delivery status updates (DLRs) back from the providers. This is achieved through webhook callbacks.

## Webhook Endpoints

All webhook endpoints are hosted at: `{API_BASE_URL}/api/v1/webhooks/dlr/{provider}`

For example, if your `API_BASE_URL` is `https://sms-api.example.com`, the webhook URLs would be:
- Termii: `https://sms-api.example.com/api/v1/webhooks/dlr/termii`
- AfricaTalking: `https://sms-api.example.com/api/v1/webhooks/dlr/africatalking`
- Simpu: `https://sms-api.example.com/api/v1/webhooks/dlr/simpu`
- eBulkSMS: `https://sms-api.example.com/api/v1/webhooks/dlr/ebulksms`
- MultiTexter: `https://sms-api.example.com/api/v1/webhooks/dlr/multitexter`

## Configuration

### 1. Set API Base URL

Configure your public API base URL in environment variables:

```bash
export API_BASE_URL="https://your-domain.com"
```

Or in your `.env` file:
```
API_BASE_URL=https://your-domain.com
```

### 2. Provider-Specific Setup

#### Termii

**Dashboard Configuration:**
1. Log in to [Termii Dashboard](https://termii.com)
2. Go to **Settings** → **Webhooks**
3. Add webhook URL: `https://your-domain.com/api/v1/webhooks/dlr/termii`
4. Select event types: **Delivery Reports**
5. Save configuration

**Webhook Format:**
Termii sends POST requests with JSON body:
```json
{
  "type": "dlr",
  "id": "unique_webhook_id",
  "message_id": "termii_message_id",
  "receiver": "2348031234567",
  "sender": "YourBrand",
  "status": "Delivered",
  "sent_at": "2024-01-15 10:30:00",
  "cost": "0.50",
  "channel": "generic"
}
```

**Status Mapping:**
- `Delivered` → delivered
- `Sent` → sent
- `Failed` → failed
- `DND Active` → failed

---

#### AfricaTalking

**Dashboard Configuration:**
1. Log in to [AfricaTalking Dashboard](https://account.africastalking.com)
2. Go to **SMS** → **Callback URLs**
3. Set **Delivery Reports Callback URL**: `https://your-domain.com/api/v1/webhooks/dlr/africatalking`
4. Save configuration

**Webhook Format:**
AfricaTalking sends POST requests with form-encoded data:
```
id=ATXid_xxxxx&status=Success&phoneNumber=+2348031234567
```

**Status Mapping:**
- `Success` → delivered
- `Sent` → sent
- `Failed` → failed
- `Rejected` → failed

---

#### Simpu

**Dashboard Configuration:**
1. Log in to [Simpu Dashboard](https://simpu.co)
2. Go to **Settings** → **API & Webhooks**
3. Add webhook URL: `https://your-domain.com/api/v1/webhooks/dlr/simpu`
4. Enable **SMS Delivery Status** events
5. Save configuration

**Webhook Format:**
Simpu sends POST requests with JSON body:
```json
{
  "event": "message.status.update",
  "data": {
    "message_id": "simpu_msg_id",
    "status": "delivered",
    "timestamp": "2024-01-15T10:30:00Z",
    "callback": "https://your-domain.com/api/v1/webhooks/dlr/simpu?conn_id=1&msg_id=12345"
  }
}
```

**Note:** Simpu's callback URL includes message metadata in the query string, which we parse using `extractMessageIDFromCallback()`.

**Status Mapping:**
- `delivered` → delivered
- `sent` → sent
- `failed` → failed
- `rejected` → failed

---

#### eBulkSMS

**Dashboard Configuration:**
1. Log in to [eBulkSMS Dashboard](https://www.ebulksms.com)
2. Go to **Account** → **API Settings**
3. Set **Delivery Callback URL**: `https://your-domain.com/api/v1/webhooks/dlr/ebulksms`
4. Save configuration

**Webhook Format:**
eBulkSMS sends POST requests with JSON body:
```json
{
  "message_id": "ebulk_msg_id",
  "status": "delivered",
  "recipient": "2348031234567",
  "delivery_time": "2024-01-15 10:30:00"
}
```

**Status Mapping:**
- `delivered` → delivered
- `sent` → sent
- `failed` → failed
- `pending` → sent

---

#### MultiTexter

**Dashboard Configuration:**
1. Log in to [MultiTexter Dashboard](https://multitexter.com)
2. Go to **Settings** → **Webhooks**
3. Add webhook URL: `https://your-domain.com/api/v1/webhooks/dlr/multitexter`
4. Enable **SMS Status Updates**
5. Save configuration

**Webhook Format:**
MultiTexter sends POST requests with JSON body:
```json
{
  "message_id": "mt_msg_id",
  "status": "1",
  "phone": "2348031234567",
  "timestamp": "2024-01-15 10:30:00"
}
```

**Status Mapping:**
- `1` → delivered
- `0` → sent
- `2` → failed
- `3` → rejected → failed

---

## Testing Webhooks

### Using curl

Test your webhook endpoints with sample payloads:

```bash
# Test Termii webhook
curl -X POST https://your-domain.com/api/v1/webhooks/dlr/termii \
  -H "Content-Type: application/json" \
  -d '{
    "type": "dlr",
    "message_id": "test_msg_123",
    "status": "Delivered"
  }'

# Test AfricaTalking webhook
curl -X POST https://your-domain.com/api/v1/webhooks/dlr/africatalking \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "id=test_msg_123&status=Success"

# Test Simpu webhook
curl -X POST https://your-domain.com/api/v1/webhooks/dlr/simpu \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message.status.update",
    "data": {
      "message_id": "test_msg_123",
      "status": "delivered"
    }
  }'
```

### Expected Response

All webhook endpoints return `200 OK` with JSON:
```json
{
  "status": "ok",
  "message": "DLR received"
}
```

Even if the message is not found in the database, the webhook returns 200 OK to prevent unnecessary retries from the provider.

---

## Troubleshooting

### Webhook Not Received

1. **Check firewall rules**: Ensure your API server accepts incoming connections from the provider's IP ranges
2. **Verify URL**: Double-check the webhook URL in the provider's dashboard
3. **Check logs**: Look for incoming webhook requests in your application logs
4. **Test locally**: Use ngrok or similar tools to test webhooks locally:
   ```bash
   ngrok http 8081
   # Use the ngrok URL as your API_BASE_URL
   ```

### Message Not Found

If webhooks are received but messages aren't updated:

1. **Check message ID mapping**: The `mno_message_id` in `message_segments` must match the ID sent by the provider
2. **Check logs**: Look for database query errors
3. **Verify timing**: Ensure the message was successfully sent before the DLR arrives

### Status Not Updated

1. **Check status normalization**: Provider statuses are normalized to standard values (delivered, sent, failed)
2. **Verify database query**: Check that `UpdateSegmentDLR` query is working correctly
3. **Check for errors**: Review application logs for any errors during DLR processing

---

## Security Considerations

### Webhook Signature Verification

Some providers support webhook signature verification:

- **Termii**: Supports HMAC signature verification (implementation exists but currently commented out)
- **Others**: Check provider documentation for signature verification support

To enable signature verification:

1. Store the provider's secret key in `mno_connections.secret_key`
2. Uncomment signature verification code in the provider's `ParseDLR()` method
3. Test thoroughly before deploying to production

### Rate Limiting

Consider adding rate limiting to webhook endpoints to prevent abuse:

```go
// Example using gin rate limiter middleware
webhookGroup.Use(ratelimit.NewMiddleware(100, time.Minute))
```

### IP Whitelisting

Restrict webhook endpoints to accept requests only from provider IP ranges. Check each provider's documentation for their webhook source IPs.

---

## Monitoring

### Key Metrics to Track

1. **Webhook success rate**: Percentage of webhooks successfully processed
2. **DLR latency**: Time between sending SMS and receiving DLR
3. **Provider comparison**: Compare DLR rates across providers
4. **Error rate**: Track failed webhook processing

### Recommended Monitoring Queries

```sql
-- DLR success rate by provider
SELECT 
  mc.base_url,
  COUNT(*) as total_segments,
  COUNT(ms.dlr_received_at) as dlr_received,
  ROUND(COUNT(ms.dlr_received_at)::NUMERIC / COUNT(*) * 100, 2) as dlr_rate_percent
FROM message_segments ms
JOIN mno_connections mc ON ms.mno_connection_id = mc.id
WHERE ms.sent_to_mno_at > NOW() - INTERVAL '24 hours'
GROUP BY mc.base_url;

-- Average DLR latency
SELECT 
  mc.base_url,
  AVG(EXTRACT(EPOCH FROM (ms.dlr_received_at - ms.sent_to_mno_at))) as avg_latency_seconds
FROM message_segments ms
JOIN mno_connections mc ON ms.mno_connection_id = mc.id
WHERE ms.dlr_received_at IS NOT NULL
  AND ms.sent_to_mno_at > NOW() - INTERVAL '24 hours'
GROUP BY mc.base_url;
```

---

## Provider API Documentation Links

- **Termii**: https://developers.termii.com/
- **AfricaTalking**: https://developers.africastalking.com/docs/sms/overview
- **Simpu**: https://developers.simpu.co/
- **eBulkSMS**: https://www.ebulksms.com/api-documentation
- **MultiTexter**: https://multitexter.com/developers

---

## Support

For issues with webhook configuration or DLR processing, check:

1. Application logs: `journalctl -u aegisbox-manager-api -f`
2. Database logs: Check PostgreSQL logs for query errors
3. Provider dashboard: Review webhook delivery status in provider's dashboard

For provider-specific issues, contact the provider's support team.
