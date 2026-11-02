# Link scanning plan

Goal: open every link in messages routed to the simulated MNO in a real browser, judge it
**safe, suspicious or malicious** with reasons and a screenshot, and report it in the portal.
Later (Phase 2, separate approval) the same verdicts can hold or block real messages.

## Ground rules (decided with the owner; do not change)
* This is a **spam and phishing check**, not engagement. Scans never pose as recipients, never
  count as clicks, and use **one fixed IP** (the server's). **No IP rotation**, no residential
  proxies, no hiding automation.
* The scanner identifies itself: a normal mobile Chrome user agent plus `AegisboxLinkCheck/1.0`.
* No Google Safe Browsing Lookup API (non commercial only). Start with our own checks;
  add URLhaus only after confirming its terms for commercial use; Google Web Risk is the paid
  option if ever needed.
* The server is in the US. Pages that look different from what the link suggests or block foreign
  visitors are marked **"may hide content outside Nigeria"**, never "safe". If this is common, the
  fix is one fixed scanning location in Nigeria or a commercial scanner, still one honest IP.
* Retention: screenshots **1 week**, scan results **4 weeks** (configurable).
* Test traffic uses a dedicated **"Aegisbox Test"** customer marked `is_test`.

## Steps

| Step | What | State |
|---|---|---|
| 1 | `aegisbox mnosim` and the simulator guard | **done** (aegisbox #53) |
| 1b | Portal switches, badges, test traffic out of totals | **done** (portal #35) |
| 2 | Link extraction and scan queue | **done** (aegisbox #57) |
| 3 | `linkscanner` service with SSRF protection and isolation | next |
| 4 | Scoring, reputation, cache | |
| 5 | Portal "Link Scans" page | |
| 6 | End to end and load tests, deployment | |

### Step 2: extraction and queue (gateway)
* **Where:** in the gateway, when a submit to a **simulator MNO** is recorded as sent (the
  `UpdateSegmentSent` path or right after a successful `Send`). Not from `mnosim`: its submit
  response can arrive before the gateway stores the ID, so a callback from it would race.
* Scan the **full message text** (`messages.short_message`), never single segments, so URLs split
  across parts are found.
* **Extraction:** `http(s)://`, `www.`, bare domains (`bit.ly/x`), IP links, punycode. Accept a
  domain only if its suffix is in the public suffix list (avoids "e.g.", "N5.000", "Mr.Smith").
  Handle GSM 7 and UCS2 text; strip trailing punctuation. Record but do not open non web links:
  `wa.me`/WhatsApp, `tel:`, `intent://`, app store links.
* Skip OTP messages by default (configurable). Messages without links are marked "no links".
* Queue table `link_scan_jobs` (message ID, URL, status, attempts, error, times), processed with
  `FOR UPDATE SKIP LOCKED`, 3 retries with backoff, unique (message ID, URL).
* Personal links (a unique token per recipient): group by domain and path and visit only a sample
  (for example 3 per campaign per hour); check the rest by reputation only.
* Do **not** open links that look like actions (`unsubscribe`, `optout`, `confirm`, `verify`,
  `approve`, one time tokens): reputation and domain checks only, since a GET can act.

### Step 3: `linkscanner` service
* Node and Playwright with Chromium, pinned version, own container, `POST /scan` returning the
  result; **no database access**; shared token auth.
* Fresh browser context per URL. Record redirects (HTTP, JavaScript, meta refresh; max 10, loop
  detection), final URL, title, visible text (20 KB), forms and field types, downloads (blocked,
  never saved; name and type recorded), screenshot (JPEG, capped), TLS issuer, HTTP status.
* Limits: 20 s per URL, GET only, no form submissions or clicks, no camera, microphone, location,
  notifications; WebRTC, service workers and WebSockets disabled; memory limits; restart the
  browser every N jobs; small concurrency.
* **SSRF protection (critical):** all browser traffic goes through an outbound proxy that resolves
  each host **once**, checks the IP and connects to exactly that IP. Block private and loopback
  ranges (10/8, 172.16/12, 192.168/16, 127/8, ::1, fc00::/7), link local (including
  169.254.169.254), the server's own public IP, the database host, and Docker network names.
* **Isolation:** its own Docker network with internet egress only; the backend joins that network
  only to call it; not on `aegisbox-network`. Non root, read only filesystem except a temp folder,
  all capabilities dropped, CPU and memory limits, Chromium sandbox kept on.
* Shorteners (bit.ly and similar): expand through their preview endpoints where possible instead
  of visiting, so scans do not show up in the link owner's click counts.

### Step 4: scoring
* Signals: reputation list hits (malicious), brand impersonation of Nigerian banks and fintechs
  (GTBank, Access, Zenith, First Bank, UBA, Opay, PalmPay, Moniepoint, Kuda) and telcos (MTN,
  Airtel) on domains they do not own, pages asking for card, PIN, OTP, BVN, NIN or password, APK or
  app downloads, new domains (RDAP; "unknown" when `.ng` RDAP is unavailable, not scored), IP or
  punycode links, unusual TLDs, no HTTPS, many redirects, final domain differing from the visible
  one, unreachable or blocked, possible geo cloaking.
* Bands: malicious 70+, suspicious 30 to 69, safe below 30; store every reason.
* Cache by full URL and by final domain: 24 h for safe, 7 days for malicious.
* Admin overrides: false positive, per customer or global allow and block lists.

### Step 5: data and portal
* Migration: `link_scan_jobs`, `link_scans` (URL, normalized URL, final URL and domain, verdict,
  score, reasons JSON, title, text excerpt, forms JSON, download, status, screenshot **path**,
  TLS issuer, scanned and expiry times), `message_link_scans`, `link_domain_lists`.
* Screenshots as files on a private volume, path in the database (the database is managed and
  remote; images would bloat it and its backups). Daily retention job.
* Portal **Network > Link Scans** (admins only, views logged): filters for customer, sender,
  verdict, Lagos dates; message list with worst verdict; detail with redirect chain, reasons,
  screenshot; false positive and list buttons; summary counts and top bad domains.
* Personal data: message text, numbers and screenshots fall under Nigeria's Data Protection Act
  2023; keep access restricted and retention short; mention scanning in customer terms.

### Step 6: tests and deployment
* Unit: extraction, SSRF guard (private IPs, DNS change mid redirect, metadata address, database
  IP), scoring.
* Fixtures on a local test site: safe page, 3 step redirect, JavaScript redirect, fake bank login,
  APK download, links to 127.0.0.1 and the database host (must be blocked), a page that never loads.
* End to end through `mnosim` with the test customer; check a real customer cannot reach the
  simulator; load test 1,000 messages with 200 distinct links.
* Compose: `linkscanner` on an isolated network; env `LINKSCAN_URL`, `LINKSCAN_TOKEN`,
  `LINKSCAN_ENABLED`, `LINKSCAN_MAX_CONCURRENCY`, optional reputation keys; health check and
  admin warnings when the queue grows or the scanner is down; `LINKSCAN_ENABLED` off stops scanning
  without a redeploy. Check server memory (Chromium needs about 1 to 2 GB) and pin and update the
  Playwright image monthly.

## Phase 2 (needs owner approval and a legal check)
Per customer scan mode: off, monitor, hold (wait for a verdict up to a set time), block (REJECTD
DLR with a clear code; wallet policy decides charging). Fail open or fail closed when the scanner
is down, per customer. Check contracts and NCC rules before blocking real traffic.
