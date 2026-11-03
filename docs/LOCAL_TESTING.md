# Testing Aegisbox locally

Every change in this project was verified end to end against a real Postgres, the real gateway
binary and, for SMPP, real SMPP clients. This is how.

## 1. Postgres
Any Postgres 16 works. Example with a throwaway data directory:

```bash
mkdir -p /var/tmp/aegispg && chown postgres /var/tmp/aegispg
su postgres -c "/usr/lib/postgresql/16/bin/initdb -D /var/tmp/aegispg/data"
su postgres -c "/usr/lib/postgresql/16/bin/pg_ctl -D /var/tmp/aegispg/data -o '-p 55432 -k /var/tmp/aegispg' -l /var/tmp/aegispg/log start"
psql -h /var/tmp/aegispg -p 55432 -U postgres -c "create database aegis"
```

If the machine restarted, remove a stale `postmaster.pid` before `pg_ctl start`.

## 2. Backend
```bash
cd aegisbox
go build -o /tmp/aegisbox ./cmd/aegisbox
export DATABASE_URL="postgres://postgres@/aegis?host=/var/tmp/aegispg&port=55432&sslmode=disable"
export API_ADDR=":18001" MANAGER_API_KEY=k SERVER_HOST="127.0.0.1:12775" \
       HTTP_CONFIG_ADDR="127.0.0.1:18000" API_DLR_ADDR="19001"
/tmp/aegisbox migrate
/tmp/aegisbox all          # Manager API :18001 (/api/v1), customer SMPP :12775, customer HTTP :18000
```

Run long lived processes with `nohup ... &` and an absolute binary path (a relative path in a
backgrounded `cd && ...` chain runs in the wrong directory). Stop them by PID; avoid
`pkill -f` patterns that also match your own shell.

Go checks: `go build ./...`, `go vet ./...`, `go test ./internal/...` (see the known failures in
HANDOFF.md). Regenerate queries with **sqlc v1.23.0** after editing `db/query/*.sql`.

## 3. Simulated MNO
```bash
MNOSIM_PASSWORD=simpass MNOSIM_DLR_DELAY_MIN=1s MNOSIM_DLR_DELAY_MAX=2s \
MNOSIM_DLR_MIX="DELIVRD:90,UNDELIV:10" nohup /tmp/aegisbox mnosim > /tmp/mnosim.log 2>&1 &
```
Then in the database (or the portal):
```sql
INSERT INTO mnos (name, country_code, status, is_simulator, created_at, updated_at)
VALUES ('Simulator', 'NG', 'active', true, now(), now()) RETURNING id;
-- An SMPP connection for it: host 127.0.0.1, port 2776, system_id simulator, password simpass,
-- bind_type trx (copy the other columns from an existing connection).
-- A test customer: service_providers.is_test = true, a routing group whose assignments point to
-- the Simulator MNO, a pricing rule, wallet credit, and an approved sender ID
-- (sender_ids.status = 'approved').
```
`mnosim` logs every submit (sender, recipient, text, part) and every DLR it sends.

## 3b. Link scanner

```bash
cd aegisbox/linkscanner && npm ci && npm test   # guard and scan tests against local fixtures
# Run it for the backend. Outside the Docker seccomp profile the Chromium sandbox may not start;
# only locally, add LINKSCAN_CHROMIUM_SANDBOX=false. LINKSCAN_ALLOW_NO_BLOCK_IPS=true is for local
# runs only; in production LINKSCAN_BLOCK_IPS must hold the server's public IP.
LINKSCAN_TOKEN=0123456789abcdef0123 LINKSCAN_ALLOW_NO_BLOCK_IPS=true PORT=8080 node src/server.js
```
Then start the backend with `LINKSCAN_ENABLED=true LINKSCAN_URL=http://127.0.0.1:8080
LINKSCAN_TOKEN=0123456789abcdef0123 LINKSCAN_SCREENSHOT_DIR=/tmp/linkscans`, send a message with a
link from the test customer, and read `link_scan_jobs` and `link_scans`. The scanner refuses
private addresses (only the tests allow 127.0.0.1, in code). In Docker, run it with the compose
settings (seccomp profile, read only, `cap_drop: ALL`) and check the log says `"sandbox":true`.

## 4. SMPP test clients
`docs/testing/smpp_send.py` (needs `pip install smpplib`) binds as a customer, sends messages
(single, multipart, a duplicated part, or only some parts) and prints the IDs and DLRs received:

```bash
python3 docs/testing/smpp_send.py --system-id CUST1 --password secret --mode multipart --wait 20
```

`docs/testing/fake_http_upstream.py` is a stand in HTTP upstream that records requests and returns
500 while a "fail" file exists (for failover and route health tests).

DLRs from an HTTP upstream can be posted to the gateway's DLR receiver:
```bash
curl -XPOST localhost:19001/aegis-dlr/callback -d '{"message_id":"<upstream id>","status":"DELIVRD"}'
```

## 5. Portal against a local database
The portal uses Clerk. For local tests, alias Clerk to a stub **temporarily** (never commit it):

1. Copy `tsconfig.json` and `lib/db.ts` somewhere safe.
2. Create `zz-clerk-stub/server.ts` and `zz-clerk-stub/client.tsx` exporting `auth`,
   `currentUser`, `clerkClient`, `clerkMiddleware`, `createRouteMatcher`, `ClerkProvider`,
   `SignIn`, `SignUp`, `SignOutButton`, `useUser`, `useOrganization`, `useOrganizationList`
   (`auth()` returns `{ userId, orgId }`, with `orgId` taken from an `x-test-org` header, default
   the admin org).
3. In `tsconfig.json` `paths`, map `@clerk/nextjs/server` and `@clerk/nextjs` to the stub files.
4. In `lib/db.ts` add `ssl: false` to the pool options (local Postgres has no TLS).
5. Run:
   ```bash
   rm -rf .next
   CLERK_ADMIN_ORG_ID=org_admin MANAGER_API_URL=http://localhost:18001 MANAGER_API_KEY=k \
   DATABASE_URL="postgres://postgres@localhost:55432/aegis?host=/var/tmp/aegispg" pnpm dev -p 3100
   ```
6. Customer pages: map an org to a customer with `INSERT INTO org_mappings (clerk_org_id,
   service_provider_id) VALUES ('org_cust', <id>)` and send `x-test-org: org_cust`.
7. **Restore `tsconfig.json` and `lib/db.ts` and delete `zz-clerk-stub/` before committing.**

Type check: `npx tsc --noEmit -p .` (3 known pre existing errors, see HANDOFF.md).
The project uses pnpm (version pinned in `package.json`).

## 6. Browser checks
Playwright with Chromium is the simplest way to click through pages and take screenshots; accept
`window.confirm` dialogs with `page.on('dialog', d => d.accept())`. Check pages at desktop width
and at 390 px for overflow.
