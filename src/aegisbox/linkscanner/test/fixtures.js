// A local test site with safe and hostile pages, for the scanner tests and manual checks.
'use strict';
const http = require('node:http');

function startFixtures() {
  const hits = { posts: 0, paths: [] };
  const server = http.createServer((req, res) => {
    hits.paths.push(`${req.method} ${req.url}`);
    if (req.method !== 'GET' && req.method !== 'HEAD') {
      hits.posts++;
      res.writeHead(200).end('ok');
      return;
    }
    const html = (body, head = '') => {
      res.writeHead(200, { 'Content-Type': 'text/html' });
      res.end(`<!doctype html><html><head>${head}</head><body>${body}</body></html>`);
    };
    const host = req.headers.host;
    switch (req.url.split('?')[0]) {
      case '/safe':
        return html('<h1>Welcome to our shop</h1><p>Shoes on sale this week.</p>', '<title>Shoe Shop</title>');
      case '/r1': res.writeHead(302, { Location: '/r2' }); return res.end();
      case '/r2': res.writeHead(301, { Location: '/r3' }); return res.end();
      case '/r3': res.writeHead(302, { Location: '/safe' }); return res.end();
      case '/js':
        return html('<p>Redirecting</p><script>setTimeout(()=>location.href="/safe",200)</script>');
      case '/meta':
        return html('<p>Redirecting</p>', '<meta http-equiv="refresh" content="0;url=/safe">');
      case '/bank':
        return html(
          `<h2>GTBank Internet Banking</h2><p>Your account has been restricted. Verify your BVN.</p>
           <form method="post" action="/collect">
             <label for="u">Account number</label><input id="u" name="account">
             <input type="password" name="password" placeholder="Password">
             <input name="card_number" placeholder="Card number">
             <input name="pin" placeholder="ATM PIN">
             <input name="bvn" placeholder="BVN">
             <button type="submit">Login</button>
           </form>
           <script>document.forms[0].submit()</script>`,
          '<title>GTBank | Secure Login</title>',
        );
      case '/apk':
        res.writeHead(200, { 'Content-Type': 'application/vnd.android.package-archive', 'Content-Disposition': 'attachment; filename="update.apk"' });
        return res.end('PK fake apk');
      case '/apkpage':
        return html('<a id="d" href="/apk">Download</a><script>document.getElementById("d").click()</script>', '<title>Get the app</title>');
      case '/loop': res.writeHead(302, { Location: '/loop?x=' + Math.random() }); return res.end();
      case '/internal': res.writeHead(302, { Location: 'http://169.254.169.254/latest/meta-data/' }); return res.end();
      case '/internal-name': res.writeHead(302, { Location: `http://internal.test:${host.split(':')[1]}/safe` }); return res.end();
      case '/sub':
        return html(`<p>Page with hidden requests</p><img src="http://10.0.0.1/pixel.png">
          <script>fetch("/collect",{method:"POST",body:"x"}).catch(()=>{});
          try{new WebSocket("ws://"+location.host+"/ws")}catch(e){}</script>`, '<title>Sub</title>');
      case '/geo':
        res.writeHead(403, { 'Content-Type': 'text/html' });
        return res.end('<html><body>Sorry, this service is not available in your country.</body></html>');
      case '/hang':
        return; // never answers
      default:
        res.writeHead(404).end('not found');
    }
  });
  return new Promise((resolve) => server.listen(0, '127.0.0.1', () => resolve({ port: server.address().port, hits, close: () => { server.closeAllConnections(); server.close(); } })));
}

module.exports = { startFixtures };
