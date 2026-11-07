// Upstream proxy pool: exit IPs for the scanned page so requests do not originate from the
// server's own address. Entries look like
//   http://user:pass_with-{session}@geo.example.com:12321
// The literal "{session}" in the username or password is replaced with a unique id per scan, so
// the provider holds one exit IP for the whole scan like a real phone with one connection.
// Entries without "{session}" stay as they are (for providers that rotate on their own terms).
'use strict';

const crypto = require('node:crypto');
const net = require('node:net');

/** Parses pool entries from the scanner configuration into upstream descriptors. */
function parsePool(entries) {
  const pool = [];
  for (const raw of entries || []) {
    const s = String(raw).trim();
    if (!s) continue;
    let u;
    try {
      u = new URL(s);
    } catch {
      throw new Error(`invalid proxy pool entry: ${s}`);
    }
    if (u.protocol !== 'http:') throw new Error(`proxy pool supports only http upstreams: ${s}`);
    if (!u.hostname) throw new Error(`proxy pool entry without a host: ${s}`);
    pool.push({
      template: s,
      host: u.hostname,
      port: Number(u.port || 80),
      username: u.username || '',
      password: u.password || '',
    });
  }
  return pool;
}

/**
 * Picks a reachable upstream for one scan. Entries are tried in random order and the first one
 * whose port accepts a connection wins. Returns an expanded descriptor (session id applied) or
 * null when nothing is reachable, so the scan falls back to the server's own address.
 * @param {object[]} pool as returned by parsePool
 */
async function pick(pool) {
  const order = (pool || []).map((p, i) => ({ p, r: Math.random() + i * 1e-12 })).sort((a, b) => a.r - b.r)
    .map((x) => x.p);
  for (const entry of order) {
    const id = crypto.randomBytes(6).toString('hex');
    const up = {
      host: entry.host,
      port: entry.port,
      username: entry.username.replace('{session}', id),
      password: entry.password.replace('{session}', id),
    };
    if (await reachable(up)) return up;
  }
  return null;
}

/** Confirms the upstream is reachable before a scan is bound to it. */
function reachable(up) {
  return new Promise((resolve) => {
    const s = net.connect({ host: up.host, port: up.port, timeout: 5000 });
    s.once('connect', () => { s.destroy(); resolve(true); });
    s.once('timeout', () => { s.destroy(); resolve(false); });
    s.once('error', () => resolve(false));
  });
}

function authorization(up) {
  if (!up.username && !up.password) return null;
  return `Basic ${Buffer.from(`${up.username}:${up.password}`).toString('base64')}`;
}

module.exports = { parsePool, pick, authorization };
