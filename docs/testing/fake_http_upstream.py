#!/usr/bin/env python3
"""A stand in HTTP upstream: records each request's sender and recipient to a log file and returns
a message ID. While the file /tmp/fail_<name> exists it returns HTTP 500 (to test failover and
route health). Usage: fake_http_upstream.py <port> <name> <logfile>"""
import json, os, sys, uuid
from http.server import BaseHTTPRequestHandler, HTTPServer

port, name, log = int(sys.argv[1]), sys.argv[2], sys.argv[3]

class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        body = json.loads(self.rfile.read(int(self.headers.get('Content-Length', 0))) or b'{}')
        failing = os.path.exists(f'/tmp/fail_{name}')
        with open(log, 'a') as f:
            f.write(json.dumps({'route': name, 'sender': body.get('sender_id'), 'to': body.get('recipients'), 'failed': failing}) + '\n')
        if failing:
            self.send_response(500); self.end_headers(); self.wfile.write(b'{"error":"down"}'); return
        self.send_response(200); self.send_header('Content-Type', 'application/json'); self.end_headers()
        self.wfile.write(json.dumps({'id': str(uuid.uuid4()), 'recipients': [body.get('recipients')]}).encode())

    def do_GET(self):
        self.send_response(200); self.end_headers(); self.wfile.write(b'{}')

    def log_message(self, *args):
        pass

HTTPServer(('127.0.0.1', port), Handler).serve_forever()
