# HTTP Provider Configuration Guide

This document describes how to configure HTTP SMS providers in the Aegisbox system using the `mno_connections` table.

## Overview

HTTP provider configurations are stored in the `provider_config` JSONB column of the `mno_connections` table. This allows for flexible, provider-specific configuration while maintaining a consistent database schema.

## Table of Contents

- [Database Schema](#database-schema)
- [Provider Config JSON Schema](#provider-config-json-schema)
- [Supported Providers](#supported-providers)
- [Configuration Examples](#configuration-examples)
- [API Usage](#api-usage)
- [Troubleshooting](#troubleshooting)

---

## Database Schema

### `mno_connections` Table

The `mno_connections` table stores connection details for both SMPP and HTTP providers.

**Key Fields for HTTP Providers**:

| Column | Type | Description |
|--------|------|-------------|
| `protocol` | VARCHAR(10) | Must be `'http'` for HTTP providers |
| `api_key` | TEXT | Provider API key (can be overridden in `provider_config`) |
| `base_url` | VARCHAR(255) | Provider API base URL |
| `username` | VARCHAR(255) | Username for auth (provider-specific) |
| `http_password` | TEXT | Password for auth (provider-specific) |
| `secret_key` | TEXT | Secret key for webhooks/signatures |
| `default_sender` | VARCHAR(100) | Default sender ID |
| `rate_limit` | INTEGER | Rate limit (requests per minute) |
| `timeout_secs` | INT | HTTP request timeout in seconds |
| `supports_webhook` | BOOLEAN | Whether provider supports DLR webhooks |
| `webhook_path` | VARCHAR(255) | Webhook callback path |
| `**provider_config**` | **JSONB** | **Provider-specific configuration (JSON)** |

---

## Provider Config JSON Schema

The `provider_config` JSONB field must contain provider-specific configuration.

### Required Fields

All HTTP providers must have these fields in `provider_config`:

```json
{
  "provider_name": "PROVIDER_NAME"
}
```

**`provider_name`** values (case-insensitive):
- `TERMII`
- `AFRICATALKING`
- `SIMPU`
- `EBULKSMS`
- `MULTITEXTER`

### Optional Fields

These fields can override values from the main table columns:

```json
{
  "provider_name": "TERMII",
  "api_key": "override_api_key_here",
  "secret_key": "override_secret_key_here"
}
```

**Priority**: Values in `provider_config` JSON take precedence over column values.

---

## Supported Providers

### 1. Termii

**Provider Name**: `TERMII`

**Required Configuration**:
```json
{
  "provider_name": "TERMII",
  "api_key": "your_termii_api_key"
}
```

**Optional Fields**:
```json
{
  "provider_name": "TERMII",
  "api_key": "your_termii_api_key",
  "secret_key": "webhook_secret_for_signature_verification"
}
```

**Base URL**: `https://api.ng.termii.com`

**Webhook Support**: Yes  
**DLR Signature Verification**: Yes (HMAC-SHA512)

**Full Example**:
```sql
INSERT INTO mno_connections (
    mno_id, protocol, status, priority,
    base_url, api_key, secret_key, default_sender,
    rate_limit, timeout_secs, supports_webhook, webhook_path,
    provider_config
) VALUES (
    1,                    -- mno_id
    'http',               -- protocol
    'active',             -- status
    1,                    -- priority
    'https://api.ng.termii.com',  -- base_url
    NULL,                 -- api_key (stored in provider_config instead)
    'webhook_secret_123', -- secret_key for DLR verification
    'YourBrand',          -- default_sender
    100,                  -- rate_limit
    30,                   -- timeout_secs
    true,                 -- supports_webhook
    '/webhook/termii',    -- webhook_path
    '{"provider_name": "TERMII", "api_key": "TL12345abcdef"}'::jsonb
);
```

---

### 2. Africa's Talking

**Provider Name**: `AFRICATALKING`

**Required Configuration**:
```json
{
  "provider_name": "AFRICATALKING",
  "api_key": "your_africastalking_api_key",
  "username": "your_africastalking_username"
}
```

**Base URL**: `https://api.africastalking.com/version1`

**Webhook Support**: Yes  
**DLR Signature Verification**: No

**Full Example**:
```sql
INSERT INTO mno_connections (
    mno_id, protocol, status, priority,
    base_url, username, default_sender,
    rate_limit, timeout_secs, supports_webhook, webhook_path,
    provider_config
) VALUES (
    2,                    -- mno_id
    'http',               -- protocol
    'active',             -- status
    1,                    -- priority
    'https://api.africastalking.com/version1',  -- base_url
    'sandbox',            -- username (can be overridden in provider_config)
    'YourBrand',          -- default_sender
    100,                  -- rate_limit
    30,                   -- timeout_secs
    true,                 -- supports_webhook
    '/webhook/africastalking',  -- webhook_path
    '{"provider_name": "AFRICATALKING", "api_key": "atsk_abc123", "username": "production"}'::jsonb
);
```

---

### 3. Simpu

**Provider Name**: `SIMPU`

**Required Configuration**:
```json
{
  "provider_name": "SIMPU",
  "api_key": "your_simpu_api_key"
}
```

**Base URL**: `https://api.simpu.co`

**Webhook Support**: Yes  
**DLR Signature Verification**: No

**Full Example**:
```sql
INSERT INTO mno_connections (
    mno_id, protocol, status, priority,
    base_url, default_sender,
    rate_limit, timeout_secs, supports_webhook, webhook_path,
    provider_config
) VALUES (
    3,                    -- mno_id
    'http',               -- protocol
    'active',             -- status
    1,                    -- priority
    'https://api.simpu.co',  -- base_url
    'YourBrand',          -- default_sender
    100,                  -- rate_limit
    30,                   -- timeout_secs
    true,                 -- supports_webhook
    '/webhook/simpu',     -- webhook_path
    '{"provider_name": "SIMPU", "api_key": "simpu_key_abc123"}'::jsonb
);
```

---

### 4. eBulkSMS

**Provider Name**: `EBULKSMS`

**Required Configuration**:
```json
{
  "provider_name": "EBULKSMS",
  "username": "your_ebulksms_username",
  "api_key": "your_ebulksms_api_key"
}
```

**Base URL**: `https://api.ebulksms.com`

**Webhook Support**: Yes  
**DLR Signature Verification**: No

**Full Example**:
```sql
INSERT INTO mno_connections (
    mno_id, protocol, status, priority,
    base_url, username, default_sender,
    rate_limit, timeout_secs, supports_webhook, webhook_path,
    provider_config
) VALUES (
    4,                    -- mno_id
    'http',               -- protocol
    'active',             -- status
    1,                    -- priority
    'https://api.ebulksms.com',  -- base_url
    'your_username',      -- username (can be overridden in provider_config)
    'YourBrand',          -- default_sender
    100,                  -- rate_limit
    30,                   -- timeout_secs
    true,                 -- supports_webhook
    '/webhook/ebulksms',  -- webhook_path
    '{"provider_name": "EBULKSMS", "username": "user@example.com", "api_key": "ebulk_key_123"}'::jsonb
);
```

---

### 5. MultiTexter

**Provider Name**: `MULTITEXTER`

**Required Configuration**:
```json
{
  "provider_name": "MULTITEXTER",
  "email": "your_email@example.com",
  "password": "your_password"
}
```

**Base URL**: `https://app.multitexter.com/v2`

**Webhook Support**: Yes  
**DLR Signature Verification**: No

**Full Example**:
```sql
INSERT INTO mno_connections (
    mno_id, protocol, status, priority,
    base_url, email, http_password, default_sender,
    rate_limit, timeout_secs, supports_webhook, webhook_path,
    provider_config
) VALUES (
    5,                    -- mno_id
    'http',               -- protocol
    'active',             -- status
    1,                    -- priority
    'https://app.multitexter.com/v2',  -- base_url
    'user@example.com',   -- email (can be overridden in provider_config)
    'password123',        -- http_password (can be overridden)
    'YourBrand',          -- default_sender
    100,                  -- rate_limit
    30,                   -- timeout_secs
    true,                 -- supports_webhook
    '/webhook/multitexter',  -- webhook_path
    '{"provider_name": "MULTITEXTER", "email": "api@example.com", "password": "secure_pass"}'::jsonb
);
```

---

## Configuration Examples

### Example 1: Simple Configuration

Store API key directly in `provider_config`:

```sql
UPDATE mno_connections 
SET provider_config = '{"provider_name": "TERMII", "api_key": "TL12345abcdef"}'::jsonb
WHERE id = 1;
```

### Example 2: Override Table Columns

Store API key in table column, but override in `provider_config`:

```sql
UPDATE mno_connections 
SET 
    api_key = 'old_key',  -- Stored but not used
    provider_config = '{"provider_name": "TERMII", "api_key": "new_key_override"}'::jsonb
WHERE id = 1;
```

**Result**: System uses `new_key_override` from `provider_config`.

### Example 3: Multiple Providers

Configure multiple providers for the same MNO:

```sql
-- Termii connection
INSERT INTO mno_connections (mno_id, protocol, base_url, priority, provider_config)
VALUES (1, 'http', 'https://api.ng.termii.com', 1, 
        '{"provider_name": "TERMII", "api_key": "termii_key"}'::jsonb);

-- AfricaTalking connection (same MNO, different provider)
INSERT INTO mno_connections (mno_id, protocol, base_url, priority, provider_config)
VALUES (1, 'http', 'https://api.africastalking.com/version1', 2, 
        '{"provider_name": "AFRICATALKING", "api_key": "at_key", "username": "sandbox"}'::jsonb);
```

**Result**: System can use either provider for MNO with ID 1. Priority determines order of failover.

---

## API Usage

### Loading Provider Configuration in Code

The `ProviderConfigLoader` handles loading providers from the database:

```go
import "github.com/thrillee/aegisbox/internal/mno"

// Create loader
loader := mno.NewProviderConfigLoader(dbQueries, logger)

// Load provider by name
provider, err := loader.GetProviderByName(ctx, "TERMII")
if err != nil {
    log.Fatal(err)
}

// Use provider
messageID, err := provider.SendSMS(ctx, "YourBrand", "+2348012345678", "Hello")
```

### Webhook Handler Usage

The webhook handler automatically loads provider config:

```go
// Webhook URL: /webhook/dlr?provider=TERMII
// Handler automatically loads Termii config from database and parses DLR
```

### DLR Forwarder Usage

The DLR forwarder automatically loads provider config:

```go
forwarder := mno.NewDLRForwarder(redisClient, dbQueries, logger)
err := forwarder.ProcessDLRWebhook(ctx, "TERMII", httpRequest)
```

---

## Configuration Validation

### Checking Current Configuration

```sql
-- View all HTTP provider configurations
SELECT 
    id,
    mno_id,
    protocol,
    base_url,
    provider_config->>'provider_name' as provider_name,
    provider_config->>'api_key' as api_key_from_json,
    api_key as api_key_from_column,
    status
FROM mno_connections
WHERE protocol = 'http';
```

### Validating JSON Schema

```sql
-- Check if provider_name exists
SELECT 
    id,
    CASE 
        WHEN provider_config->>'provider_name' IS NULL 
        THEN 'MISSING provider_name'
        ELSE 'OK'
    END as validation_status
FROM mno_connections
WHERE protocol = 'http';
```

### Finding Missing Configurations

```sql
-- Find HTTP connections without provider_config
SELECT id, mno_id, base_url, status
FROM mno_connections
WHERE protocol = 'http' 
  AND (provider_config IS NULL OR provider_config->>'provider_name' IS NULL);
```

---

## Updating Existing Connections

### Batch Update: Add provider_name

Update all Termii connections to include `provider_name`:

```sql
UPDATE mno_connections
SET provider_config = jsonb_set(
    COALESCE(provider_config, '{}'::jsonb),
    '{provider_name}',
    '"TERMII"'::jsonb
)
WHERE protocol = 'http' 
  AND base_url LIKE '%termii%'
  AND (provider_config IS NULL OR provider_config->>'provider_name' IS NULL);
```

### Batch Update: Move api_key to JSON

Move `api_key` from column to `provider_config`:

```sql
UPDATE mno_connections
SET provider_config = jsonb_set(
    COALESCE(provider_config, '{}'::jsonb),
    '{api_key}',
    to_jsonb(api_key)
)
WHERE protocol = 'http' 
  AND api_key IS NOT NULL
  AND (provider_config->>'api_key' IS NULL);
```

### Update Single Connection

```sql
UPDATE mno_connections
SET provider_config = jsonb_build_object(
    'provider_name', 'TERMII',
    'api_key', 'new_api_key_value',
    'secret_key', 'new_secret_key'
)
WHERE id = 1;
```

---

## Troubleshooting

### Issue: "Unknown provider" error

**Symptom**: `unknown provider: TERMII` error in logs

**Cause**: Missing or incorrect `provider_name` in `provider_config`

**Solution**:
```sql
-- Check provider_config
SELECT id, provider_config FROM mno_connections WHERE id = 1;

-- Add provider_name if missing
UPDATE mno_connections 
SET provider_config = jsonb_set(
    COALESCE(provider_config, '{}'::jsonb),
    '{provider_name}',
    '"TERMII"'::jsonb
)
WHERE id = 1;
```

### Issue: "Failed to load provider from database" error

**Symptom**: Error loading provider configuration

**Possible Causes**:
1. No active HTTP connection for the provider
2. Invalid JSON in `provider_config`
3. Provider name mismatch

**Solution**:
```sql
-- Check status
SELECT id, status, protocol, provider_config->>'provider_name' 
FROM mno_connections 
WHERE provider_config->>'provider_name' = 'TERMII';

-- Activate connection if inactive
UPDATE mno_connections 
SET status = 'active' 
WHERE id = 1;
```

### Issue: API key not working

**Symptom**: Authentication errors from provider API

**Cause**: API key in `provider_config` doesn't match provider account

**Solution**:
```sql
-- Update API key
UPDATE mno_connections
SET provider_config = jsonb_set(
    provider_config,
    '{api_key}',
    '"correct_api_key_here"'::jsonb
)
WHERE id = 1;
```

### Issue: DLR webhooks not parsing

**Symptom**: "Failed to parse DLR webhook" errors

**Possible Causes**:
1. Provider name mismatch in webhook URL
2. Missing API keys for signature verification (Termii)

**Solution**:
```sql
-- Verify webhook configuration
SELECT 
    id,
    provider_config->>'provider_name' as provider,
    provider_config->>'secret_key' as secret_key_configured,
    webhook_path
FROM mno_connections 
WHERE protocol = 'http';

-- Add secret_key if missing (for Termii)
UPDATE mno_connections
SET provider_config = jsonb_set(
    provider_config,
    '{secret_key}',
    '"your_webhook_secret"'::jsonb
)
WHERE provider_config->>'provider_name' = 'TERMII';
```

---

## Best Practices

### 1. Store Sensitive Data in provider_config

Store API keys and secrets in `provider_config` JSON instead of separate columns:

**Recommended**:
```json
{
  "provider_name": "TERMII",
  "api_key": "sensitive_key_here",
  "secret_key": "webhook_secret_here"
}
```

**Reasoning**: Easier to manage and backup as a single field.

### 2. Use Descriptive Provider Names

Always use the exact provider name in uppercase:

**Correct**: `"provider_name": "TERMII"`  
**Incorrect**: `"provider_name": "termii"` or `"provider_name": "Termii"`

**Reasoning**: Case-insensitive matching is handled by code, but consistency is better.

### 3. Set Appropriate Timeouts

Configure timeouts based on provider SLA:

```sql
UPDATE mno_connections 
SET timeout_secs = 30  -- 30 seconds for most providers
WHERE protocol = 'http';
```

### 4. Configure Webhooks

Always set `supports_webhook = true` and provide `webhook_path`:

```sql
UPDATE mno_connections
SET 
    supports_webhook = true,
    webhook_path = '/webhook/termii'
WHERE provider_config->>'provider_name' = 'TERMII';
```

### 5. Test Configuration

After updating configuration, test with a simple SMS:

```sql
-- Check configuration
SELECT 
    id,
    provider_config->>'provider_name' as provider,
    provider_config->>'api_key' IS NOT NULL as has_api_key,
    status
FROM mno_connections
WHERE id = 1;

-- Send test SMS via API to verify
```

---

## Security Considerations

### 1. API Key Storage

**Current**: API keys stored in plain text in database

**Recommendation**: 
- Use database encryption at rest
- Limit database access to authorized users only
- Consider using encrypted JSONB or external secrets manager

### 2. Webhook Signature Verification

**Termii**: Supports HMAC-SHA512 signature verification

**Setup**:
```sql
UPDATE mno_connections
SET provider_config = jsonb_set(
    provider_config,
    '{secret_key}',
    '"your_webhook_secret_from_termii"'::jsonb
)
WHERE provider_config->>'provider_name' = 'TERMII';
```

**Result**: System verifies webhook authenticity before processing.

### 3. Rate Limiting

Configure rate limits to prevent abuse:

```sql
UPDATE mno_connections
SET rate_limit = 100  -- 100 requests per minute
WHERE protocol = 'http';
```

---

## Migration Guide

### Migrating from Old Schema

If you have existing connections without `provider_config`:

**Step 1**: Add provider_name
```sql
UPDATE mno_connections
SET provider_config = jsonb_build_object('provider_name', 'TERMII')
WHERE base_url LIKE '%termii%' AND protocol = 'http';
```

**Step 2**: Move API key to JSON
```sql
UPDATE mno_connections
SET provider_config = jsonb_set(
    provider_config,
    '{api_key}',
    to_jsonb(api_key)
)
WHERE protocol = 'http' AND api_key IS NOT NULL;
```

**Step 3**: Verify migration
```sql
SELECT 
    id,
    provider_config->>'provider_name' as provider,
    provider_config->>'api_key' IS NOT NULL as has_api_key
FROM mno_connections
WHERE protocol = 'http';
```

---

## JSON Schema Reference

### Complete Schema

```json
{
  "type": "object",
  "required": ["provider_name"],
  "properties": {
    "provider_name": {
      "type": "string",
      "enum": ["TERMII", "AFRICATALKING", "SIMPU", "EBULKSMS", "MULTITEXTER"]
    },
    "api_key": {
      "type": "string",
      "description": "Provider API key (overrides api_key column)"
    },
    "secret_key": {
      "type": "string",
      "description": "Webhook secret key (overrides secret_key column)"
    },
    "username": {
      "type": "string",
      "description": "Username for authentication (provider-specific)"
    },
    "password": {
      "type": "string",
      "description": "Password for authentication (provider-specific)"
    },
    "email": {
      "type": "string",
      "description": "Email for authentication (MultiTexter)"
    }
  }
}
```

### Validation Examples

**Valid**:
```json
{"provider_name": "TERMII", "api_key": "key123"}
```

**Invalid** (missing required field):
```json
{"api_key": "key123"}
```

**Invalid** (unknown provider):
```json
{"provider_name": "UNKNOWN_PROVIDER"}
```

---

## Summary

- **Storage**: Provider configuration stored in `provider_config` JSONB column
- **Required Field**: `provider_name` must be present
- **Priority**: JSON values override table column values
- **Loading**: `ProviderConfigLoader` handles database loading
- **Usage**: Webhook and DLR handlers automatically use database config

**Benefits**:
- Flexible, provider-specific configuration
- Easy to add new providers
- Centralized configuration management
- No code changes needed for config updates
