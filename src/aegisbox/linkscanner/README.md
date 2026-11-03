# linkscanner

Opens links from simulator traffic in a locked-down Chromium and reports what each page does. The
aegisbox backend calls it, scores the result and stores it; this service has no database access.

* `POST /scan` with `Authorization: Bearer $LINKSCAN_TOKEN` and `{"url": "...", "blockHosts": ["db.host"]}`
  returns the redirect chain, final URL, HTTP status, title, visible text (20 KB), forms and field
  types (never values), downloads (cancelled, never saved), requests the SSRF guard refused, TLS
  issuer and a JPEG screenshot (base64). 429 when busy, 500 when the browser failed (the backend
  retries both).
* `GET /healthz`.

## Safety

* **SSRF guard** (`src/guard.js`): every browser connection goes through a proxy started for that
  scan. It resolves each host once, refuses the host if any address is private, loopback, link
  local (including 169.254.169.254), reserved, the server's own IP (`LINKSCAN_BLOCK_IPS`) or the
  database (`LINKSCAN_BLOCK_HOSTS`, plus the host aegisbox sends), and connects to exactly the
  checked address. Ports below 1024 other than 80 and 443 are refused.
* GET only: form submissions and script POSTs are blocked. No WebSockets, WebRTC, service workers,
  QUIC, popups, dialogs, camera, microphone, location or notifications. Fresh context per URL,
  20 s per URL, 25 MB per scan, browser restarted every 50 scans.
* Honest: mobile Chrome user agent plus `AegisboxLinkCheck/1.0`, one fixed IP, no click simulation.
  Shortener links are expanded with a HEAD request (bot user agent) so the short link itself is
  not loaded in the browser.
* Container: non root (`pwuser`), read only root filesystem with a `/tmp` tmpfs, all capabilities
  dropped, `no-new-privileges`, memory and CPU limits, Chromium sandbox **on** via
  `seccomp-chromium.json` (Docker's default profile plus `clone`, `unshare`, `setns`, `chroot`).
  Without `LINKSCAN_TOKEN` or `LINKSCAN_BLOCK_IPS` it stays idle and starts no browser
  (`LINKSCAN_ALLOW_NO_BLOCK_IPS=true` lifts the second for local runs only).

## Settings

| Variable | Default | |
|---|---|---|
| `LINKSCAN_TOKEN` | (none: idle) | shared with aegisbox, at least 16 characters |
| `LINKSCAN_MAX_CONCURRENCY` | 2 | pages open at once |
| `LINKSCAN_BLOCK_IPS` | (none: idle) | **required**: the server's public IP; comma separated IPs or CIDRs |
| `LINKSCAN_BLOCK_HOSTS` | | comma separated host names to block |
| `LINKSCAN_TIMEOUT_MS` | 20000 | per URL |
| `LINKSCAN_CHROMIUM_SANDBOX` | true | only for local tests as root |

## Tests

`npm test` runs the guard tests and opens a local fixture site (safe page, redirect chains,
JavaScript and meta redirects, fake bank login that auto submits, APK download, redirect loop,
redirects to the metadata address and to private names, hidden POST and WebSocket, geo blocked
page, a page that never loads). Run as root on a workstation with `LINKSCAN_CHROMIUM_SANDBOX`
unset; inside the image the same tests pass with the sandbox on:

```bash
docker build -t linkscanner .
docker run --rm --read-only --tmpfs /tmp --cap-drop ALL --security-opt no-new-privileges \
  --security-opt seccomp=$PWD/seccomp-chromium.json -e LINKSCAN_CHROMIUM_SANDBOX=true \
  -v $PWD/test:/app/test:ro linkscanner node --test --test-concurrency=1 test/guard.test.js test/scan.test.js
```

Keep the Playwright image tag in the `Dockerfile` and the `playwright` version in `package.json`
the same, and update both monthly.
