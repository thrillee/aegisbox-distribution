// Opens one link in a fresh, locked-down browser context and reports what the page does.
'use strict';

const { Guard, startProxy } = require('./guard');
const { pick: pickProxy } = require('./proxies');
const { randomProfile } = require('./devices');

const MAX_NAVIGATIONS = 10;
const TEXT_LIMIT = 20 * 1024;

// Link shorteners: the browser follows the shortener's redirect exactly like a person
// tapping the link; the scanner flags it so the scorer ignores the domain change.
const SHORTENERS = new Set([
  'bit.ly', 'bitly.com', 'tinyurl.com', 't.co', 'goo.gl', 'ow.ly', 'is.gd', 'buff.ly', 'cutt.ly',
  'rb.gy', 'shorturl.at', 't.ly', 'rebrand.ly', 'tiny.cc', 's.id', 'bit.do', 'lnkd.in', 'v.gd',
]);

// Fingerprint patches that run before any page script. The stock headless build reveals
// navigator.webdriver and lacks navigator.userAgentData and window.chrome, which real
// Chrome on Android always has; these are rebuilt to match the profile's user agent.
function stealthScript() {
  return () => {
    const ua = navigator.userAgent;
    const full = (/Chrome\/([0-9.]+)/.exec(ua) || [])[1] || '';
    const major = full.split('.')[0] || '';
    const mobile = /\bMobile\b/.test(ua) || /Android/.test(ua);
    const and = (/Android ([0-9.]+)/.exec(ua) || [])[1] || '';
    const model = (() => {
      if (!/Android/.test(ua)) return '';
      const i = ua.indexOf(' Android ');
      const rest = ua.slice(i + 9);
      const semi = rest.indexOf('; ');
      const seg = semi < 0 ? '' : rest.slice(semi + 2);
      const end = seg.indexOf(') ');
      return end < 0 ? '' : seg.slice(0, end).trim();
    })();

    try {
      Object.defineProperty(Object.getPrototypeOf(navigator), 'webdriver', { get: () => false, configurable: true });
    } catch {}
    try {
      delete navigator.webdriver;
      Object.defineProperty(navigator, 'webdriver', { get: () => false, configurable: true });
    } catch {}

    // Client hints (real Chrome for Android always exposes these).
    if (!navigator.userAgentData && major) {
      const brand =
        /SamsungBrowser/.test(ua) ? { brand: 'Samsung Internet', version: major } :
        /EdgA\//.test(ua) ? { brand: 'Microsoft Edge', version: major } :
        { brand: 'Google Chrome', version: major };
      const brands = [{ brand: 'Chromium', version: major }, brand, { brand: 'Not A(Brand', version: '99' }];
      const data = { brands, mobile, platform: 'Android', toJSON: () => data };
      const high = {};
      Object.defineProperty(navigator, 'userAgentData', { get: () => data, configurable: true });
      data.getHighEntropyValues = (hints) =>
        Promise.resolve(Object.fromEntries(hints.map((h) => [h, high[h]])));
      high.fullVersionList = brands.map((b) => ({ ...b, version: full || b.version }));
      high.uaFullVersion = full;
      high.model = model;
      high.platformVersion = and ? (Number(and.split('.')[0]) || and) + '.0.0' : '';
      high.architecture = 'arm';
      high.bitness = '64';
      high.mobile = mobile;
      high.platform = 'Android';
    }

    // window.chrome (present in real Chrome, missing in headless).
    if (!window.chrome) {
      const empty = () => {};
      window.chrome = {
        app: { isInstalled: false, getIsInstalled: empty, installState: () => 'not_installed', runningState: () => 'cannot_run' },
        runtime: {},
        csi: () => ({}),
        loadTimes: () => ({}),
      };
    }
  };
}

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

/**
 * Full e2e click: the link is placed on a blank page and really activated like a person
 * does (touch events on phones, mouse otherwise), so the navigation happens through the
 * browser's normal link activation path with no referrer. Resolves with the response.
 */
async function openByClick(page, target, cfg) {
  await page.evaluate((href) => {
    const a = document.createElement('a');
    a.id = 'scan-target';
    a.href = href;
    a.textContent = 'Open link';
    (document.body || document.documentElement).appendChild(a);
  }, target);

  // A short pause before the click, like a person glancing at the page.
  await new Promise((r) => setTimeout(r, 150 + Math.floor(Math.random() * 500)));

  let navErr = null;
  const nav = page
    .waitForNavigation({ waitUntil: 'load', timeout: cfg.timeoutMs })
    .catch((e) => {
      navErr = e;
      return null;
    });
  // The activation starts the navigation; the tap or click call itself may be aborted by
  // the page change, which is fine. The navigation result arrives through `nav`.
  await page
    .tap('#scan-target', { timeout: Math.max(1000, Math.min(5000, cfg.timeoutMs)) })
    .catch(() => {});
  const response = await nav;
  if (navErr) throw navErr;
  if (!response) throw new Error('the click navigated without a response');
  return response;
}

/**
 * Opens the page the way a person does: new context with a random mobile device and
 * browser profile, a blank page, and a click on the link itself.
 * @param {import('playwright').Browser} browser
 * @param {object} cfg scanner settings
 * @param {{url: string, blockHosts?: string[]}} job
 */
async function scanUrl(browser, cfg, job) {
  const started = Date.now();
  const profile = randomProfile();
  const guard = new Guard({
    blockIPs: cfg.blockIPs,
    blockHosts: [...cfg.blockHosts, ...(job.blockHosts || [])],
    ...(cfg.testGuardOptions || {}), // tests only
  });
  await guard.resolveBlockedHosts();

  const result = {
    url: job.url,
    device: profile.model,
    browser: profile.browser,
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
    exitProxy: null, // exit proxy host:port, null when no upstream is configured in the pool
    durationMs: 0,
  };

  // One exit per scan: the pool holds provider entries whose {session} placeholder makes the
  // proxy hold a single IP for the whole page load, like one phone on one connection.
  const exit = await pickProxy(cfg.proxyPool);
  if (cfg.proxyPool && cfg.proxyPool.length && !exit) {
    // Nothing reachable: a silent fallback would let the scan run from the server's address.
    result.error = 'no exit proxy in the pool is reachable';
    result.errorKind = 'other'; // it is a scanner setting problem, not a property of the page
    return result;
  }
  result.exitProxy = exit ? `${exit.host}:${exit.port}` : null;

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

  // Shorteners are left as they are: the browser follows the redirect exactly like a
  // person tapping the link would, and shortenerExpanded is set below from the hops.

  const proxy = await startProxy(guard, { maxBytes: cfg.maxBytes, upstream: exit });
  let context;
  try {
    context = await browser.newContext({
      proxy: { server: `http://127.0.0.1:${proxy.port}` },
      userAgent: profile.userAgent,
      viewport: profile.viewport,
      deviceScaleFactor: profile.deviceScaleFactor,
      screen: profile.screen,
      isMobile: profile.isMobile,
      hasTouch: profile.hasTouch,
      locale: profile.locale,
      timezoneId: profile.timezoneId,
      permissions: [], // no camera, microphone, location, notifications
      serviceWorkers: 'block',
      acceptDownloads: false, // downloads are cancelled and never saved
      javaScriptEnabled: true,
    });
    // No WebRTC: it can open UDP connections that do not go through the proxy. The same
    // script patches headless browser fingerprints (navigator.webdriver, client hints,
    // window.chrome) to match a real Chrome on the profile's device.
    await context.addInitScript(stealthScript());
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
    // Client hint request headers: the browser would not send sec-ch-ua* through a proxy,
    // but a real Chrome always does. Re-add them, consistent with the profile.
    const hintHeaders = {
      'sec-ch-ua': profile.clientHints.brands
        .map((b) => `"${b.brand}";v="${b.version}"`)
        .join(', '),
      'sec-ch-ua-mobile': profile.clientHints.mobile ? '?1' : '?0',
      'sec-ch-ua-platform': '"Android"',
    };
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
      return route.continue({ headers: { ...req.headers(), ...hintHeaders } });
    });

    const page = await context.newPage();
    page.setDefaultTimeout(cfg.timeoutMs);

    // Client hints (sec-ch-ua* headers and navigator.userAgentData) would otherwise
    // report the actual headless browser version, not the profile's Chrome release.
    // Set this before the first navigation; the newContext option alone is not enough.
    const cdp = await context.newCDPSession(page);
    await cdp
      .send('Network.setUserAgentOverride', {
        userAgent: profile.userAgent,
        acceptLanguage: profile.locale,
        platform: 'Android',
        userAgentMetadata: profile.clientHints,
      })
      .catch(() => {});

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
      response = await openByClick(page, target, cfg);
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

    // A redirect through a shortener: the final domain differing from the link's is expected.
    for (const hop of result.redirects) {
      try {
        if (SHORTENERS.has(new URL(hop.url).hostname.replace(/^www\./, ''))) result.shortenerExpanded = true;
      } catch {
        // only URLs are pushed into redirects, so this cannot happen
      }
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

module.exports = { scanUrl, openByClick, stealthScript, SHORTENERS };
