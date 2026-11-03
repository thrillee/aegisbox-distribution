'use strict';
// Opens the fixture pages in real Chromium through the guard.
const test = require('node:test');
const assert = require('node:assert');
const { chromium } = require('playwright');
const { scanUrl } = require('../src/scan');
const { startFixtures } = require('./fixtures');

let browser, fx, cfg, base;

test.before(async () => {
  fx = await startFixtures();
  browser = await chromium.launch({ headless: true, chromiumSandbox: process.env.LINKSCAN_CHROMIUM_SANDBOX === 'true', args: ['--disable-quic'] });
  base = `http://fixture.test:${fx.port}`;
  cfg = {
    timeoutMs: 8000, settleMs: 1500, maxBytes: 5 * 1024 * 1024, maxScreenshotBytes: 1024 * 1024,
    blockIPs: [], blockHosts: [],
    testGuardOptions: {
      testAllowIPs: ['127.0.0.1'],
      lookup: async (h) => ({ 'fixture.test': [{ address: '127.0.0.1', family: 4 }], 'internal.test': [{ address: '10.0.0.5', family: 4 }] })[h] || Promise.reject(Object.assign(new Error('ENOTFOUND'), { code: 'ENOTFOUND' })),
    },
  };
});
test.after(async () => { await browser.close(); fx.close(); });

const scan = (path, extra = {}) => scanUrl(browser, { ...cfg, ...extra }, { url: path.startsWith('http') ? path : base + path });

test('safe page', async () => {
  const r = await scan('/safe');
  assert.strictEqual(r.error, null);
  assert.strictEqual(r.status, 200);
  assert.strictEqual(r.title, 'Shoe Shop');
  assert.match(r.text, /Shoes on sale/);
  assert.ok(r.screenshot && r.screenshot.length > 1000);
  assert.deepStrictEqual(r.blocked, []);
});

test('three step HTTP redirect is recorded', async () => {
  const r = await scan('/r1');
  assert.strictEqual(r.finalUrl, base + '/safe');
  assert.deepStrictEqual(r.redirects.map((x) => x.url.replace(base, '')), ['/r1', '/r2', '/r3', '/safe']);
  assert.deepStrictEqual(r.redirects.map((x) => x.status), [302, 301, 302, 200]);
});

test('JavaScript and meta refresh redirects are followed', async () => {
  for (const p of ['/js', '/meta']) {
    const r = await scan(p);
    assert.strictEqual(r.finalUrl, base + '/safe', p);
    assert.ok(r.redirects.some((x) => x.type === 'script_or_meta'), p);
  }
});

test('fake bank login: forms recorded, nothing submitted', async () => {
  const before = fx.hits.posts;
  const r = await scan('/bank');
  assert.match(r.title, /GTBank/);
  const names = r.forms.flatMap((f) => f.fields.map((x) => x.name));
  for (const n of ['password', 'card_number', 'pin', 'bvn']) assert.ok(names.includes(n), n);
  assert.strictEqual(r.forms[0].fields.find((x) => x.name === 'account').label, 'Account number');
  assert.strictEqual(fx.hits.posts, before, 'the auto submitting form must not reach the server');
  assert.ok(r.blockedWrites >= 1);
  assert.strictEqual(r.finalUrl, base + '/bank');
});

test('APK download is recorded and not saved', async () => {
  for (const p of ['/apk', '/apkpage']) {
    const r = await scan(p);
    assert.ok(r.downloads.some((d) => d.filename === 'update.apk'), `${p}: ${JSON.stringify(r.downloads)}`);
  }
});

test('redirect loop stops', async () => {
  const r = await scan('/loop');
  assert.ok(r.error, 'loop must be reported');
  assert.ok(r.redirects.length <= 30);
});

test('redirect to the cloud metadata address is blocked', async () => {
  const r = await scan('/internal');
  assert.ok(r.blocked.some((b) => b.ip === '169.254.169.254'), JSON.stringify(r.blocked));
  assert.strictEqual(r.errorKind, 'blocked');
  assert.ok(!r.text || !/meta-data/.test(r.text));
});

test('redirect to a name resolving to a private address is blocked', async () => {
  const r = await scan('/internal-name');
  assert.ok(r.blocked.some((b) => b.ip === '10.0.0.5'), JSON.stringify(r.blocked));
});

test('direct links to loopback and private addresses are blocked', async () => {
  for (const u of ['http://127.0.0.2:8080/', 'http://10.0.0.1/', 'http://[::1]:8080/', 'http://localhost:8080/']) {
    const r = await scanUrl(browser, cfg, { url: u });
    assert.strictEqual(r.errorKind, 'blocked', `${u}: ${r.error} ${JSON.stringify(r.blocked)}`);
  }
});

test('database host from the request is blocked', async () => {
  const c = { ...cfg, testGuardOptions: { ...cfg.testGuardOptions, testAllowIPs: [] , lookup: async (h) => [{ address: '93.184.216.34', family: 4 }] } };
  const r = await scanUrl(browser, c, { url: 'http://db.example.com:5432/', blockHosts: ['db.example.com'] });
  assert.strictEqual(r.errorKind, 'blocked');
});

test('hidden requests: private subresource blocked, POST and WebSocket not sent', async () => {
  const before = fx.hits.posts;
  const r = await scan('/sub');
  assert.strictEqual(r.title, 'Sub');
  assert.ok(r.blocked.some((b) => b.ip === '10.0.0.1'), JSON.stringify(r.blocked));
  await new Promise((res) => setTimeout(res, 300));
  assert.strictEqual(fx.hits.posts, before, 'POST from script must not be sent');
  assert.ok(!fx.hits.paths.some((p) => p.includes('/ws')), 'WebSocket must not connect');
});

test('geo blocked page is reported with its status', async () => {
  const r = await scan('/geo');
  assert.strictEqual(r.status, 403);
  assert.match(r.text, /not available in your country/);
});

test('page that never loads times out', async () => {
  const r = await scan('/hang', { timeoutMs: 3000, settleMs: 500 });
  assert.strictEqual(r.errorKind, 'timeout');
  assert.ok(r.durationMs < 6000);
});

test('unknown host is a DNS error', async () => {
  const r = await scan('http://does-not-exist.test/');
  assert.strictEqual(r.errorKind, 'dns');
});
