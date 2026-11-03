// Opens one link in a fresh, locked-down browser context and reports what the page does.
'use strict';

const http = require('node:http');
const https = require('node:https');
const { Guard, GuardError, startProxy } = require('./guard');

const MOBILE_UA =
  'Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) ' +
  'Chrome/126.0.0.0 Mobile Safari/537.36 AegisboxLinkCheck/1.0';
const BOT_UA = 'AegisboxLinkCheck/1.0 (+link safety check; no clicks counted)';

const MAX_NAVIGATIONS = 10;
const TEXT_LIMIT = 20 * 1024;

// Link shorteners: the destination is read from the shortener's redirect with a plain HEAD request
// (bot user agent, redirects not followed) and the browser opens the destination directly, so the
// scan does not load the shortener's page in a browser.
const SHORTENERS = new Set([
  'bit.ly', 'bitly.com', 'tinyurl.com', 't.co', 'goo.gl', 'ow.ly', 'is.gd', 'buff.ly', 'cutt.ly',
  'rb.gy', 'shorturl.at', 't.ly', 'rebrand.ly', 'tiny.cc', 's.id', 'bit.do', 'lnkd.in', 'v.gd',
]);

// Runs inside the page: forms and input fields (types, names, labels; never values).
function collectForms() {
  const describe = (el) => {
    const label =
      (el.id && document.querySelector(`label[for="${CSS.escape(el.id)}"]`)?.innerText) ||
      el.closest('label')?.innerText ||
      el.getAttribute('aria-label') ||
      '';
    return {
      tag: el.tagName.toLowerCase(),
      type: (el.getAttribute('type') || '').toLowerCase(),
      name: el.getAttribute('name') || '',
      id: el.id || '',
      placeholder: el.getAttribute('placeholder') || '',
      autocomplete: el.getAttribute('autocomplete') || '',
      label: label.trim().slice(0, 100),
    };
  };
  const fieldSel = 'input:not([type=hidden]):not([type=submit]):not([type=button]), select, textarea';
  const forms = [...document.querySelectorAll('form')].slice(0, 20).map((f) => ({
    action: f.getAttribute('action') || '',
    method: (f.getAttribute('method') || 'get').toLowerCase(),
    fields: [...f.querySelectorAll(fieldSel)].slice(0, 50).map(describe),
  }));
  const loose = [...document.querySelectorAll(fieldSel)].filter((el) => !el.closest('form')).slice(0, 50);
  if (loose.length) forms.push({ action: '', method: '', fields: loose.map(describe) });
  return forms;
}

// HEAD request through the guard, without following redirects. Returns the Location, if any.
function headLocation(url, guard, timeoutMs = 8000) {
  return new Promise((resolve, reject) => {
    const u = new URL(url);
    const lib = u.protocol === 'https:' ? https : http;
    const port = Number(u.port || (u.protocol === 'https:' ? 443 : 80));
    const req = lib.request(
      {
        method: 'HEAD',
        hostname: u.hostname,
        port,
        path: u.pathname + u.search,
        headers: { 'User-Agent': BOT_UA },
        timeout: timeoutMs,
        lookup: (_h, opts, cb) => {
          guard
            .resolve(u.hostname, port)
            .then((ip) => (opts && opts.all ? cb(null, [{ address: ip, family: ip.includes(':') ? 6 : 4 }]) : cb(null, ip, ip.includes(':') ? 6 : 4)))
            .catch(cb);
        },
      },
      (res) => {
        res.resume();
        const loc = res.headers.location;
        resolve({ status: res.statusCode, location: loc ? new URL(loc, url).toString() : null });
      },
    );
    req.on('timeout', () => req.destroy(new Error('timeout')));
    req.on('error', reject);
    req.end();
  });
}

/**
 * @param {import('playwright').Browser} browser
 * @param {object} cfg scanner settings
 * @param {{url: string, blockHosts?: string[]}} job
 */
async function scanUrl(browser, cfg, job) {
  const started = Date.now();
  const guard = new Guard({
    blockIPs: cfg.blockIPs,
    blockHosts: [...cfg.blockHosts, ...(job.blockHosts || [])],
    ...(cfg.testGuardOptions || {}), // tests only
  });
  await guard.resolveBlockedHosts();

  const result = {
    url: job.url,
    finalUrl: null,
    status: null,
    title: null,
    text: null,
    redirects: [],
    forms: [],
    downloads: [],
    blocked: [],
    tlsIssuer: null,
    screenshot: null,
    error: null,
    errorKind: null, // dns, timeout, tls, blocked, http, redirect_loop, size_limit, other
    shortenerExpanded: false,
    blockedWrites: 0, // form submissions and other non GET requests the page tried
    durationMs: 0,
  };

  let target = job.url;
  let u;
  try {
    u = new URL(target);
  } catch {
    result.error = 'invalid URL';
    result.errorKind = 'other';
    return result;
  }
  if (!['http:', 'https:'].includes(u.protocol)) {
    result.error = 'only http and https links are opened';
    result.errorKind = 'other';
    return result;
  }

  // Shortener: read the destination without loading the shortener in the browser.
  if (SHORTENERS.has(u.hostname.replace(/^www\./, ''))) {
    try {
      const { status, location } = await headLocation(target, guard);
      if (location && status >= 300 && status < 400) {
        result.redirects.push({ url: target, status, type: 'shortener' });
        target = location;
        result.shortenerExpanded = true;
      }
    } catch (e) {
      if (e instanceof GuardError && e.kind === 'blocked') {
        result.blocked.push({ host: e.host, ip: e.ip, reason: e.reason });
      }
      // Otherwise fall back to opening the short link in the browser.
    }
  }

  const proxy = await startProxy(guard, { maxBytes: cfg.maxBytes });
  let context;
  try {
    context = await browser.newContext({
      proxy: { server: `http://127.0.0.1:${proxy.port}` },
      userAgent: MOBILE_UA,
      viewport: { width: 390, height: 844 },
      deviceScaleFactor: 2,
      isMobile: true,
      hasTouch: true,
      locale: 'en-NG',
      timezoneId: 'Africa/Lagos',
      permissions: [], // no camera, microphone, location, notifications
      serviceWorkers: 'block',
      acceptDownloads: false, // downloads are cancelled and never saved
      javaScriptEnabled: true,
    });
    // No WebRTC: it can open UDP connections that do not go through the proxy.
    await context.addInitScript(() => {
      for (const k of ['RTCPeerConnection', 'webkitRTCPeerConnection', 'RTCDataChannel']) {
        try { Object.defineProperty(window, k, { value: undefined, configurable: false }); } catch {}
      }
    });
    // No WebSockets.
    if (typeof context.routeWebSocket === 'function') {
      await context.routeWebSocket(/.*/, (ws) => ws.close());
    }

    let navigations = 0;
    let loopHit = false;
    // GET only: no form submissions or other writes, even from scripts.
    await context.route('**/*', (route) => {
      const req = route.request();
      if (!['GET', 'HEAD'].includes(req.method())) {
        result.blockedWrites++;
        // A form submission answered with 204 leaves the page as it is (an aborted one would
        // replace it with an error page); other writes are simply cancelled.
        if (req.isNavigationRequest()) return route.fulfill({ status: 204, body: '' });
        return route.abort('blockedbyclient');
      }
      return route.continue();
    });

    const page = await context.newPage();
    page.setDefaultTimeout(cfg.timeoutMs);

    page.on('request', (req) => {
      if (!req.isNavigationRequest() || req.frame() !== page.mainFrame() || req.method() !== 'GET') return;
      navigations++;
      if (navigations > MAX_NAVIGATIONS + 1) loopHit = true;
      const from = req.redirectedFrom();
      if (result.redirects.length < 30) {
        result.redirects.push({ url: req.url(), status: null, type: from ? 'http' : navigations === 1 ? 'start' : 'script_or_meta' });
      }
    });
    page.on('response', (res) => {
      const req = res.request();
      if (!req.isNavigationRequest() || req.frame() !== page.mainFrame()) return;
      const hop = [...result.redirects].reverse().find((r) => r.url === req.url() && r.status === null);
      if (hop) hop.status = res.status();
      const cd = res.headers()['content-disposition'] || '';
      if (/attachment/i.test(cd) && result.downloads.length < 10) {
        result.downloads.push({ url: req.url(), filename: (cd.match(/filename\*?=(?:UTF-8'')?"?([^";]+)/i) || [])[1] || '', mime: res.headers()['content-type'] || '' });
      }
    });
    page.on('download', (d) => {
      if (result.downloads.length < 10) {
        result.downloads.push({ url: d.url(), filename: d.suggestedFilename(), mime: '' });
      }
      d.cancel().catch(() => {});
    });
    page.on('dialog', (d) => d.dismiss().catch(() => {}));
    page.on('popup', (p) => p.close().catch(() => {}));

    let response = null;
    try {
      response = await page.goto(target, { waitUntil: 'load', timeout: cfg.timeoutMs });
    } catch (e) {
      const msg = String(e.message || e);
      result.error = msg.split('\n')[0].slice(0, 300);
      const targetHost = (() => { try { return new URL(target).hostname; } catch { return ''; } })();
      const hostIs = (list) => list.some((b) => b.host.replace(/^\[|\]$/g, '') === targetHost);
      if (/ERR_CERT|ERR_SSL/.test(msg)) result.errorKind = 'tls';
      else if (hostIs(proxy.events.blocked)) result.errorKind = 'blocked';
      else if (/ERR_NAME_NOT_RESOLVED/.test(msg) || hostIs(proxy.events.dnsErrors)) result.errorKind = 'dns';
      else if (/Timeout|ERR_TIMED_OUT/.test(msg)) result.errorKind = 'timeout';
      else if (proxy.events.blocked.length) result.errorKind = 'blocked';
      else if (/ERR_ABORTED/.test(msg) && result.downloads.length) result.errorKind = null; // a download, not a failure
      else result.errorKind = 'other';
      if (result.errorKind === null) result.error = null;
    }

    // Let script and meta refresh redirects run, within the time limit.
    const left = cfg.timeoutMs - (Date.now() - started);
    if (left > 500) {
      await page.waitForLoadState('networkidle', { timeout: Math.min(cfg.settleMs, left - 300) }).catch(() => {});
    }

    if (loopHit) {
      result.error = result.error || `more than ${MAX_NAVIGATIONS} redirects`;
      result.errorKind = result.errorKind || 'redirect_loop';
    }

    const shown = page.url();
    result.finalUrl = /^https?:/.test(shown) ? shown : null; // not about:blank or a browser error page
    if (response) {
      result.status = response.status();
      const sec = await response.securityDetails().catch(() => null);
      if (sec) result.tlsIssuer = sec.issuer || null;
    }
    // Status and TLS of the page finally shown (after script redirects).
    const finalHop = [...result.redirects].reverse().find((r) => r.status !== null);
    if (finalHop) result.status = finalHop.status;

    if (result.finalUrl) {
      result.title = (await page.title().catch(() => '')).slice(0, 300) || null;
      const text = await page.evaluate(() => document.body?.innerText || '').catch(() => '');
      result.text = text.slice(0, TEXT_LIMIT) || null;
      result.forms = await page.evaluate(collectForms).catch(() => []);
      const shot = await page.screenshot({ type: 'jpeg', quality: 60, fullPage: false, timeout: 5000 }).catch(() => null);
      if (shot && shot.length <= cfg.maxScreenshotBytes) result.screenshot = shot.toString('base64');
    }
    // A page that rendered but kept loading (or whose own form submission interrupted the load)
    // was checked: do not report it as unreachable.
    if (result.errorKind === 'timeout' && (result.title || result.text)) {
      result.error = null;
      result.errorKind = null;
    }
  } finally {
    if (context) await context.close().catch(() => {});
    await proxy.close();
  }

  result.blocked.push(...proxy.events.blocked);
  if (proxy.events.limitHit) {
    result.error = result.error || 'page larger than the download limit';
    result.errorKind = result.errorKind || 'size_limit';
  }
  if (!result.error && result.blocked.length && !result.finalUrl) {
    result.error = 'blocked by the SSRF guard';
    result.errorKind = 'blocked';
  }
  result.durationMs = Date.now() - started;
  return result;
}

module.exports = { scanUrl, MOBILE_UA, SHORTENERS };
