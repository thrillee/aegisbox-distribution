# Aegisbox SMS Gateway - Setup & Test Guide

## Overview
This is a complete SMS gateway platform with Admin Portal and Service Provider Portal. The system uses:
- **Database**: Neon PostgreSQL
- **Authentication**: Clerk
- **Backend**: Next.js App Router with API Routes
- **Frontend**: React with shadcn/ui components

## Architecture

### Database Schema
The system includes 9 core tables:
- `admins` - Platform administrators
- `service_providers` - SMS service providers
- `mnos` - Mobile Network Operators (MTN, Airtel, Glo, 9mobile)
- `provider_mno_mappings` - Links providers to MNOs
- `messages` - SMS messages sent through the platform
- `transactions` - Wallet transactions
- `wallets` - Provider wallet accounts
- `delivery_logs` - Message delivery tracking
- `api_keys` - Provider API credentials

### Application Flow

1. **Home Page** (`/`) - Landing page with login options
2. **Authentication** - Clerk handles sign-up/sign-in
3. **Auth Check** (`/auth-check`) - Routes users to correct portal
4. **Admin Dashboard** (`/admin/dashboard`) - View providers & MNOs
5. **Provider Dashboard** (`/provider/dashboard`) - Send SMS & manage wallet

## Setup Instructions

### 1. Environment Setup

Ensure these environment variables are set in your Vercel project:

```
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=your_clerk_publishable_key
CLERK_SECRET_KEY=your_clerk_secret_key
DATABASE_URL=your_neon_connection_string
```

### 2. Database Setup

The database schema was created from `/scripts/setup-database.sql`. If you need to reset:

```bash
# Execute the schema creation (already done)
psql < scripts/setup-database.sql

# Optional: Seed test data
psql < scripts/seed-test-data.sql
```

### 3. Clerk Configuration

1. Go to your Clerk dashboard
2. Create an application (if not done)
3. Copy your publishable and secret keys
4. Add these to your environment variables
5. Configure allowed redirect URIs to include your app URL

## Testing Instructions

### Test User Setup

First, you need to create two test users in Clerk:

1. **Admin User**
   - Email: `admin@aegisbox.com`
   - Role: Admin
   - After signing up, note the Clerk User ID

2. **Provider User**
   - Email: `provider@aegisbox.com`
   - Role: Service Provider
   - After signing up, note the Clerk User ID

### Manual Data Setup

Execute these SQL commands to register your test users:

```sql
-- Register admin user (replace 'user_abc123' with actual Clerk ID)
INSERT INTO admins (name, email, clerk_id, status, created_at)
VALUES ('Test Admin', 'admin@aegisbox.com', 'user_abc123', 'active', NOW());

-- Register provider user (replace 'user_xyz789' with actual Clerk ID)
INSERT INTO service_providers (name, email, clerk_id, business_type, wallet_balance, status, created_at)
VALUES ('Test Provider', 'provider@aegisbox.com', 'user_xyz789', 'bulk-sms', 100000, 'active', NOW());

-- Link provider to MNOs
INSERT INTO provider_mno_mappings (provider_id, mno_id, created_at)
SELECT sp.id, m.id, NOW()
FROM service_providers sp, mnos m
WHERE sp.email = 'provider@aegisbox.com';
```

### Test Flow

#### 1. Test Home Page
- Navigate to `/`
- Verify you see the Aegisbox SMS Gateway landing page
- See two login buttons: "Service Provider Login" and "Admin Login"

#### 2. Test Admin Portal
- Click "Admin Login"
- Sign in with admin@aegisbox.com
- Should redirect to `/admin/dashboard`
- Verify you can see:
  - Service Providers section (list of providers)
  - Mobile Network Operators section (list of MNOs with provider counts)

#### 3. Test Service Provider Portal
- Sign out from admin dashboard
- Go back to home page
- Click "Service Provider Login"
- Sign in with provider@aegisbox.com
- Should redirect to `/provider/dashboard`
- Verify you can see:
  - Wallet Balance (₦100000)
  - Total Messages count
  - Send SMS form
  - Recent Messages section

#### 4. Test SMS Sending
- In Provider Dashboard, fill the SMS form:
  - Recipient: `+2348012345678`
  - Message: `Test SMS message`
- Click "Send SMS"
- Verify:
  - Success message appears
  - Message appears in "Recent Messages"
  - Wallet balance decreases (cost = ₦5 per 160 characters)

#### 5. Test Wallet Logic
- Try sending SMS when wallet has insufficient balance:
  - Lower provider's wallet to ₦2
  - Try to send SMS
  - Should see error: "Insufficient wallet balance"

#### 6. Test Authentication
- Try accessing `/admin/dashboard` while logged in as provider
  - Should access successfully (user verification only)
- Try accessing `/provider/dashboard` without signing in
  - Should redirect to `/sign-in`
- Sign out and try accessing protected routes
  - Should redirect to home page

### API Endpoints

#### Admin Endpoints
- `GET /api/admin/providers` - List all service providers
- `GET /api/admin/mnos` - List all MNOs with provider counts

#### Provider Endpoints
- `GET /api/provider/data` - Get current provider data & wallet
- `GET /api/provider/messages` - Get recent messages (last 50)
- `POST /api/provider/send-sms` - Send new SMS message

### Test API with cURL

```bash
# Get admin providers (requires auth header)
curl -X GET http://localhost:3000/api/admin/providers \
  -H "Authorization: Bearer YOUR_CLERK_TOKEN"

# Get provider data
curl -X GET http://localhost:3000/api/provider/data \
  -H "Authorization: Bearer YOUR_CLERK_TOKEN"

# Send SMS
curl -X POST http://localhost:3000/api/provider/send-sms \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_CLERK_TOKEN" \
  -d '{
    "recipient": "+2348012345678",
    "content": "Test message"
  }'
```

## Common Issues & Solutions

### Issue: "Unauthorized" when accessing dashboards
**Solution**: Ensure you're signed in with a valid Clerk user

### Issue: "Provider not found" on provider dashboard
**Solution**: 
1. Get your Clerk User ID from Clerk Dashboard
2. Insert a record in service_providers table with your Clerk ID
3. Link to at least one MNO via provider_mno_mappings

### Issue: "Insufficient wallet balance" error
**Solution**: 
- Check wallet_balance in service_providers table
- Update balance: `UPDATE service_providers SET wallet_balance = 100000 WHERE clerk_id = 'your_id'`

### Issue: SMS form not submitting
**Solution**:
- Check browser console for errors
- Verify provider has active status in database
- Ensure wallet has sufficient balance

## Development Notes

- Hot reload works: edit files and changes appear immediately
- Database queries are async: use `await db.query()`
- Auth checks happen via Clerk middleware
- All user data is tied to Clerk User IDs for security

## Next Steps (Optional Enhancements)

1. Add wallet top-up functionality
2. Implement message delivery status tracking
3. Add SMS templates
4. Create detailed analytics dashboard
5. Add rate limiting for API endpoints
6. Implement audit logging
7. Add support for bulk SMS uploads
