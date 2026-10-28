# HTTP Provider Quick Reference

## Adding a New HTTP Provider

### 1. Database Setup

```sql
-- Create MNO
INSERT INTO mnos (name, country_code, is_active)
VALUES ('MTN Nigeria', 'NG', true)
RETURNING id;  -- Returns mno_id

-- Create HTTP Connection
INSERT INTO mno_connections (
    mno_id,
    protocol,
    status,
    base_url,
    provider_config
) VALUES (
    1,  -- mno_id from above
    'http',
    'active',
    'https://api.ng.termii.com',
    '{"provider_name": "TERMII", "api_key": "YOUR_API_KEY", "secret_key": "WEBHOOK_SECRET"}'::jsonb
);
```

### 2. Supported Providers

| Provider | provider_name | Required Fields |
|----------|---------------|-----------------|
| TERMII | `TERMII` | `api_key`, `secret_key` |
| AfricaTalking | `AFRICATALKING` | `api_key`, `username` |
| Simpu | `SIMPU` | `api_key`, `secret_key` |
| Simpu OTP | `SIMPU_OTP` | `api_key`, `secret_key` |
| eBulkSMS | `EBULKSMS` | `username`, `api_key` |
| MultiTexter | `MULTITEXTER` | `email`, `password` |

### 3. Verification

```bash
# Check connection loaded
curl http://localhost:8081/api/v1/mnos

# Send test message
curl -X POST http://localhost:8080/send \
  -d "from=TestApp&to=2348012345678&message=Hello"

# Check logs
tail -f logs/gateway.log | grep "HTTP provider"
```

## Message Flow Summary

```
SMS Request → Processor → Router → MNO ID
                                      ↓
                          Manager.GetConnector(mnoID)
                                      ↓
                    ┌─────────────────┴─────────────────┐
                    │                                   │
              Status=Bound                      Status=HttpOk
                    │                                   │
            SMPPMNOConnector                  HTTPProviderConnector
                    │                                   │
              SMPP Protocol                    provider.SendSMS()
                    │                                   │
                    └─────────────────┬─────────────────┘
                                      ↓
                              SubmitResult
                                      ↓
                          Update Database Segments
```

## Configuration Examples

### TERMII
```json
{
  "provider_name": "TERMII",
  "api_key": "TL12345abcdefghijklmnop",
  "secret_key": "webhook_secret_key_here"
}
```

### AfricaTalking
```json
{
  "provider_name": "AFRICATALKING",
  "api_key": "atsk_abc123...",
  "username": "sandbox"
}
```

### Simpu
```json
{
  "provider_name": "SIMPU",
  "api_key": "sk_live_abc123...",
  "secret_key": "webhook_secret"
}
```

### Simpu OTP
```json
{
  "provider_name": "SIMPU_OTP",
  "api_key": "ssk__ctrI4I7Ksyd1eJxb5zBfoxfI5yf8yawFkZoR4O9EcPRQYHAxXD1xR163K7gCeRyt8ZeToKNuhzEEUKOq4K9E9",
  "secret_key": "webhook_secret"
}
```

**Note**: Simpu OTP automatically extracts OTP codes from messages (4-6 digits). The OTP endpoint (`/otp/request_token`) is optimized for sending verification codes.

## Common Commands

```bash
# Build
make build

# Run gateway
make run/gateway

# Run manager API
make run/manager-api

# Run tests
go test ./internal/mno/ -v

# Check metrics
curl http://localhost:9090/metrics | grep sms_
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| "unknown provider" | Check `provider_name` is one of: TERMII, AFRICATALKING, SIMPU, SIMPU_OTP, EBULKSMS, MULTITEXTER |
| "no connectors available" | Verify `status='active'` and provider_config is valid JSON |
| Webhook not working | Check `webhook_path` and `supports_webhook=true` |
| Messages fail | Verify API key is correct and provider account has credit |

## Key Files

```
internal/mno/
  ├── provider_config_loader.go      # Loads configs from DB
  ├── http_provider_connector.go     # Wraps HTTPProvider
  ├── http_providers.go               # Provider implementations
  ├── manager.go                      # Manages SMPP + HTTP
  └── interfaces.go                   # Connector interface

internal/managerapi/handlers/
  └── webhook.go                      # DLR webhook handler

docs/
  ├── HTTP_PROVIDER_INTEGRATION_GUIDE.md         # Complete guide
  ├── PHASE_6_IMPLEMENTATION_SUMMARY.md          # Implementation summary
  └── HTTP_PROVIDER_CONFIGURATION.md             # Config reference
```

## Status Codes

| Status | Meaning | Applies To |
|--------|---------|------------|
| `bound` | SMPP connected | SMPP only |
| `http_ok` | HTTP ready | HTTP only |
| `disconnected` | Not connected | Both |
| `error` | Error state | Both |

## Protocol Selection

The Manager automatically selects the best available protocol:

1. Gets all connectors for MNO ID
2. Filters by status (`bound` for SMPP, `http_ok` for HTTP)
3. Returns active connector (random if multiple)

No configuration needed - it just works!
