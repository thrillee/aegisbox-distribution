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
