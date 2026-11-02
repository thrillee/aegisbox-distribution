#!/usr/bin/env python3
"""Bind to the Aegisbox customer SMPP server as a customer, send test messages, print the message
IDs returned and the DLRs received. Needs: pip install smpplib.

Modes:
  single     one short message
  multipart  one long message as 2 parts (concatenation UDH)
  dup        the same, with part 1 sent twice (should get the same ID back)
  missing    part 1 of a 3 part message only (sent after the gateway's 2 minute timeout)
"""
import argparse, re, time
import smpplib.client, smpplib.consts, smpplib.gsm

ap = argparse.ArgumentParser()
ap.add_argument('--host', default='127.0.0.1')
ap.add_argument('--port', type=int, default=12775)
ap.add_argument('--system-id', required=True)
ap.add_argument('--password', required=True)
ap.add_argument('--sender', default='TEST')
ap.add_argument('--to', default='2348031234567')
ap.add_argument('--mode', choices=['single', 'multipart', 'dup', 'missing'], default='single')
ap.add_argument('--wait', type=float, default=15, help='seconds to wait for DLRs')
a = ap.parse_args()

ids, dlrs = [], []
c = smpplib.client.Client(a.host, a.port)
c.connect()
c.set_message_sent_handler(lambda pdu: ids.append(pdu.message_id.decode()))
c.set_message_received_handler(lambda pdu: dlrs.append(pdu.short_message.decode()))
c.bind_transceiver(system_id=a.system_id, password=a.password)

if a.mode == 'single':
    parts, enc, esm = smpplib.gsm.make_parts('Test message https://example.com/a')
elif a.mode == 'missing':
    parts, enc, esm = smpplib.gsm.make_parts('Missing part test ' + 'q' * 300)
    parts = parts[:1]
else:
    parts, enc, esm = smpplib.gsm.make_parts('Part test ' + 'abcdefghij' * 15 + ' https://example.com/x')
    if a.mode == 'dup':
        parts = parts[:1] + parts
for part in parts:
    c.send_message(source_addr_ton=5, source_addr=a.sender, dest_addr_ton=1, dest_addr_npi=1,
                   destination_addr=a.to, short_message=part, data_coding=enc,
                   esm_class=smpplib.consts.SMPP_MSGMODE_DEFAULT | esm, registered_delivery=True)
    c.read_once()

c._socket.settimeout(1)
end = time.time() + a.wait
while time.time() < end:
    try:
        c.read_once()
    except Exception:
        pass
print('message IDs returned:', ids)
for d in dlrs:
    m = re.search(r'id:(\S+).*stat:(\S+) err:(\S+)', d)
    print('DLR:', m.groups() if m else d)
c.unbind()
c.disconnect()
