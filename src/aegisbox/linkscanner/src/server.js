// linkscanner: opens links in a locked-down Chromium and reports what the pages do.
// POST /scan {"url": "...", "blockHosts": ["db.example.com"]} with "Authorization: Bearer <token>".
// GET /healthz. No database access: the aegisbox backend stores the results.
'use strict';

const http = require('node:http');
const crypto = require('node:crypto');
const { chromium } = require('playwright');
const { scanUrl } = require('./scan');
const { Guard } = require('./guard');

function list(v) {
  return (v || '').split(',').map((s) => s.trim()).filter(Boolean);
}
function int(v, d) {
  const n = Number.parseInt(v, 10);
  return Number.isFinite(n) && n > 0 ? n : d;
}

function loadConfig(env = process.env) {
  const cfg = {
    port: int(env.PORT, 8080),
    token: env.LINKSCAN_TOKEN || '',
    maxConcurrency: int(env.LINKSCAN_MAX_CONCURRENCY, 2),
    maxQueue: int(env.LINKSCAN_MAX_QUEUE, 20),
    timeoutMs: int(env.LINKSCAN_TIMEOUT_MS, 20000),
    settleMs: int(env.LINKSCAN_SETTLE_MS, 3000),
    restartEvery: int(env.LINKSCAN_RESTART_EVERY, 50),
    maxBytes: int(env.LINKSCAN_MAX_BYTES, 25 * 1024 * 1024),
    maxScreenshotBytes: int(env.LINKSCAN_MAX_SCREENSHOT_BYTES, 1024 * 1024),
    blockIPs: list(env.LINKSCAN_BLOCK_IPS),
    blockHosts: list(env.LINKSCAN_BLOCK_HOSTS),
    sandbox: env.LINKSCAN_CHROMIUM_SANDBOX !== 'false',
  };
  if (cfg.token && cfg.token.length < 16) throw new Error('LINKSCAN_TOKEN must be at least 16 characters');
  // Fail at start on a malformed address rather than on every scan.
  new Guard({ blockIPs: cfg.blockIPs });
  // The server's own public IP must be blocked, or a scanned page could reach the services the
  // server exposes. Private ranges are blocked anyway, but the public IP cannot be guessed safely,
  // so scanning stays off until it is set. LINKSCAN_ALLOW_NO_BLOCK_IPS=true is for local testing.
  cfg.allowNoBlockIPs = env.LINKSCAN_ALLOW_NO_BLOCK_IPS === 'true';
  if (cfg.token.length < 16) cfg.disabledReason = 'LINKSCAN_TOKEN is not set';
  else if (!cfg.blockIPs.length && !cfg.allowNoBlockIPs) cfg.disabledReason = "LINKSCAN_BLOCK_IPS is not set (the server's public IP)";
  cfg.enabled = !cfg.disabledReason;
  return cfg;
}

class BrowserPool {
  constructor(cfg, launcher = chromium) {
    this.cfg = cfg;
    this.launcher = launcher;
    this.browser = null;
    this.used = 0;
    this.active = 0;
    this.launching = null;
  }

  async get() {
    if (this.browser && this.browser.isConnected() && this.used < this.cfg.restartEvery) {
      this.used++;
      return this.browser;
    }
    if (!this.launching) {
      this.launching = (async () => {
        const old = this.browser;
        this.browser = null;
        // Restart only when no scan uses the old browser.
        if (old) {
          while (this.active > 0) await new Promise((r) => setTimeout(r, 200));
          await old.close().catch(() => {});
        }
        this.browser = await this.launcher.launch({
          headless: true,
          chromiumSandbox: this.cfg.sandbox,
          args: [
            // Do not expose navigator.webdriver.
            '--disable-blink-features=AutomationControlled',
            '--disable-quic', // QUIC (UDP) would not go through the proxy
            '--force-webrtc-ip-handling-policy=disable_non_proxied_udp',
            '--disable-background-networking',
            '--disable-component-update',
            '--disable-domain-reliability',
            '--disable-sync',
            '--no-first-run',
            '--disable-features=MediaRouter,OptimizationHints,Translate,DialMediaRouteProvider',
          ],
        });
        this.used = 1;
      })().finally(() => {
        this.launching = null;
      });
      await this.launching;
      return this.browser;
    }
    await this.launching;
    return this.get();
  }

  async run(fn) {
    const browser = await this.get();
    this.active++;
    try {
      return await fn(browser);
    } finally {
      this.active--;
    }
  }

  async close() {
    if (this.browser) await this.browser.close().catch(() => {});
  }
}

function tokenOK(header, token) {
  const m = /^Bearer (.+)$/.exec(header || '');
  if (!m) return false;
  const a = crypto.createHash('sha256').update(m[1]).digest();
  const b = crypto.createHash('sha256').update(token).digest();
  return crypto.timingSafeEqual(a, b);
}

function readJSON(req, limit = 16 * 1024) {
  return new Promise((resolve, reject) => {
    let size = 0;
    const chunks = [];
    req.on('data', (c) => {
      size += c.length;
      if (size > limit) {
        reject(new Error('body too large'));
        req.destroy();
      } else chunks.push(c);
    });
    req.on('end', () => {
      try {
        resolve(JSON.parse(Buffer.concat(chunks).toString('utf8')));
      } catch (e) {
        reject(e);
      }
    });
    req.on('error', reject);
  });
}

function createServer(cfg, pool, scan = scanUrl) {
  let running = 0;
  const waiting = [];
  const stats = { scans: 0, failures: 0 };

  const acquire = () =>
    new Promise((resolve) => {
      if (running < cfg.maxConcurrency) {
        running++;
        resolve();
      } else waiting.push(resolve);
    });
  const release = () => {
    const next = waiting.shift();
    if (next) next();
    else running--;
  };

  const send = (res, code, body) => {
    const data = JSON.stringify(body);
    res.writeHead(code, { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(data) });
    res.end(data);
  };

  return http.createServer(async (req, res) => {
    if (req.method === 'GET' && req.url === '/healthz') {
      return send(res, 200, { ok: true, enabled: cfg.enabled !== false, running, waiting: waiting.length, ...stats });
    }
    if (req.method !== 'POST' || req.url !== '/scan') return send(res, 404, { error: 'not found' });
    if (cfg.enabled === false) return send(res, 503, { error: `scanner disabled: ${cfg.disabledReason || 'not configured'}` });
    if (!tokenOK(req.headers.authorization, cfg.token)) return send(res, 401, { error: 'unauthorized' });
    if (waiting.length >= cfg.maxQueue) return send(res, 429, { error: 'busy' });

    let body;
    try {
      body = await readJSON(req);
    } catch {
      return send(res, 400, { error: 'invalid JSON' });
    }
    if (typeof body.url !== 'string' || body.url.length > 4096) return send(res, 400, { error: 'url required' });
    const blockHosts = Array.isArray(body.blockHosts) ? body.blockHosts.filter((h) => typeof h === 'string').slice(0, 10) : [];

    await acquire();
    try {
      const result = await pool.run((browser) => scan(browser, cfg, { url: body.url, blockHosts }));
      stats.scans++;
      send(res, 200, result);
    } catch (e) {
      stats.failures++;
      // A scanner failure (browser crash), not a page problem: the backend retries the job.
      console.error(JSON.stringify({ level: 'error', msg: 'scan failed', error: String(e && e.message) }));
      send(res, 500, { error: 'scan failed' });
    } finally {
      release();
    }
  });
}

if (require.main === module) {
  const cfg = loadConfig();
  const pool = new BrowserPool(cfg);
  const server = createServer(cfg, pool);
  if (!cfg.enabled) {
    // Deployed but not configured: stay up and idle, without starting a browser.
    server.listen(cfg.port, () => {
      console.log(JSON.stringify({ level: 'warn', msg: `linkscanner idle: ${cfg.disabledReason}`, port: cfg.port }));
    });
    process.on('SIGTERM', () => server.close(() => process.exit(0)));
    return;
  }
  // Fail at start, not on the first scan, if Chromium cannot launch (for example the sandbox).
  pool
    .get()
    .then(() => {
      server.listen(cfg.port, () => {
        console.log(JSON.stringify({ level: 'info', msg: 'linkscanner listening', port: cfg.port, sandbox: cfg.sandbox, maxConcurrency: cfg.maxConcurrency, blockIPs: cfg.blockIPs.length, blockHosts: cfg.blockHosts.length }));
      });
    })
    .catch((e) => {
      console.error(JSON.stringify({ level: 'error', msg: 'Chromium failed to start', error: String(e.message).split('\n')[0] }));
      process.exit(1);
    });
  const stop = () => server.close(() => pool.close().then(() => process.exit(0)));
  process.on('SIGTERM', stop);
  process.on('SIGINT', stop);
}

module.exports = { loadConfig, createServer, BrowserPool, tokenOK };
