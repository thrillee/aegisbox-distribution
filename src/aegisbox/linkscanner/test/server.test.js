'use strict';
const test = require('node:test');
const assert = require('node:assert');
const { loadConfig, tokenOK } = require('../src/server');

const TOKEN = '0123456789abcdef0123';

test('stays idle without a token', () => {
  const cfg = loadConfig({});
  assert.strictEqual(cfg.enabled, false);
  assert.match(cfg.disabledReason, /LINKSCAN_TOKEN/);
});

test('rejects a short token', () => {
  assert.throws(() => loadConfig({ LINKSCAN_TOKEN: 'short' }), /16 characters/);
});

test("stays idle until the server's public IP is blocked", () => {
  const cfg = loadConfig({ LINKSCAN_TOKEN: TOKEN });
  assert.strictEqual(cfg.enabled, false);
  assert.match(cfg.disabledReason, /LINKSCAN_BLOCK_IPS/);
  assert.strictEqual(loadConfig({ LINKSCAN_TOKEN: TOKEN, LINKSCAN_BLOCK_IPS: '198.51.100.7' }).enabled, true);
  assert.strictEqual(loadConfig({ LINKSCAN_TOKEN: TOKEN, LINKSCAN_ALLOW_NO_BLOCK_IPS: 'true' }).enabled, true);
});

test('fails at start on a malformed blocked address', () => {
  assert.throws(() => loadConfig({ LINKSCAN_TOKEN: TOKEN, LINKSCAN_BLOCK_IPS: 'not-an-ip' }), /invalid blocked address/);
});

test('token check', () => {
  assert.strictEqual(tokenOK(`Bearer ${TOKEN}`, TOKEN), true);
  assert.strictEqual(tokenOK('Bearer wrong', TOKEN), false);
  assert.strictEqual(tokenOK(undefined, TOKEN), false);
});
