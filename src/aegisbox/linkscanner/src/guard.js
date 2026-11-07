// SSRF guard: every connection the browser makes goes through a local proxy that resolves the host
// once, checks every address against a block list and connects to exactly the checked address. A
// page can therefore never reach private networks, the cloud metadata address, the server itself,
// the database or other containers, even through redirects or DNS that changes between lookups.
'use strict';

const dns = require('node:dns').promises;
const http = require('node:http');
const net = require('node:net');
const { authorization } = require('./proxies');

// Ranges no scanned page may reach.
const BLOCKED_V4 = [
  ['0.0.0.0', 8], // "this network"
  ['10.0.0.0', 8], // private
  ['100.64.0.0', 10], // carrier grade NAT
  ['127.0.0.0', 8], // loopback
  ['169.254.0.0', 16], // link local, includes 169.254.169.254 (cloud metadata)
  ['172.16.0.0', 12], // private, includes Docker networks
  ['192.0.0.0', 24], // IETF protocol assignments
  ['192.0.2.0', 24], // documentation
  ['192.88.99.0', 24], // 6to4 relay
  ['192.168.0.0', 16], // private
  ['198.18.0.0', 15], // benchmarking
  ['198.51.100.0', 24], // documentation
  ['203.0.113.0', 24], // documentation
  ['224.0.0.0', 4], // multicast
  ['240.0.0.0', 4], // reserved and broadcast
];
const BLOCKED_V6 = [
  ['::', 128], // unspecified
  ['::1', 128], // loopback
  // IPv4 mapped (::ffff:0:0/96) and NAT64 (64:ff9b::/96) addresses are judged by the IPv4 address
  // they embed (see embeddedV4); listing the prefixes here would make BlockList match every IPv4.
  ['64:ff9b:1::', 48], // local NAT64
  ['100::', 64], // discard
  ['2001:db8::', 32], // documentation
  ['2002::', 16], // 6to4 (embeds IPv4)
  ['fc00::', 7], // unique local
  ['fe80::', 10], // link local
  ['ff00::', 8], // multicast
];

function baseBlockList() {
  const bl = new net.BlockList();
  for (const [a, p] of BLOCKED_V4) bl.addSubnet(a, p, 'ipv4');
  for (const [a, p] of BLOCKED_V6) bl.addSubnet(a, p, 'ipv6');
  return bl;
}
const BASE = baseBlockList();

// Embedded IPv4 in IPv4 mapped or NAT64 IPv6 addresses, so ::ffff:10.0.0.1 is judged as 10.0.0.1.
function embeddedV4(ip) {
  const lower = ip.toLowerCase();
  const m = lower.match(/^(?:::ffff:|64:ff9b::)(\d+\.\d+\.\d+\.\d+)$/);
  if (m) return m[1];
  const hex = lower.match(/^(?:::ffff:|64:ff9b::)([0-9a-f]{1,4}):([0-9a-f]{1,4})$/);
  if (hex) {
    const a = parseInt(hex[1], 16);
    const b = parseInt(hex[2], 16);
    return `${a >> 8}.${a & 255}.${b >> 8}.${b & 255}`;
  }
  return null;
}

// Ports a page may connect to: web ports and unprivileged ports. Blocks SSH, SMTP, databases on
// their usual ports and the like.
function portAllowed(port) {
  return port === 80 || port === 443 || (port >= 1024 && port <= 65535);
}

class Guard {
  /**
   * @param {object} opts
   * @param {string[]} [opts.blockIPs] extra addresses or CIDRs (the server's public IP)
   * @param {string[]} [opts.blockHosts] host names whose addresses are blocked (the database host)
   * @param {(host: string) => Promise<{address: string, family: number}[]>} [opts.lookup]
   */
  constructor(opts = {}) {
    this.extra = new net.BlockList();
    this.extraCount = 0;
    for (const entry of opts.blockIPs || []) this.addBlocked(entry);
    this.blockHosts = (opts.blockHosts || []).map((h) => h.toLowerCase());
    this.lookup = opts.lookup || ((host) => dns.lookup(host, { all: true, verbatim: true }));
    // Tests only (set in code, never from the environment): addresses allowed despite the block
    // list, so fixtures on 127.0.0.1 can be scanned.
    this.testAllowIPs = new Set(opts.testAllowIPs || []);
  }

  addBlocked(entry) {
    const [addr, bits] = String(entry).trim().split('/');
    if (!addr) return;
    const family = net.isIPv6(addr) ? 'ipv6' : net.isIPv4(addr) ? 'ipv4' : null;
    if (!family) throw new Error(`invalid blocked address: ${entry}`);
    if (bits !== undefined) this.extra.addSubnet(addr, Number(bits), family);
    else this.extra.addAddress(addr, family);
    this.extraCount++;
  }

  // Resolves the blocked host names and blocks their addresses. A Guard is made for each scan, so a
  // changed database address is picked up.
  async resolveBlockedHosts() {
    for (const host of this.blockHosts) {
      if (!host) continue;
      if (net.isIP(host)) {
        this.addBlocked(host);
        continue;
      }
      try {
        for (const { address } of await this.lookup(host)) this.addBlocked(address);
      } catch {
        // An unresolvable host cannot be reached by the browser either.
      }
    }
  }

  /** @returns {string|null} why the address is blocked, or null when allowed */
  ipBlockedReason(ip) {
    const family = net.isIPv6(ip) ? 'ipv6' : net.isIPv4(ip) ? 'ipv4' : null;
    if (!family) return 'not an IP address';
    if (this.testAllowIPs.has(ip)) return null;
    const v4 = family === 'ipv6' ? embeddedV4(ip) : null;
    if (v4) {
      const r = this.ipBlockedReason(v4);
      if (r) return r;
    }
    if (BASE.check(ip, family)) return 'private or reserved address';
    if (this.extra.check(ip, family)) return 'blocked address (server or database)';
    return null;
  }

  /**
   * Resolves host and returns the one address to connect to, or throws GuardError. Every address
   * the name resolves to must be allowed: a name that also points inside is refused outright.
   */
  async resolve(host, port) {
    if (!portAllowed(port)) throw new GuardError(host, null, `port ${port} not allowed`);
    const bare = host.replace(/^\[|\]$/g, '');
    if (net.isIP(bare)) {
      const reason = this.ipBlockedReason(bare);
      if (reason) throw new GuardError(host, bare, reason);
      return bare;
    }
    if (bare.toLowerCase() === 'localhost' || bare.toLowerCase().endsWith('.localhost')) {
      throw new GuardError(host, null, 'localhost');
    }
    let addrs;
    try {
      addrs = await this.lookup(bare);
    } catch (e) {
      throw new GuardError(host, null, `DNS lookup failed: ${e.code || e.message}`, 'dns');
    }
    if (!addrs.length) throw new GuardError(host, null, 'no address', 'dns');
    for (const { address } of addrs) {
      const reason = this.ipBlockedReason(address);
      if (reason) throw new GuardError(host, address, reason);
    }
    // Prefer IPv4: the container usually has no IPv6 route.
    const v4 = addrs.find((a) => a.family === 4);
    return (v4 || addrs[0]).address;
  }
}

class GuardError extends Error {
  constructor(host, ip, reason, kind = 'blocked') {
    super(`${host}${ip ? ` (${ip})` : ''}: ${reason}`);
    this.host = host;
    this.ip = ip;
    this.reason = reason;
    this.kind = kind; // blocked (SSRF guard) or dns (name does not resolve)
  }
}

/**
 * Starts a forward proxy on 127.0.0.1 for one scan. Handles CONNECT (HTTPS, WebSocket over TLS)
 * and plain HTTP requests; everything else is refused. Only GET and HEAD go out over plain HTTP.
 *
 * When an upstream (an external proxy from the pool, see proxies.js) is given, connections are
 * chained through it so the page is seen as coming from that exit IP instead of this server. The
 * guard still resolves and vetts every host itself first: a host that resolves only to blocked or
 * private addresses never reaches the upstream, and tunnels open with the exact verified address.
 * @param {object[]} opts
 * @param {number} [opts.maxBytes]
 * @param {{host: string, port: number, username?: string, password?: string}|null} [opts.upstream]
 */
async function startProxy(guard, { maxBytes = 25 * 1024 * 1024, upstream = null } = {}) {
  const events = { blocked: [], dnsErrors: [], bytes: 0, limitHit: false };
  const sockets = new Set();

  const record = (err) => {
    const item = { host: err.host, ip: err.ip, reason: err.reason };
    if (err.kind === 'dns') {
      if (events.dnsErrors.length < 50) events.dnsErrors.push(item);
    } else if (events.blocked.length < 50) {
      events.blocked.push(item);
    }
  };
  const count = (n, kill) => {
    events.bytes += n;
    if (events.bytes > maxBytes) {
      events.limitHit = true;
      kill();
    }
  };

  const server = http.createServer(async (req, res) => {
    // Plain HTTP through the proxy: the request line carries the absolute URL.
    let target;
    try {
      target = new URL(req.url);
    } catch {
      res.writeHead(400).end();
      return;
    }
    if (target.protocol !== 'http:' || !['GET', 'HEAD'].includes(req.method)) {
      res.writeHead(405).end();
      return;
    }
    const port = Number(target.port || 80);
    let ip;
    try {
      ip = await guard.resolve(target.hostname, port);
    } catch (e) {
      if (e instanceof GuardError) record(e);
      // Drop the connection rather than answer, so the browser reports a failed load, not a page.
      req.socket.destroy();
      return;
    }
    // Chained through an upstream proxy when one is set; the request keeps looking like it comes
    // from the browser (our own hint headers were not sent to the upstream this way).
    const headers = { ...req.headers };
    delete headers['proxy-connection'];
    delete headers['proxy-authorization'];
    headers.host = target.host;
    const originPort = port !== 80 ? `:${port}` : '';
    // A browser sends a proxy an absolute URI, not the path alone.
    const requestTarget = upstream
      ? `http://${target.hostname}${originPort}${target.pathname}${target.search}`
      : target.pathname + target.search;
    if (upstream) {
      const auth = authorization(upstream);
      if (auth) headers['proxy-authorization'] = auth;
    }
    const up = http.request(
      {
        host: upstream ? upstream.host : ip,
        port: upstream ? upstream.port : port,
        method: req.method,
        path: requestTarget,
        headers,
        setHost: false,
        timeout: 20000,
      },
      (upRes) => {
        res.writeHead(upRes.statusCode || 502, upRes.headers);
        upRes.on('data', (c) => count(c.length, () => { upRes.destroy(); res.destroy(); }));
        upRes.pipe(res);
      },
    );
    up.on('timeout', () => up.destroy(new Error('timeout')));
    up.on('error', () => { if (!res.headersSent) res.writeHead(502); res.end(); });
    up.end();
  });

  server.on('connect', async (req, client, head) => {
    sockets.add(client);
    client.on('close', () => sockets.delete(client));
    client.on('error', () => {});
    const m = /^(\[[^\]]+\]|[^:]+):(\d+)$/.exec(req.url || '');
    if (!m) {
      client.end('HTTP/1.1 400 Bad Request\r\n\r\n');
      return;
    }
    const host = m[1];
    const port = Number(m[2]);
    let ip;
    try {
      ip = await guard.resolve(host, port);
    } catch (e) {
      if (e instanceof GuardError) record(e);
      client.end(`HTTP/1.1 ${e.kind === 'dns' ? '502 Bad Gateway' : '403 Forbidden'}\r\n\r\n`);
      return;
    }
    // A direct connection targets the verified address. When an upstream is set the tunnel is
    // opened through it instead, with the verified address as the CONNECT target, the same way
    // the browser's own request line is preserved. The guard never sees the tunnel payload
    // (TLS), but only ever hands it an address it checked itself.
    const out = net.connect({ host: upstream ? upstream.host : ip, port: upstream ? upstream.port : port, timeout: 20000 });
    sockets.add(out);
    out.on('close', () => sockets.delete(out));
    out.on('timeout', () => out.destroy());
    out.on('error', () => client.destroy());
    out.on('connect', () => {
      if (!upstream) {
        client.write('HTTP/1.1 200 Connection Established\r\n\r\n');
        if (head && head.length) out.write(head);
        out.on('data', (c) => count(c.length, () => { out.destroy(); client.destroy(); }));
        out.pipe(client);
        client.pipe(out);
        return;
      }
      const auth = authorization(upstream);
      out.write(
        `CONNECT ${ip}:${port} HTTP/1.1\r\n` +
        `Host: ${ip}:${port}\r\n` +
        (auth ? `Proxy-Authorization: ${auth}\r\n` : '') +
        '\r\n',
      );
      let reply = '';
      const handshake = (chunk) => {
        reply += chunk.toString('latin1');
        const end = reply.indexOf('\r\n\r\n');
        if (end < 0) {
          if (reply.length > 4096) { out.destroy(); client.destroy(); }
          return;
        }
        out.removeListener('data', handshake);
        if (!/^HTTP\/1\.[01] 200/.test(reply)) {
          client.end('HTTP/1.1 502 Bad Gateway\r\n\r\n');
          out.destroy();
          return;
        }
        client.write('HTTP/1.1 200 Connection Established\r\n\r\n');
        if (head && head.length) out.write(head);
        // Bytes past the end of the handshake reply belong to the tunnel.
        const rest = Buffer.from(reply.slice(end + 4), 'latin1');
        if (rest.length) client.write(rest);
        out.on('data', (c) => count(c.length, () => { out.destroy(); client.destroy(); }));
        out.pipe(client);
        client.pipe(out);
      };
      out.on('data', handshake);
    });
    client.on('close', () => out.destroy());
  });
  server.on('clientError', (_e, socket) => socket.destroy());

  await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
  return {
    port: server.address().port,
    events,
    close: () =>
      new Promise((resolve) => {
        for (const s of sockets) s.destroy();
        server.close(() => resolve());
      }),
  };
}

module.exports = { Guard, GuardError, startProxy, portAllowed };
