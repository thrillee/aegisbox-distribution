# Operational SQL

Scripts for the production database, which agents cannot reach directly. The owner runs them with
psql. Each takes psql variables, so nothing customer specific is stored here:

```bash
psql "$DATABASE_URL" -v customer_email='someone@example.com' -v day_start='2026-10-30 00:00:00+01' \
     -v day_end='2026-10-31 00:00:00+01' -f diagnose_waiting.sql
```

Days are Lagos days: write them as `YYYY-MM-DD 00:00:00+01`.

| Script | Read only | What it does |
|---|---|---|
| `diagnose_waiting.sql` | yes | Shows why a customer's messages in a period have no final DLR (accepted by the upstream, no answer, failed to send). |
| `send_delivrd_for_waiting.sql` | no | Marks the customer's waiting messages DELIVRD and queues the DLR to them (SMPP logins only), in one transaction, safe to run twice. Same rules as the portal's Pending DLRs page with "no answer" included. |

Prefer the portal (Messaging > Pending DLRs, or Auto DLR on the customer's DLR tab) when it is
deployed; the scripts are for when the portal cannot be used.
