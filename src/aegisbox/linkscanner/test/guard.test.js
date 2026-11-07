'use strict';
const test = require('node:test');
const assert = require('node:assert');
const net = require('node:net');
const http = require('node:http');
const { Guard, GuardError, startProxy, portAllowed } = require('../src/guard');

test('blocks private, loopback, metadata and reserved addresses', () => {
  const g = new Guard();
  for (const ip of [
    '10.1.2.3', '172.17.0.2', '192.168.1.1', '127.0.0.1', '127.8.8.8', '169.254.169.254', '0.0.0.0',
    '100.64.0.1', '224.0.0.1', '255.255.255.255', '::1', '::', 'fd00:ec2::254', 'fe80::1',
    '::ffff:10.0.0.1', '::ffff:7f00:1', '64:ff9b::a00:1', '::ffff:169.254.169.254',
  ]) {
    assert.ok(g.ipBlockedReason(ip), `${ip} should be blocked`);
  }
  for (const ip of ['8.8.8.8', '41.58.1.1', '2a00:1450:4001:80b::200e']) {
    assert.strictEqual(g.ipBlockedReason(ip), null, `${ip} should be allowed`);
  }
});

test('blocks configured server addresses and database host addresses', async () => {
  const g = new Guard({
    blockIPs: ['8.8.4.4', '41.0.0.0/8'],
    blockHosts: ['db.example.com'],
    lookup: async (h) => (h === 'db.example.com' ? [{ address: '35.1.2.3', family: 4 }] : []),
  });
  await g.resolveBlockedHosts();
  assert.ok(g.ipBlockedReason('8.8.4.4'));
  assert.ok(g.ipBlockedReason('41.58.1.1'));
  assert.ok(g.ipBlockedReason('35.1.2.3'));
  assert.strictEqual(g.ipBlockedReason('35.1.2.4'), null);
});

test('refuses a name when any of its addresses is blocked (DNS pointing inside)', async () => {
  const g = new Guard({
    lookup: async (h) =>
      ({
        'mixed.example': [{ address: '93.184.216.34', family: 4 }, { address: '10.0.0.1', family: 4 }],
        'ok.example': [{ address: '93.184.216.34', family: 4 }],
      })[h] || [],
  });
  await assert.rejects(g.resolve('mixed.example', 443), GuardError);
  assert.strictEqual(await g.resolve('ok.example', 443), '93.184.216.34');
  await assert.rejects(g.resolve('localhost', 80), GuardError);
  await assert.rejects(g.resolve('[::1]', 443), GuardError);
  await assert.rejects(g.resolve('ok.example', 22), /port 22/);
  await assert.rejects(g.resolve('nothing.example', 443), (e) => e.kind === 'dns');
});

test('ports', () => {
  assert.ok(portAllowed(80) && portAllowed(443) && portAllowed(8080));
  assert.ok(!portAllowed(22) && !portAllowed(25) && !portAllowed(0) && !portAllowed(70000));
});

test('proxy refuses CONNECT to blocked addresses and records them', async () => {
  const g = new Guard({ lookup: async () => [{ address: '10.0.0.9', family: 4 }] });
  const proxy = await startProxy(g);
  const status = await new Promise((resolve) => {
    const s = net.connect(proxy.port, '127.0.0.1', () => s.write('CONNECT internal.example:443 HTTP/1.1\r\nHost: internal.example:443\r\n\r\n'));
    s.once('data', (d) => { resolve(d.toString().split('\r\n')[0]); s.destroy(); });
  });
  assert.match(status, /403/);
  assert.strictEqual(proxy.events.blocked[0].ip, '10.0.0.9');
  await proxy.close();
});

test('proxy refuses plain HTTP POST and blocked GET', async () => {
  const g = new Guard();
  const proxy = await startProxy(g);
  const get = (method, url) =>
    new Promise((resolve) => {
      const r = http.request({ host: '127.0.0.1', port: proxy.port, method, path: url }, (res) => { res.resume(); resolve(res.statusCode); });
      r.on('error', () => resolve('dropped'));
      r.end();
    });
  assert.strictEqual(await get('POST', 'http://example.com/'), 405);
  assert.strictEqual(await get('GET', 'http://169.254.169.254/latest/meta-data/'), 'dropped');
  assert.strictEqual(proxy.events.blocked[0].ip, '169.254.169.254');
  await proxy.close();
});

// A minimal stand-in for a pool exit (like geo.iproyal.com): answers plain proxy GETs with the
// request target and the path from an absolute URI, and tunnels CONNECT like a real proxy would.
function fakeExit() {
  const seen = [];
  const server = http.createServer((req, res) => {
    seen.push(req.headers['proxy-authorization'] ? 'auth' : 'noauth');
    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end(`exit was asked for ${req.url}`);
  });
  server.on('connect', (req, client, head) => {
    const m = /^([^:]+):(\d+)$/.exec(req.url || '');
    if (!m) { client.end('HTTP/1.1 400 Bad Request\r\n\r\n'); return; }
    seen.push(`tunnel ${m[1]}:${m[2]}`);
    const up = net.connect(Number(m[2]), m[1], () => {
      client.write('HTTP/1.1 200 OK\r\n\r\n');
      if (head && head.length) up.write(head);
      up.pipe(client);
      client.pipe(up);
    });
    up.on('error', () => client.destroy());
  });
  return new Promise((resolve) => server.listen(0, '127.0.0.1', () => resolve({ server, port: server.address().port, seen })));
}

test('plain HTTP and CONNECT are chained through an upstream, still guard-vetted', async () => {
  const exit = await fakeExit();
  // DNS resolves the test target to 127.0.0.1 (allowed for fixtures) and internal names to a
  // private address: the guard must decide locally, so internal.example must never reach the
  // fake exit even though the exit could serve it.
  const g = new Guard({
    testAllowIPs: ['127.0.0.1'],
    lookup: async (h) => [{ address: h.startsWith('internal') ? '10.9.9.9' : '127.0.0.1', family: 4 }],
  });
  const proxy = await startProxy(g, { upstream: { host: '127.0.0.1', port: exit.port, username: 'u', password: 'p' } });

  const plain = await new Promise((resolve, reject) => {
    // Absolute-URI GET through the guard, with Host pointing at the origin.
    http.get({ host: '127.0.0.1', port: proxy.port, path: 'http://external.example/x', headers: { host: 'external.example' } }, (r) => {
      let body = '';
      r.on('data', (c) => { body += c; });
      r.on('end', () => resolve(body));
    }).on('error', reject);
  });
  assert.match(plain, /external\.example\/x/);
  assert.strictEqual(exit.seen[0], 'auth', 'upstream credentials are sent');

  // CONNECT through the exit, to a public host:after the handshake the origin page must answer.
  const tlsTarget = http.createServer((_q, r) => { r.writeHead(200); r.end('origin page'); });
  await new Promise((resolve) => tlsTarget.listen(0, '127.0.0.1', resolve));
  const hello = await new Promise((resolve, reject) => {
    const out = net.connect(proxy.port, '127.0.0.1');
    const body = [];
    out.write(`CONNECT external.example:${tlsTarget.address().port} HTTP/1.1\r\nHost: external.example\r\n\r\n`);
    out.on('data', function onData(d) {
      body.push(d);
      if (Buffer.concat(body).toString().includes('200 Connection Established')) {
        out.removeListener('data', onData);
        out.write(`GET / HTTP/1.1\r\nHost: external.example\r\nConnection: close\r\n\r\n`);
      } else if (body.length > 10) reject(new Error('no 200'));
    });
    out.on('data', (d) => {
      if (d.toString().includes('origin page')) { resolve('saw origin page'); out.destroy(); tlsTarget.close(); }
    });
    out.on('error', reject);
  });
  assert.strictEqual(hello, 'saw origin page');

  // The blocked name refuses locally and never appears at the exit.
  await new Promise((resolve) => {
    const s = net.connect(proxy.port, '127.0.0.1', () => s.write('CONNECT internal.example:443 HTTP/1.1\r\nHost: i:443\r\n\r\n'));
    s.once('data', (d) => { assert.match(d.toString(), /403/); s.destroy(); resolve(); });
  });
  assert.ok(!exit.seen.some((x) => String(x).includes('internal')), 'blocked host must not reach the exit');

  await proxy.close();
  exit.server.close();
  tlsTarget.close();
});
