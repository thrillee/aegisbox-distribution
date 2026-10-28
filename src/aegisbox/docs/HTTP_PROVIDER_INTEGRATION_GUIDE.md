# HTTP Provider Integration Guide

## Overview

Aegisbox now supports **both SMPP and HTTP** protocols for sending SMS messages through Mobile Network Operators (MNOs). The routing system automatically selects the appropriate protocol based on database configuration.

---

## Architecture

### Key Components

1. **HTTPProvider Interface** - Provider-specific implementations (TERMII, AfricaTalking, Simpu, eBulkSMS, MultiTexter)
2. **HTTPProviderConnector** - Wrapper that bridges HTTPProvider with the Connector interface
3. **ProviderConfigLoader** - Loads provider configurations from database
4. **Manager** - Manages both SMPP and HTTP connectors, handles routing

### Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                      Service Provider                            │
│                   (Sends SMS via SMPP/HTTP)                      │
└────────────────────┬─────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                      SMS Processor                               │
│                  (Routes to MNO based on                         │
│                   intelligent routing rules)                     │
└────────────────────┬─────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                    DefaultSender.Send()                          │
│   connector, err := provider.GetConnector(ctx, mnoID)           │
└────────────────────┬─────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Manager.GetConnector()                          │
│                                                                  │
│  ┌──────────────────────────────────────────────────┐          │
│  │ 1. Looks up mnoID in connector map                │          │
│  │ 2. Filters active connectors:                     │          │
│  │    - SMPP: StatusBound                            │          │
│  │    - HTTP: StatusHttpOk                           │          │
│  │ 3. Selects connector (random if multiple)        │          │
│  └──────────────────────────────────────────────────┘          │
└────────────────────┬─────────────────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
        ▼                         ▼
┌──────────────────┐    ┌──────────────────────┐
│ SMPPMNOConnector │    │ HTTPProviderConnector│
│                  │    │                      │
│ - Sends via SMPP │    │ Wraps HTTPProvider   │
│ - Handles bind   │    │ - TERMII             │
│ - Segments SMS   │    │ - AfricaTalking      │
│                  │    │ - Simpu              │
└──────────────────┘    │ - eBulkSMS           │
                        │ - MultiTexter        │
                        └──────────────────────┘
```

---

## Database Configuration

### Step 1: Create MNO Entry

```sql
-- Create MNO entry
INSERT INTO mnos (name, country_code, description, is_active)
VALUES ('MTN Nigeria', 'NG', 'MTN Nigeria via TERMII', true)
RETURNING id;  -- Let's say this returns 1
```

### Step 2: Create HTTP Connection with Provider Config

```sql
-- Create HTTP connection with provider-specific configuration
INSERT INTO mno_connections (
    mno_id,
    protocol,
    status,
    base_url,
    provider_config,
    supports_webhook,
    webhook_path,
    timeout_secs
) VALUES (
    1,                          -- mno_id from step 1
    'http',                     -- Protocol: 'http' or 'smpp'
    'active',                   -- Status
    'https://api.ng.termii.com', -- Base URL
    '{
        "provider_name": "TERMII",
        "api_key": "TL12345abcdefghijklmnop",
        "secret_key": "webhook_secret_key_here"
    }'::jsonb,                  -- Provider-specific config
    true,                       -- Supports webhook DLRs
    '/webhooks/dlr/termii',     -- Webhook path
    30                          -- Timeout in seconds
);
```

### Step 3: Create SMPP Connection (Optional - for fallback)

```sql
-- Optionally add SMPP connection for the same MNO
INSERT INTO mno_connections (
    mno_id,
    protocol,
    status,
    system_id,
    password,
    host,
    port,
    bind_type,
    use_tls
) VALUES (
    1,                          -- Same mno_id
    'smpp',                     -- Protocol
    'active',                   -- Status
    'mtn_user',                 -- System ID
    'mtn_password',             -- Password
    'smpp.mtn.ng',              -- Host
    2775,                       -- Port
    'transceiver',              -- Bind type
    true                        -- Use TLS
);
```

---

## Supported HTTP Providers

### 1. TERMII

**Configuration**:
```json
{
  "provider_name": "TERMII",
  "api_key": "TL...",
  "secret_key": "webhook_secret"
}
```

**Base URL**: `https://api.ng.termii.com`

**Features**:
- ✅ Send SMS
- ✅ DLR webhooks
- ✅ Signature verification
- ✅ Bulk sending

---

### 2. AfricaTalking

**Configuration**:
```json
{
  "provider_name": "AFRICATALKING",
  "api_key": "atsk_...",
  "username": "sandbox"
}
```

**Base URL**: `https://api.africastalking.com/version1`

**Features**:
- ✅ Send SMS
- ✅ DLR webhooks
- ✅ Premium SMS
- ✅ Bulk sending

---

### 3. Simpu

**Configuration**:
```json
{
  "provider_name": "SIMPU",
  "api_key": "sk_...",
  "secret_key": "webhook_secret"
}
```

**Base URL**: `https://api.simpu.co/v1`

**Features**:
- ✅ Send SMS
- ✅ DLR webhooks
- ✅ Signature verification

---

### 4. eBulkSMS

**Configuration**:
```json
{
  "provider_name": "EBULKSMS",
  "username": "your_username",
  "api_key": "your_api_key"
}
```

**Base URL**: `https://api.ebulksms.com`

**Features**:
- ✅ Send SMS
- ✅ Balance check
- Limited DLR support

---

### 5. MultiTexter

**Configuration**:
```json
{
  "provider_name": "MULTITEXTER",
  "email": "user@example.com",
  "password": "your_password"
}
```

**Base URL**: `https://app.multitexter.com/v2`

**Features**:
- ✅ Send SMS
- ✅ Sender ID management
- Limited DLR support

---

## Message Flow

### Sending SMS via HTTP Provider

```go
// 1. Message arrives at processor
msg := database.GetMessageDetailsForSendingRow{
    ID:                     12345,
    RoutedMnoID:            &mnoID,  // MNO ID from routing
    OriginalSourceAddr:     "MyApp",
    OriginalDestinationAddr: "2348012345678",
    ShortMessage:            "Hello World",
    TotalSegments:           1,
}

// 2. DefaultSender gets connector from Manager
connector, err := s.mnoConnectorProvider.GetConnector(ctx, *msg.RoutedMnoID)
// Returns: *HTTPProviderConnector (wrapping TermiiProvider)

// 3. DefaultSender submits message
preparedMsg := mno.PreparedMessage{
    InternalMessageID: msg.ID,
    TotalSegments:     msg.TotalSegments,
    SenderID:          msg.OriginalSourceAddr,
    DestinationMSISDN: msg.OriginalDestinationAddr,
    MessageContent:    msg.ShortMessage,
    RequestDLR:        true,
    DBSeqs:            []int64{segmentID1, segmentID2},
}

result, err := connector.SubmitMessage(ctx, preparedMsg)

// 4. HTTPProviderConnector calls provider's SendSMS method
messageID, err := provider.SendSMS(ctx, "MyApp", "2348012345678", "Hello World")

// 5. Result maps provider message ID to database segments
result = SubmitResult{
    Segments: []SegmentSubmitInfo{
        {
            Seqn:         1,
            MnoMessageID: &messageID,  // "termii_msg_abc123"
            IsSuccess:    true,
            DBSegID:      &segmentID1,
        },
    },
    WasSubmitted: true,
}
```

---

## DLR (Delivery Report) Flow

### Webhook Configuration

```sql
-- Configure webhook path in mno_connections
UPDATE mno_connections
SET 
    supports_webhook = true,
    webhook_path = '/webhooks/dlr/termii'
WHERE id = 1;
```

### Webhook Endpoint

The Manager API exposes webhook endpoints:

```
POST /api/v1/webhooks/dlr/termii
POST /api/v1/webhooks/dlr/africatalking
POST /api/v1/webhooks/dlr/simpu
POST /api/v1/webhooks/dlr/ebulksms
POST /api/v1/webhooks/dlr/multitexter
POST /api/v1/webhooks/dlr/generic?provider=TERMII
```

### DLR Processing Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Provider sends DLR webhook                                    │
│    POST /api/v1/webhooks/dlr/termii                             │
│    {                                                             │
│      "message_id": "termii_msg_abc123",                         │
│      "status": "delivered",                                      │
│      "receiver": "2348012345678"                                 │
│    }                                                             │
└────────────────────┬─────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. WebhookHandler receives request                              │
│    - Verifies signature (if provider supports)                  │
│    - Extracts provider name from URL path                       │
└────────────────────┬─────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. ProviderConfigLoader loads provider                          │
│    provider, err := loader.GetProviderByName(ctx, "TERMII")     │
└────────────────────┬─────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. Provider parses DLR                                           │
│    messageID, status, err := provider.ParseDLR(ctx, req)        │
│    // Returns: "termii_msg_abc123", "delivered", nil            │
└────────────────────┬─────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. DLRForwarder finds segment and updates status                │
│    - Looks up segment by mno_message_id                         │
│    - Updates segment status to "delivered"                      │
│    - Records delivery timestamp                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Protocol Selection

The Manager automatically selects the appropriate protocol when `GetConnector()` is called.

### Selection Logic

```go
func (m *Manager) GetConnectorWithProtocol(ctx context.Context, mnoID int32, preferredProtocol string) (Connector, error) {
    // 1. Get all connectors for this MNO
    connIDs := m.mnoConnectors.Load(mnoID)
    
    // 2. Filter by status
    for _, connID := range connIDs {
        conn := m.connectors.Load(connID)
        status := conn.Status()
        
        // Check if connector is ready
        if status == codes.StatusBound {      // SMPP is bound
            availableConnectors = append(availableConnectors, conn)
        } else if status == codes.StatusHttpOk { // HTTP is ready
            availableConnectors = append(availableConnectors, conn)
        }
    }
    
    // 3. Select random connector if multiple available
    return availableConnectors[rand.Intn(len(availableConnectors))]
}
```

### Preferred Protocol

You can optionally specify a preferred protocol:

```go
// Prefer HTTP if available
connector, err := manager.GetConnectorWithProtocol(ctx, mnoID, "http")

// Prefer SMPP if available  
connector, err := manager.GetConnectorWithProtocol(ctx, mnoID, "smpp")

// No preference - use any active connector
connector, err := manager.GetConnector(ctx, mnoID)
```

---

## Configuration Priority

When loading HTTP provider configurations, values are merged with this priority:

1. **Highest**: Values in `provider_config` JSON
2. **Fallback**: Values in table columns

### Example

```sql
-- Database record
INSERT INTO mno_connections (
    mno_id,
    protocol,
    api_key,        -- Column value: "old_key_123"
    base_url,       -- Column value: "https://api.termii.com"
    provider_config -- JSON value: {"provider_name": "TERMII", "api_key": "new_key_456"}
) VALUES (1, 'http', 'old_key_123', 'https://api.termii.com', 
         '{"provider_name": "TERMII", "api_key": "new_key_456"}'::jsonb);

-- Resulting configuration used:
-- APIKey: "new_key_456"  ← From JSON (overrides column)
-- BaseURL: "https://api.termii.com"  ← From column (no JSON override)
```

---

## Testing

### Unit Tests

```bash
# Test provider interfaces
go test ./internal/mno/ -v -run TestProviderInterfaces

# Test config loader
go test ./internal/mno/ -v -run TestProviderConfigLoader

# Test connector wrapper
go test ./internal/mno/ -v -run TestHTTPProviderConnector
```

### Integration Testing

```bash
# 1. Start local database
docker-compose up -d postgres

# 2. Run migrations
make db/migrate

# 3. Insert test HTTP connection
psql $DATABASE_URL -c "
INSERT INTO mnos (name, country_code, is_active) VALUES ('Test MNO', 'NG', true);
INSERT INTO mno_connections (mno_id, protocol, status, base_url, provider_config)
VALUES (1, 'http', 'active', 'https://api.ng.termii.com',
        '{\"provider_name\": \"TERMII\", \"api_key\": \"test_key\"}'::jsonb);
"

# 4. Start gateway
make run/gateway

# 5. Send test message (should route to HTTP provider)
```

---

## Monitoring

### Metrics

HTTP provider connectors expose the same metrics as SMPP:

```
# Messages submitted via provider
sms_messages_submitted_total{protocol="http", provider="TERMII"}

# Messages failed
sms_messages_failed_total{protocol="http", provider="TERMII"}

# DLRs received
sms_dlrs_received_total{protocol="http", provider="TERMII"}

# Provider latency
sms_provider_latency_seconds{protocol="http", provider="TERMII"}
```

### Logs

```json
{
  "level": "info",
  "msg": "Submitting message via HTTP provider",
  "provider": "TERMII",
  "message_id": 12345,
  "destination": "2348012345678",
  "conn_id": 1,
  "mno_id": 1
}

{
  "level": "info",
  "msg": "Message submitted successfully via HTTP provider",
  "provider": "TERMII",
  "segment_count": 1,
  "message_id": "termii_msg_abc123"
}
```

---

## Troubleshooting

### Issue: HTTP provider not loading

**Symptoms**: Messages always route to SMPP even when HTTP configured

**Solution**:
```sql
-- Check provider_config has provider_name
SELECT id, protocol, provider_config 
FROM mno_connections 
WHERE protocol = 'http';

-- Should return: {"provider_name": "TERMII", ...}
-- If missing, update:
UPDATE mno_connections
SET provider_config = jsonb_set(
    COALESCE(provider_config, '{}'::jsonb),
    '{provider_name}',
    '"TERMII"'
)
WHERE id = 1;
```

### Issue: Webhook DLRs not processing

**Symptoms**: Messages show as "sent" but never "delivered"

**Solution**:
```bash
# 1. Check webhook path is configured
SELECT id, webhook_path, supports_webhook 
FROM mno_connections 
WHERE protocol = 'http';

# 2. Test webhook endpoint
curl -X POST http://localhost:8081/api/v1/webhooks/dlr/termii \
  -H "Content-Type: application/json" \
  -d '{"message_id": "test_123", "status": "delivered"}'

# 3. Check logs for signature verification errors
grep "webhook signature" logs/manager-api.log
```

### Issue: "unknown provider: TERMII"

**Symptoms**: Error when loading provider

**Solution**:
```sql
-- Provider name is case-insensitive, but must match exactly
-- Valid values: TERMII, AFRICATALKING, SIMPU, EBULKSMS, MULTITEXTER

-- Check current value
SELECT provider_config->>'provider_name' FROM mno_connections WHERE id = 1;

-- Fix if incorrect
UPDATE mno_connections
SET provider_config = jsonb_set(provider_config, '{provider_name}', '"TERMII"')
WHERE id = 1;
```

---

## Migration from SMPP-Only

If you currently have SMPP-only connections and want to add HTTP:

### Step 1: Add HTTP connection without removing SMPP

```sql
-- Keep existing SMPP connection active
-- Add new HTTP connection for same MNO
INSERT INTO mno_connections (
    mno_id,               -- Same as SMPP connection
    protocol,             -- 'http'
    status,               -- 'active'
    priority,             -- Lower than SMPP initially
    base_url,
    provider_config
) VALUES (
    (SELECT mno_id FROM mno_connections WHERE id = 1),  -- Copy MNO ID
    'http',
    'active',
    10,  -- Higher priority number = lower priority
    'https://api.ng.termii.com',
    '{"provider_name": "TERMII", "api_key": "TL..."}'::jsonb
);
```

### Step 2: Monitor both connections

```bash
# Watch logs for both protocols
tail -f logs/gateway.log | grep -E "protocol=(smpp|http)"

# Check metrics
curl http://localhost:9090/metrics | grep sms_messages_submitted_total
```

### Step 3: Gradually shift traffic to HTTP

```sql
-- Increase HTTP priority (lower number = higher priority)
UPDATE mno_connections SET priority = 1 WHERE protocol = 'http' AND mno_id = 1;

-- Decrease SMPP priority
UPDATE mno_connections SET priority = 10 WHERE protocol = 'smpp' AND mno_id = 1;
```

### Step 4: Disable SMPP when confident

```sql
-- Deactivate SMPP
UPDATE mno_connections SET status = 'inactive' WHERE protocol = 'smpp' AND mno_id = 1;
```

---

## Best Practices

### 1. **Use Circuit Breakers**

Circuit breakers are automatically enabled for both SMPP and HTTP connections. Configure thresholds:

```go
// In manager.go
newCB := NewCircuitBreaker(CircuitBreakerConfig{
    FailureThreshold: 5,    // Open after 5 failures
    SuccessThreshold: 3,    // Close after 3 successes
    Timeout:          30 * time.Second,
    RequestTimeout:   10 * time.Second,
})
```

### 2. **Set Appropriate Timeouts**

```sql
-- Short timeout for fast providers
UPDATE mno_connections SET timeout_secs = 10 WHERE provider_config->>'provider_name' = 'TERMII';

-- Longer timeout for slower providers
UPDATE mno_connections SET timeout_secs = 30 WHERE provider_config->>'provider_name' = 'AFRICATALKING';
```

### 3. **Secure Webhook Endpoints**

```go
// Webhook handler automatically verifies signatures for providers that support it
// Ensure secret_key is set in provider_config

UPDATE mno_connections
SET provider_config = jsonb_set(
    provider_config,
    '{secret_key}',
    '"your_webhook_secret"'
)
WHERE protocol = 'http';
```

### 4. **Monitor Provider Health**

```bash
# Use health check endpoint
curl http://localhost:8081/api/v1/routing/health/stats

# Response includes per-provider metrics
{
  "providers": {
    "TERMII": {
      "success_rate": 0.99,
      "avg_latency_ms": 450,
      "last_error": null
    }
  }
}
```

---

## Summary

✅ **HTTP providers are fully integrated** alongside SMPP
✅ **Automatic protocol selection** based on database configuration
✅ **5 providers supported**: TERMII, AfricaTalking, Simpu, eBulkSMS, MultiTexter
✅ **DLR webhooks** with signature verification
✅ **Circuit breakers** for reliability
✅ **Comprehensive logging and metrics**

The routing system treats HTTP and SMPP as equals - the best available connector is selected based on status, priority, and health.
