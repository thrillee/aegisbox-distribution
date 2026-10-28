# Simpu OTP Integration Guide

## Overview

Simpu OTP is a specialized endpoint for sending OTP (One-Time Password) messages through Simpu's API. This integration allows you to have two separate MNO connections using Simpu:
1. **SIMPU** - Standard SMS endpoint (`/sms/send`)
2. **SIMPU_OTP** - OTP-specific endpoint (`/otp/request_token`)

## Key Features

- **Automatic OTP Extraction**: Automatically extracts 4-6 digit OTP codes from message text
- **Configurable Expiry**: OTP tokens expire in 15 minutes by default
- **Same Infrastructure**: Uses the same Simpu account with different API endpoints
- **Separate Routing**: Can be configured as separate MNO connections for routing control

## API Endpoint

**Endpoint**: `POST https://api.simpu.co/otp/request_token`

### Request Format
```json
{
  "recipient": "2348131302228",
  "length": 4,
  "expires_in": 15,
  "token": "123456",
  "channel": "sms"
}
```

### Response Format
```json
{
  "data": {
    "otp": {
      "created_datetime": "2026-02-03T09:13:27.431235+00:00",
      "expires_in": 15,
      "id": "86cf50500acf529f849babd3430ebe5c",
      "length": 4,
      "organisation_id": "20c7500d0f345e55b2353f2d60c7b1bc",
      "recipient": 2348131302228,
      "status": 1,
      "status_desc": "SENT",
      "valid_until": "2026-02-03T09:28:27.431235+00:00"
    }
  },
  "message": "Successful",
  "meta": {
    "charge": {
      "amount": 6.0,
      "available_balance": 3.7,
      "balance_before": 9.7,
      "encoding": "GSM"
    },
    "method": "POST",
    "resource": "/otp/request_token"
  },
  "status": 200,
  "type": "success"
}
```

## Setup Instructions

### 1. Database Configuration

You can create two separate MNO connections for Simpu using the same API credentials:

```sql
-- Create MNO for Simpu Standard SMS (if not exists)
INSERT INTO mnos (name, country_code, is_active)
VALUES ('Simpu SMS', 'NG', true)
RETURNING id;  -- Returns mno_id_1

-- Create MNO for Simpu OTP
INSERT INTO mnos (name, country_code, is_active)
VALUES ('Simpu OTP', 'NG', true)
RETURNING id;  -- Returns mno_id_2

-- Create Simpu Standard SMS Connection
INSERT INTO mno_connections (
    mno_id,
    protocol,
    status,
    base_url,
    provider_config,
    default_sender,
    supports_webhook,
    webhook_path
) VALUES (
    1,  -- mno_id_1 from above
    'http',
    'active',
    'https://api.simpu.co',
    '{"provider_name": "SIMPU", "api_key": "sk_live_your_api_key_here", "secret_key": "webhook_secret"}'::jsonb,
    'YourApp',
    true,
    '/api/v1/webhooks/dlr/simpu'
);

-- Create Simpu OTP Connection
INSERT INTO mno_connections (
    mno_id,
    protocol,
    status,
    base_url,
    provider_config,
    default_sender,
    supports_webhook,
    webhook_path
) VALUES (
    2,  -- mno_id_2 from above
    'http',
    'active',
    'https://api.simpu.co',
    '{"provider_name": "SIMPU_OTP", "api_key": "ssk__ctrI4I7Ksyd1eJxb5zBfoxfI5yf8yawFkZoR4O9EcPRQYHAxXD1xR163K7gCeRyt8ZeToKNuhzEEUKOq4K9E9", "secret_key": "webhook_secret"}'::jsonb,
    'YourApp',
    true,
    '/api/v1/webhooks/dlr/simpu-otp'
);
```

### 2. Provider Configuration

The `provider_config` JSON must include:

```json
{
  "provider_name": "SIMPU_OTP",
  "api_key": "ssk__your_simpu_api_key_here",
  "secret_key": "webhook_secret_for_dlr_verification"
}
```

**Configuration Fields**:
- `provider_name`: Must be exactly `"SIMPU_OTP"` (case-sensitive)
- `api_key`: Your Simpu API key (starts with `ssk__` for OTP endpoint)
- `secret_key`: Optional webhook secret for DLR verification

### 3. Webhook Configuration

Configure webhook URL in Simpu dashboard:

**Webhook URL**: `https://your-domain.com/api/v1/webhooks/dlr/simpu-otp?conn_id={connection_id}`

Replace:
- `your-domain.com` with your actual domain
- `{connection_id}` with the MNO connection ID from database

### 4. Message Format

The system automatically extracts OTP codes from messages:

**Supported Message Formats**:
```
"Your OTP is 1234"
"Your verification code: 123456"
"Code 54321 expires in 10 minutes"
"Use 9876 to login"
```

**OTP Extraction Rules**:
- Looks for 4-6 digit numbers in the message
- Uses regex pattern: `\b(\d{4,6})\b`
- If no OTP found, uses default 4-digit code "1234"

## How It Works

### Sending Process

1. **Message Received**: Gateway receives SMS with OTP code
   ```
   "Your verification code is 123456"
   ```

2. **OTP Extraction**: Provider automatically extracts OTP
   ```go
   Token: "123456"
   Length: 6
   ```

3. **API Request**: Sends to Simpu OTP endpoint
   ```json
   {
     "recipient": "2348131302228",
     "length": 6,
     "expires_in": 15,
     "token": "123456",
     "channel": "sms"
   }
   ```

4. **Response**: Returns OTP ID for tracking
   ```json
   {
     "data": {
       "otp": {
         "id": "86cf50500acf529f849babd3430ebe5c",
         "status_desc": "SENT"
       }
     }
   }
   ```

### DLR (Delivery Receipt) Processing

1. Simpu sends webhook to: `/api/v1/webhooks/dlr/simpu-otp?conn_id=2`
2. System extracts OTP ID and delivery status
3. Updates segment status in database
4. Returns 200 OK to Simpu

## Use Cases

### Scenario 1: Route OTP vs Regular SMS

Configure routing rules to send OTP messages through SIMPU_OTP and regular messages through SIMPU:

```sql
-- Create routing rules
INSERT INTO routing_rules (priority, mno_id, pattern, is_active)
VALUES
  (10, 2, '%OTP%', true),           -- OTP messages to Simpu OTP
  (10, 2, '%verification code%', true), -- Verification codes to Simpu OTP
  (20, 1, '%', true);                -- All other messages to Simpu Standard
```

### Scenario 2: Load Balancing

Use both connections for redundancy:
- Primary: Simpu OTP for OTP messages
- Secondary: Simpu Standard SMS for regular messages
- Automatic failover between connections

### Scenario 3: Cost Optimization

If Simpu OTP has different pricing, route specific message types accordingly.

## Testing

### 1. Test OTP Message

```bash
curl -X POST http://localhost:8080/send \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "from=YourApp&to=2348012345678&message=Your%20verification%20code%20is%201234"
```

### 2. Verify in Logs

```bash
tail -f logs/gateway.log | grep "Simpu OTP"
```

Expected output:
```
INFO Sending OTP via Simpu OTP API from=YourApp to=2348012345678
INFO OTP sent successfully via Simpu OTP API otp_id=86cf50500acf529f849babd3430ebe5c
```

### 3. Check Database

```sql
-- Verify segment was created and sent
SELECT id, message_id, status, external_message_id
FROM segments
WHERE external_message_id LIKE '86cf50500acf529f849babd3430ebe5c'
ORDER BY created_at DESC
LIMIT 1;
```

### 4. Test Webhook

```bash
curl -X POST http://localhost:8081/api/v1/webhooks/dlr/simpu-otp?conn_id=2 \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "object": {
        "transaction_id": "86cf50500acf529f849babd3430ebe5c",
        "status": "delivered"
      }
    },
    "type": "message.status.update"
  }'
```

## Comparison: SIMPU vs SIMPU_OTP

| Feature | SIMPU | SIMPU_OTP |
|---------|-------|-----------|
| **Endpoint** | `/sms/send` | `/otp/request_token` |
| **Use Case** | General SMS | OTP/Verification codes |
| **API Key** | `sk_live_...` | `ssk__...` |
| **Request Format** | Standard SMS fields | OTP-specific fields |
| **Auto-extract OTP** | No | Yes |
| **Expiry Management** | N/A | 15 minutes default |
| **Message Tracking** | Message ID | OTP ID |

## Troubleshooting

### Issue: "No OTP found in message"

**Solution**: Ensure message contains 4-6 digit number:
```
✓ "Your OTP is 1234"
✓ "Code: 123456"
✗ "Your code is ABC123"  (contains letters)
✗ "PIN: 123"             (only 3 digits)
```

### Issue: "API request failed with status 401"

**Solution**: Verify API key starts with `ssk__` for OTP endpoint (not `sk_live_`)

### Issue: "Webhook not received"

**Solution**: 
1. Check webhook URL includes correct `conn_id`
2. Verify webhook path in database: `/api/v1/webhooks/dlr/simpu-otp`
3. Check firewall allows incoming webhooks from Simpu IPs

### Issue: "OTP messages going to wrong connection"

**Solution**: Check routing rules priority and pattern matching

## Configuration Reference

### Minimal Configuration

```json
{
  "provider_name": "SIMPU_OTP",
  "api_key": "ssk__your_api_key_here"
}
```

### Full Configuration

```json
{
  "provider_name": "SIMPU_OTP",
  "api_key": "ssk__ctrI4I7Ksyd1eJxb5zBfoxfI5yf8yawFkZoR4O9EcPRQYHAxXD1xR163K7gCeRyt8ZeToKNuhzEEUKOq4K9E9",
  "secret_key": "webhook_secret_key_for_verification"
}
```

### Connection-Level Settings

```sql
-- In mno_connections table
base_url: 'https://api.simpu.co'
default_sender: 'YourApp'
timeout_secs: 30
rate_limit: 10  -- messages per second
supports_webhook: true
webhook_path: '/api/v1/webhooks/dlr/simpu-otp'
```

## Implementation Details

### Code Structure

- **Provider**: `internal/mno/http_providers/simpu_otp.go`
- **Factory**: `internal/mno/http_providers/common.go` (NewHTTPProvider)
- **Webhook Handler**: `internal/managerapi/handlers/webhook.go`
- **Routes**: `internal/managerapi/handlers/routes.go`
- **Tests**: `internal/mno/http_providers_test.go`

### OTP Extraction Logic

```go
// Extracts 4-6 digit OTP from message
func extractOTPFromMessage(message string) (string, int) {
    re := regexp.MustCompile(`\b(\d{4,6})\b`)
    matches := re.FindStringSubmatch(message)
    if len(matches) > 1 {
        token := matches[1]
        return token, len(token)
    }
    return "", 0
}
```

### Request Structure

```go
type SimpuOTPSendRequest struct {
    Recipient string `json:"recipient"`  // Phone number (digits only)
    Length    int    `json:"length"`     // OTP length (4-6)
    ExpiresIn int    `json:"expires_in"` // Minutes (default: 15)
    Token     string `json:"token"`      // The OTP code
    Channel   string `json:"channel"`    // Always "sms"
}
```

## Security Considerations

1. **API Key Protection**: Store API keys in database `provider_config` (not in code)
2. **Webhook Verification**: Use `secret_key` to verify webhook authenticity
3. **Rate Limiting**: Configure appropriate `rate_limit` to prevent abuse
4. **HTTPS Only**: Ensure webhook URLs use HTTPS in production

## Additional Resources

- [Simpu API Documentation](https://docs.simpu.co/)
- [Main HTTP Provider Guide](HTTP_PROVIDER_INTEGRATION_GUIDE.md)
- [Webhook Setup Guide](HTTP_PROVIDER_WEBHOOK_SETUP.md)
- [Quick Reference](HTTP_PROVIDER_QUICK_REFERENCE.md)

## Support

For issues specific to Simpu OTP integration:
1. Check logs: `tail -f logs/gateway.log | grep "Simpu OTP"`
2. Verify database configuration
3. Test webhook endpoint manually
4. Check Simpu dashboard for API status

For general HTTP provider issues, see [HTTP_PROVIDER_INTEGRATION_GUIDE.md](HTTP_PROVIDER_INTEGRATION_GUIDE.md)
