# Quick Start Guide - Aegisbox SMS Gateway

## 60-Second Setup

### 1. Ensure Environment Variables Are Set
Go to your Vercel project settings and add:
```
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_live_...
CLERK_SECRET_KEY=sk_live_...
DATABASE_URL=postgresql://user:password@ep-xyz.neon.tech/dbname
```

### 2. Database Already Created
- Schema: ✅ Created via `/scripts/setup-database.sql`
- MNOs: ✅ Automatically populated (MTN, Airtel, Glo, 9mobile)

### 3. Test Admin Access

Run this SQL in your Neon database console:
```sql
INSERT INTO admins (name, email, clerk_id, status, created_at)
VALUES ('Admin Test', 'admin-test@example.com', 'user_YOUR_CLERK_ID_HERE', 'active', NOW());
```

Then:
1. Go to your app home page
2. Click "Admin Login"
3. Sign in or create account
4. Should see Admin Dashboard with providers/MNOs

### 4. Test Provider Access

Run this SQL:
```sql
INSERT INTO service_providers (name, email, clerk_id, business_type, wallet_balance, status, created_at)
VALUES ('Provider Test', 'provider-test@example.com', 'user_YOUR_CLERK_ID_HERE', 'bulk-sms', 100000, 'active', NOW());

-- Link to all MNOs
INSERT INTO provider_mno_mappings (provider_id, mno_id, created_at)
SELECT sp.id, m.id, NOW() FROM service_providers sp, mnos m 
WHERE sp.email = 'provider-test@example.com';
```

Then:
1. Go to home page
2. Click "Service Provider Login"
3. Sign in or create account
4. Should see Provider Dashboard with wallet balance

### 5. Send Test SMS

In Provider Dashboard:
1. Enter recipient: `+2348012345678`
2. Enter message: `Test SMS`
3. Click "Send SMS"
4. Verify:
   - Message appears in "Recent Messages"
   - Wallet balance decreased by ₦5

## Where to Find What

| Feature | File | URL |
|---------|------|-----|
| Home Page | `/app/page.tsx` | `/` |
| Admin Dashboard | `/app/admin/dashboard/page.tsx` | `/admin/dashboard` |
| Provider Dashboard | `/app/provider/dashboard/page.tsx` | `/provider/dashboard` |
| Admin Providers API | `/app/api/admin/providers/route.ts` | `/api/admin/providers` |
| Admin MNOs API | `/app/api/admin/mnos/route.ts` | `/api/admin/mnos` |
| Provider Data API | `/app/api/provider/data/route.ts` | `/api/provider/data` |
| Messages API | `/app/api/provider/messages/route.ts` | `/api/provider/messages` |
| Send SMS API | `/app/api/provider/send-sms/route.ts` | `/api/provider/send-sms` |

## Verify Everything Works

### Step 1: Check Home Page
```
GET /
Expected: Aegisbox landing page with login buttons
```

### Step 2: Check Admin Access
```
1. Click "Admin Login"
2. Sign in with admin account
Expected: Admin Dashboard loaded
```

### Step 3: Check Provider Access
```
1. Back to home, click "Service Provider Login"
2. Sign in with provider account
Expected: Provider Dashboard with ₦100000 wallet
```

### Step 4: Check SMS Sending
```
1. In Provider Dashboard, fill SMS form
2. Click "Send SMS"
Expected: Message sent, wallet decreased
```

### Step 5: Check API
```bash
curl -X GET http://localhost:3000/api/provider/data \
  -H "Cookie: __session=YOUR_AUTH_COOKIE"
Expected: Provider data JSON response
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "Unauthorized" error | Make sure you're signed in with Clerk |
| "Provider not found" | Add provider to database with your Clerk ID |
| Dashboard not loading | Check browser console for errors |
| SMS not sending | Check wallet balance (needs ≥ ₦5) |
| API 401 error | Verify DATABASE_URL is set correctly |

## Common SQL Queries

Get your Clerk User ID:
```
Look in Clerk Dashboard > Users > Select user > Under ID, copy the user_id
```

Check if admin exists:
```sql
SELECT * FROM admins WHERE email = 'your-email@example.com';
```

Check if provider exists:
```sql
SELECT * FROM service_providers WHERE email = 'your-email@example.com';
```

Update wallet balance:
```sql
UPDATE service_providers SET wallet_balance = 100000 
WHERE email = 'provider@example.com';
```

View sent messages:
```sql
SELECT recipient, content, status, created_at FROM messages 
ORDER BY created_at DESC LIMIT 10;
```

View transactions:
```sql
SELECT * FROM transactions ORDER BY created_at DESC LIMIT 10;
```

## Deploy to Production

1. Push code to GitHub (if using git)
2. Go to Vercel Dashboard
3. Click "Deploy" or connect repo for auto-deploy
4. Verify environment variables are set
5. Test admin & provider access after deploy
6. Monitor logs for errors

## Next: Full Documentation

For detailed setup, testing, and verification:
- Read `/SETUP_AND_TEST.md` - Complete guide with all details
- Read `/VERIFICATION_CHECKLIST.md` - Step-by-step testing
- Read `/BUILD_SUMMARY.md` - Architecture and features

## Support

If anything doesn't work:
1. Check browser console (F12 > Console tab)
2. Check server logs (Vercel Dashboard > Logs)
3. Verify environment variables are set
4. Verify database schema exists (check SETUP_AND_TEST.md)
5. Verify Clerk configuration matches your app

You're all set! 🚀
