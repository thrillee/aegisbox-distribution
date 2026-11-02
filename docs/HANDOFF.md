# Aegisbox handoff guide

Start here if you are picking up work on Aegisbox (person or AI agent). This guide reflects the
system as of 2026-11-02. `AGENTS.md` files in the three repos predate much of this; where they
disagree, this guide wins.

## 1. The three repositories

| Repo | What it is |
|---|---|
| [thrillee/aegisbox](https://github.com/thrillee/aegisbox) | Go backend. One binary `aegisbox` runs migrations, the Manager API, the SMPP gateway, the SP HTTP API and the MNO DLR receiver. Also `aegisbox mnosim`, a simulated MNO. |
| [thrillee/aegisbox-portal](https://github.com/thrillee/aegisbox-portal) | Next.js 16 admin and customer portal. Drizzle ORM against the same Postgres; Clerk for auth; calls the Manager API for some actions. |
| [thrillee/aegisbox-distribution](https://github.com/thrillee/aegisbox-distribution) | What gets deployed. Contains a **copy** of both app repos under `src/` (no submodules), the Dokploy compose file and `.env.example`. |

Deployment is on **Dokploy**, from `aegisbox-distribution` main, compose path
`./docker-compose.dokploy.yml`. Services: `aegisbox` (backend, `aegisbox all`), `portal`, and
`mnosim` (internal only, no published port). Monitoring (Grafana, Prometheus, Loki) was removed
from the Dokploy stack on purpose; its configs are still in the repo.

## 2. How changes are made (follow this exactly)

The owner's working rules:

1. **Verify, never assume.** Read the code before changing it, test against a real database, and say
   plainly what was and was not verified.
2. **Every change goes through a PR merged to `main`.** Work on a branch, open a PR with a clear
   description (what, why, verification), squash merge it.
3. **After merging app changes, sync the distribution:** in `aegisbox-distribution`, run
   `scripts/sync-sources.sh` (copies both repos' `main` into `src/` and updates `SOURCE_VERSIONS`),
   commit, PR, merge. The owner then redeploys on Dokploy.
4. **Never edit `src/` in the distribution by hand.** Change the app repo and sync.
5. Database changes are **goose migrations** in `aegisbox/db/migration` (embedded in the binary,
   run on start). Queries are **sqlc** (`db/query/*.sql`, generated into `internal/database`).
   Use sqlc **v1.23.0** (newer versions generate different code). Never edit generated files.
6. Only add migrations; never edit a migration that has been deployed.
7. Write for the owner in plain language; the owner prefers no dashes in prose.

## 3. Domain rules settled with the owner

These were decided deliberately; do not change them without asking.

### Money and billing
* Customers are charged when the message is sent. **A failed DLR does not refund automatically**;
  refunds are made by hand from the wallet (reversal transactions).
* `messages.total_segments` is the number of SMS segments **counted from the stored text**. It is
  what is billed and reported. Segment rows (`message_segments`) are the units actually submitted:
  one per SMPP segment, one per message for an HTTP API upstream. They can differ.
* Platform totals (admin dashboard, Message Analytics, Daily Report) **exclude test customers**
  (`service_providers.is_test`); the two reports have an "Include test customers" option.

### Time
* All reports use **Lagos days** (Africa/Lagos, UTC+1, no daylight saving). Portal helpers:
  `lib/time.ts` (`reportDayStart`, `reportDayRange`, `reportDateTime`), `lib/time-sql.ts`.

### Long SMS
* A long SMS sent over SMPP as several `submit_sm` with a concatenation UDH is stored part by part
  (`incoming_concat_parts`), each part answered with its own message ID, then **joined into one
  message** when all parts arrive. Missing parts: sent with what arrived after 2 minutes.
* The joined message keeps the other parts' IDs (`messages.extra_client_message_ids`); the SMPP DLR
  forwarder sends the DLR to **each** of them.

### DLRs
* **One final DLR per message** is forwarded to the customer, when every segment has a final
  status: the first failure's status and code, else DELIVRD. Intermediate statuses are not
  forwarded. Guarded by `dlr_forwarding_queue.dedupe_key`.
* A delivered receipt always carries `err:000`; upstream 3 character codes pass through.
* **Manual DLR** (portal Messages page, message detail, Pending DLRs page; API
  `POST /api/v1/messages/:id/dlr` and `/messages/dlr/bulk`) marks waiting segments with a status
  and forwards it. It never touches failed sends or refunded messages.
* **"Unconfirmed" / "no answer"** segments: sent, no MNO message ID, and either no error recorded
  (`TIMEOUT` is only the "no DLR in time" marker of the DLR timeout worker) **or** the submit timed
  out (`MNO_TIMEOUT`). Explicit rejections (with an error description) are never included. Manual
  DLR includes them only with `include_unconfirmed`.
* **Auto DLR** (per customer, DLR tab): after a delay, sends a chosen status for messages with no
  DLR. `include_unconfirmed` also covers unanswered submits. Ask the owner which customers use
  it and with which settings.
* A real DLR arriving **after** a manual or auto DLR is ignored (logged), so the customer never
  gets a second, contradicting DLR.

### Routing
* Weighted random choice between a network's usable routes (`routing_assignments.weight`),
  editable as a traffic split on the routing group page. Route health takes a route out after
  failed **submissions** (not DLR outcomes) and retries it after a minute.
* Route sender pools: per connection, a customer opted in has its sender replaced by one from the
  pool of the connection used.
* **Simulator guard:** an MNO with `is_simulator` only takes messages from customers with
  `is_test`; the sender refuses otherwise (fails closed) without charging route health.

### SMPP server (customer side)
* Up to 10 sessions per login; DLRs go to receiver or transceiver sessions in turn.
* Passwords are bcrypt; legacy plain values are upgraded on the next bind.
* Suspended, inactive or cancelled customers cannot bind or use the HTTP API.
* Every bind attempt is recorded in `smpp_sessions` (90 day retention); portal page
  Customers > SMPP Sessions.

## 4. Current state (all merged to main)

Highlights of recent work, newest first. See each repo's merged PRs for detail.

| Area | Status |
|---|---|
| Portal switches "Simulated MNO" and "Test customer"; test traffic out of totals | done (portal #35) |
| Long SMS: segments from text, SMPP part joining, one DLR per message | done (aegisbox #55) |
| Forwarded DLR error code `err:000` instead of `008` | done (aegisbox #54) |
| Simulated MNO `aegisbox mnosim` and simulator guard | done (aegisbox #53) |
| Late real DLR ignored after manual or auto DLR | done (aegisbox #52) |
| Auto DLR for unanswered submits | done (aegisbox #51, portal #34) |
| Pending DLRs page; Messages page DLR tools | done (portal #32, #33) |
| SMPP submit timeout recorded; error codes widened; DLR ack not answered with an error | done (aegisbox #47 to #50) |
| SMPP session history | done (aegisbox #45, #46; portal #30, #31) |
| Lagos time zone; analytics fixes; customer Connection page | done |
| SMPP security (sessions, bcrypt, bind responses, DLR port key) | done |

Migrations up to **00015** exist; they run automatically on start.

## 5. Open work

### 5.1 Link scanning (in progress; steps 1 and 1b done)
The full plan, including the gaps already resolved, is in [LINK_SCANNING_PLAN.md](LINK_SCANNING_PLAN.md).
Next is **step 2**: link extraction and the scan queue.

### 5.2 Waiting on the owner
* Two billing questions for one customer (a small test refund, and a one off charge for traffic
  sent outside Aegisbox): ask the owner whether they were done.
* Whether to add a per message Refund button.
* HTTP API keys are stored in plain text (`sp_credentials.api_key_identifier` is the bearer token).
  Moving to hashed keys changes how keys are issued; needs a decision.
* SMPP wrong password attempts are not rate limited (now visible in SMPP Sessions).

### 5.3 Known code debt
* `internal/mno` tests: `TestStatusNormalization`, `TestTermii_*`, `TestConnectionPooling` fail.
  They encode status mappings that were changed on purpose (for example Simpu "sent" means
  delivered, confirmed by the owner). Update the tests, not the code.
* `internal/managerapi/handlers` `TestTermiiWebhook_Success` fails on main (pre-existing).
* `gofmt` reports `internal/app/gateway.go`, `internal/sms/manager.go`,
  `internal/sms/otp_sender_reset_worker_test.go` (pre-existing).
* Portal `tsc` has 3 pre-existing errors: `app/admin/otp-alternative-senders/page.tsx`,
  `app/admin/routing/health/routes/page.tsx`, `app/admin/sender-ids/page.tsx`.
* SAR TLV concatenation (instead of UDH) and `message_payload` are not handled for SMPP submits.

## 6. Configuration

Backend (`aegisbox`, all `envconfig`, see `internal/config/config.go`): `DATABASE_URL` (required),
`MANAGER_API_KEY`, `API_ADDR`, `SERVER_HOST` (customer SMPP, container port 6886 in Dokploy),
`HTTP_CONFIG_ADDR` (customer HTTP API, 8000), `API_DLR_ADDR` (upstream DLR callbacks, 9001),
`API_DLR_KEY` (optional key required on DLR callbacks), `LOG_LEVEL`, worker intervals.

Simulator (`aegisbox mnosim`): `MNOSIM_ADDR`, `MNOSIM_SYSTEM_ID`, `MNOSIM_PASSWORD`,
`MNOSIM_DLR_DELAY_MIN/MAX`, `MNOSIM_DLR_MIX` (for example `DELIVRD:90,UNDELIV:5,EXPIRED:2,NONE:3`),
`MNOSIM_REJECT_PERCENT`, `MNOSIM_REJECT_STATUS`, `MNOSIM_NO_RESPONSE_PERCENT`.

Portal: `DATABASE_URL`, `CLERK_SECRET_KEY`, `CLERK_ADMIN_ORG_ID`,
`NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`, `NEXT_PUBLIC_CLERK_ADMIN_ORG_ID`, `NEXT_PUBLIC_APP_URL`
(build time), `MANAGER_API_URL`, `MANAGER_API_KEY`, `SMPP_PUBLIC_HOST`, `SMPP_PUBLIC_PORT`,
`HTTP_API_PUBLIC_URL`.

All are documented in `aegisbox-distribution/.env.example`. The Manager API routes are under
`/api/v1` and require `X-API-Key: $MANAGER_API_KEY`.

**Production database:** reachable only from the owner's environment, not from cloud agent
sessions. For production data questions, write read only SQL for the owner to run; for changes,
write a single transaction script that checks its preconditions and is safe to run twice (see
`docs/sql/`).

## 7. Testing locally (how everything above was verified)

See [LOCAL_TESTING.md](LOCAL_TESTING.md) for the full recipe: local Postgres, running the backend
and `mnosim`, SMPP test clients, running the portal against a local database with a Clerk stub,
and browser checks with Playwright.
