#!/bin/bash

# Webhook Test Script
# This script tests all HTTP provider webhook endpoints with sample payloads

set -e

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8081}"
WEBHOOK_BASE="${API_BASE_URL}/api/v1/webhooks/dlr"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "  HTTP Provider Webhook Tests"
echo "========================================="
echo "API Base URL: ${API_BASE_URL}"
echo ""

# Test counter
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to test webhook
test_webhook() {
    local provider=$1
    local content_type=$2
    local payload=$3
    local test_name=$4
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -n "Testing ${test_name}... "
    
    response=$(curl -s -w "\n%{http_code}" -X POST \
        "${WEBHOOK_BASE}/${provider}" \
        -H "Content-Type: ${content_type}" \
        -d "${payload}")
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ PASSED${NC} (HTTP $http_code)"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ FAILED${NC} (HTTP $http_code)"
        echo "  Response: $body"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

echo "========================================="
echo "1. Testing Termii Webhooks"
echo "========================================="

test_webhook "termii" "application/json" \
'{
  "type": "dlr",
  "id": "webhook_001",
  "message_id": "termii_test_msg_123",
  "receiver": "2348031234567",
  "sender": "TestBrand",
  "status": "Delivered",
  "sent_at": "2024-01-15 10:30:00",
  "cost": "0.50",
  "channel": "generic"
}' "Termii - Delivered Status"

test_webhook "termii" "application/json" \
'{
  "type": "dlr",
  "message_id": "termii_test_msg_124",
  "status": "Failed"
}' "Termii - Failed Status"

test_webhook "termii" "application/json" \
'{
  "type": "dlr",
  "message_id": "termii_test_msg_125",
  "status": "Sent"
}' "Termii - Sent Status"

echo ""
echo "========================================="
echo "2. Testing AfricaTalking Webhooks"
echo "========================================="

test_webhook "africatalking" "application/x-www-form-urlencoded" \
"id=at_test_msg_123&status=Success&phoneNumber=%2B2348031234567" \
"AfricaTalking - Success Status"

test_webhook "africatalking" "application/x-www-form-urlencoded" \
"id=at_test_msg_124&status=Failed&phoneNumber=%2B2348031234567" \
"AfricaTalking - Failed Status"

test_webhook "africatalking" "application/x-www-form-urlencoded" \
"id=at_test_msg_125&status=Sent&phoneNumber=%2B2348031234567" \
"AfricaTalking - Sent Status"

echo ""
echo "========================================="
echo "3. Testing Simpu Webhooks"
echo "========================================="

test_webhook "simpu" "application/json" \
'{
  "event": "message.status.update",
  "data": {
    "message_id": "simpu_test_msg_123",
    "status": "delivered",
    "timestamp": "2024-01-15T10:30:00Z",
    "callback": "https://example.com/webhook?conn_id=1&msg_id=12345"
  }
}' "Simpu - Delivered Status"

test_webhook "simpu" "application/json" \
'{
  "event": "message.status.update",
  "data": {
    "message_id": "simpu_test_msg_124",
    "status": "failed"
  }
}' "Simpu - Failed Status"

test_webhook "simpu" "application/json" \
'{
  "event": "message.status.update",
  "data": {
    "message_id": "simpu_test_msg_125",
    "status": "sent"
  }
}' "Simpu - Sent Status"

echo ""
echo "========================================="
echo "4. Testing eBulkSMS Webhooks"
echo "========================================="

test_webhook "ebulksms" "application/json" \
'{
  "message_id": "ebulk_test_msg_123",
  "status": "delivered",
  "recipient": "2348031234567",
  "delivery_time": "2024-01-15 10:30:00"
}' "eBulkSMS - Delivered Status"

test_webhook "ebulksms" "application/json" \
'{
  "message_id": "ebulk_test_msg_124",
  "status": "failed",
  "recipient": "2348031234567"
}' "eBulkSMS - Failed Status"

test_webhook "ebulksms" "application/json" \
'{
  "message_id": "ebulk_test_msg_125",
  "status": "sent",
  "recipient": "2348031234567"
}' "eBulkSMS - Sent Status"

echo ""
echo "========================================="
echo "5. Testing MultiTexter Webhooks"
echo "========================================="

test_webhook "multitexter" "application/json" \
'{
  "message_id": "mt_test_msg_123",
  "status": "1",
  "phone": "2348031234567",
  "timestamp": "2024-01-15 10:30:00"
}' "MultiTexter - Delivered Status (1)"

test_webhook "multitexter" "application/json" \
'{
  "message_id": "mt_test_msg_124",
  "status": "2",
  "phone": "2348031234567"
}' "MultiTexter - Failed Status (2)"

test_webhook "multitexter" "application/json" \
'{
  "message_id": "mt_test_msg_125",
  "status": "0",
  "phone": "2348031234567"
}' "MultiTexter - Sent Status (0)"

echo ""
echo "========================================="
echo "6. Testing Generic Webhook Endpoint"
echo "========================================="

test_webhook "generic?provider=TERMII&conn_id=1" "application/json" \
'{
  "type": "dlr",
  "message_id": "generic_test_msg_123",
  "status": "Delivered"
}' "Generic Endpoint - Termii"

test_webhook "generic?provider=AFRICATALKING&conn_id=1" "application/x-www-form-urlencoded" \
"id=generic_test_msg_124&status=Success" \
"Generic Endpoint - AfricaTalking"

echo ""
echo "========================================="
echo "  Test Summary"
echo "========================================="
echo "Total Tests: ${TOTAL_TESTS}"
echo -e "Passed: ${GREEN}${PASSED_TESTS}${NC}"
echo -e "Failed: ${RED}${FAILED_TESTS}${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed. Check the output above for details.${NC}"
    exit 1
fi
