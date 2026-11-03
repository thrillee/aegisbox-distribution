# Turning on the simulated MNO and link scanning

This guide covers the whole path, start to finish: test traffic goes to the simulated MNO, every
link in it is opened in a locked-down browser and judged **safe, suspicious or malicious**, and
whitelisted numbers still get real delivery.

What happens to a message from a **test customer**:

1. Routing sends it to the simulated MNO (`mnosim`). The customer gets a normal DLR (the mix you
   set with `MNOSIM_DLR_MIX`) and is charged as usual. Nothing reaches a phone.
2. **Unless** the number is on the Simulator Whitelist: then it goes through a real network in the
   customer's routing group, the phone gets it and the real DLR comes back. These messages are not
   scanned (they did not go to the simulator).
3. The gateway pulls the links out of the full text and queues them (`link_scan_jobs`).
4. The scan worker sends each link to the `linkscanner` service, which opens it in Chromium (one
   fixed IP, identifies itself as `AegisboxLinkCheck/1.0`, never submits forms or clicks), and gets
   back the redirects, final page, forms, downloads and a screenshot.
5. The worker scores the result and stores it in `link_scans`, with every reason.

Customers that are **not** test customers never use the simulator: if their routing group also has
a simulated route, they skip it and use the real routes.

## 1. Before you start

* These must be merged and the distribution synced: aegisbox (link scanning steps 3 and 4,
  whitelist), aegisbox-portal (Simulator Whitelist page), and this repo's compose change.
* **Server memory:** Chromium needs about 1 to 2 GB. The scanner is limited to 2 GB
  (`LINKSCAN_MEMORY`). Check that the server has that free next to the other services.
* The first deploy pulls the Playwright image (about 1.5 GB) once.

## 2. Settings (Dokploy, Environment tab)

| Variable | Value | Why |
|---|---|---|
| `LINKSCAN_TOKEN` | a long random value, e.g. from `openssl rand -hex 32` | shared secret between aegisbox and the scanner (at least 16 characters) |
| `LINKSCAN_ENABLED` | `true` | starts the scan worker in aegisbox |
| `LINKSCAN_BLOCK_IPS` | the server's public IP, e.g. `203.0.113.10` (comma separate several) | **required**: so a scanned page can never call the server itself. The scanner stays idle until it is set. On the server: `curl -4 ifconfig.me` |
| `LINKSCAN_BLOCK_HOSTS` | optional extra host names to block | the **database host is blocked automatically** (aegisbox reads it from `DATABASE_URL`) |

Optional: `LINKSCAN_MAX_CONCURRENCY` (default 2, pages open at once; keep the same value for both
services), `LINKSCAN_MEMORY` (2g), `LINKSCAN_CPUS` (1.5), `LINKSCAN_SKIP_OTP` (true),
`LINKSCAN_SAMPLE_PER_HOUR` (3), `LINKSCAN_SCREENSHOT_RETENTION` (168h), `LINKSCAN_RESULT_RETENTION`
(672h). All are listed in `.env.example`.

Do **not** give the `linkscanner` service a domain or a published port in Dokploy. It must stay
alone on its own network.

Then **redeploy**.

## 3. Check that it runs

In Dokploy, open the logs:

* `linkscanner`: `linkscanner listening ... "sandbox":true`. If it says `linkscanner idle:`, the
  rest of the line names the missing setting (`LINKSCAN_TOKEN` or `LINKSCAN_BLOCK_IPS`).
* `aegisbox`: `Link scan worker started`. If it says
  `Link scanning is enabled but not configured`, the token is missing or too short.

## 4. Simulated MNO and test customer (portal)

Skip what is already set up.

1. **MNO Networks > Add**: name it `Simulator`, switch **Simulated MNO** on. Add an SMPP
   connection: host `mnosim`, port `2776`, system ID `simulator` (or your `MNOSIM_SYSTEM_ID`),
   password = `MNOSIM_PASSWORD`, bind type transceiver.
2. **Service Providers**: create or open `Aegisbox Test`, switch **Test customer** on. Give it an
   SMPP login (or HTTP key), a price, wallet credit and an approved sender ID.
3. **Routing**: give the test customer's login a routing group whose routes point to `Simulator`
   for the networks you test (MTN, Airtel, Glo, 9mobile).

## 5. Whitelisted numbers (real delivery)

1. In the **same routing group**, also add a **real** route for each network the whitelisted numbers
   are on. The traffic split only matters between routes of the same kind: test traffic uses the
   simulator routes, whitelisted numbers use the real ones.
2. Make sure the test customer has a price for the real network (or a default price) and credit:
   whitelisted messages are real messages and are charged normally.
3. **Network > Simulator Whitelist**: add the numbers. `08031234567` and `+234 803 123 4567` are both
   stored as `2348031234567`. Choose "All test customers" or one test customer.

If a whitelisted number's network has no real route in the group, the message fails with
"no usable route ... simulator routes only ...". Removing a number sends its test messages to the simulator again.

## 6. Send a test

Send from the test customer (SMPP or HTTP API) to a non whitelisted number, for example:

```
Your account is restricted. Verify at https://example.com/login
```

* **Messages** page: routed MNO `Simulator`, DLR `DELIVRD` (or as `MNOSIM_DLR_MIX` decides).
* `aegisbox` logs: `Link extraction for simulator message ... outcome=links`, then `Link scanned
  ... verdict=...` within a few seconds.

Then send to a whitelisted number: the phone receives it, and the routed MNO is the real network.

## 7. See the results

The portal page for scan results is the next step (step 5 of the plan). Until then, read only SQL:

```sql
-- Latest results with the message they came from
SELECT j.message_id, sp.name AS customer, j.normalized_url, s.verdict, s.score,
       s.final_url, s.scanned_at AT TIME ZONE 'Africa/Lagos' AS scanned_lagos,
       s.reasons
FROM link_scan_jobs j
JOIN link_scans s ON s.id = j.link_scan_id
JOIN service_providers sp ON sp.id = j.service_provider_id
ORDER BY s.scanned_at DESC
LIMIT 50;

-- Links not opened, and why (WhatsApp, tel, app store, unsubscribe/verify links, sampled)
SELECT message_id, normalized_url, link_kind, scan_mode, mode_reason, status
FROM link_scan_jobs WHERE scan_mode <> 'visit' ORDER BY id DESC LIMIT 50;

-- Queue health: waiting, retrying and failed jobs
SELECT status, COUNT(*), MAX(last_error) FROM link_scan_jobs GROUP BY status;

-- Messages checked: links, no_links, skipped_otp
SELECT outcome, COUNT(*) FROM message_link_extractions
WHERE extracted_at > NOW() - INTERVAL '1 day' GROUP BY outcome;
```

Screenshots are files on the `linkscan-data` volume (`/data/linkscans/<day>/scan-*.jpg` in the
`aegisbox` container); `link_scans.screenshot_path` is relative to that folder.

### What the verdict means

Score 70 or more is **malicious**, 30 to 69 **suspicious**, below 30 **safe**. Signals and points:

| Signal | Points |
|---|---|
| Domain names a Nigerian bank, fintech or telco it does not belong to (also punycode look alikes) | 40 |
| Page asks for card details, PIN, BVN/NIN, OTP or password (not on the brand's own domain) | 20 to 60 |
| Page uses a brand name and asks for secrets | 30 |
| Android app (APK) download / other risky download | 40 / 30 |
| Page or redirect reaches a private or internal address | 40 |
| IP address link, punycode domain | 25 each |
| Invalid HTTPS certificate | 20 |
| Unusual TLD (.xyz, .top, .click...) | 15 |
| No HTTPS, more than 3 redirects, final domain differs, auto submitting form, unreachable | 10 each |

A page that refuses the check (HTTP 403 or 451) or says it is not available in your country is
marked **"may hide content outside Nigeria"** and is never judged safe. The server is in the US;
if this shows up often, the fix is a fixed scanning location in Nigeria.

Not opened, only checked from the address itself (domain, look alikes, TLD): unsubscribe, opt out, confirm, verify and
approve links and links with one time tokens (a GET could act on them), and per recipient links
beyond 3 per campaign per hour (they share the visited ones' verdict). WhatsApp, `tel:`, app store
and `intent://` links are recorded only. OTP looking messages are skipped (`LINKSCAN_SKIP_OTP`).

## 8. Turning it off

* Stop opening links: `LINKSCAN_ENABLED=false`, restart `aegisbox`. Links keep being queued and are
  deleted after the result retention (4 weeks).
* Stop queueing links too: `LINKSCAN_EXTRACT_ENABLED=false`.
* Remove the token to leave the scanner idle (no browser).

## 9. Troubleshooting

| Symptom | Cause and fix |
|---|---|
| `linkscanner` exits with `Chromium failed to start` / `No usable sandbox` | The host blocks the seccomp profile or user namespaces. Check that Dokploy applies `security_opt` (`docker inspect` shows `seccomp=...`). On Ubuntu 23.10+ hosts, AppArmor may block unprivileged user namespaces: `sysctl kernel.apparmor_restrict_unprivileged_userns=0`. Do not turn the sandbox off in production. |
| aegisbox logs `Link scan postponed: scanner not available` | The scanner is down, starting, idle (`503`, see its log) or the token differs (`401`). Links wait and are scanned once it is back; none are lost. |
| `Link scan queue is growing: is the linkscanner service up?` | Scanner down, too slow, or out of memory. Check its logs; raise `LINKSCAN_MAX_CONCURRENCY` only if memory allows. |
| `could not save screenshot` | The `linkscan-data` volume is not writable; recreate it (it is created owned by the app user). |
| Whitelisted message failed with "no usable route" | Add a real route for that network to the test customer's routing group. |

## 10. Maintenance

* Update the Playwright image tag (`linkscanner/Dockerfile`) and the `playwright` version
  (`linkscanner/package.json`) together, monthly.
* Retention runs hourly: screenshots 7 days, results and jobs 4 weeks.
* Personal data: scanned pages and screenshots can show personal data (Nigeria Data Protection
  Act 2023). Keep access to the database and the volume restricted, keep retention short, and
  mention link scanning in customer terms.
