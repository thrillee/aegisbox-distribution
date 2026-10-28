# Manual and Auto DLR

When an MNO never returns a final DLR, an admin can send one manually, or a service provider can be set up to receive one automatically after a delay. Both behave like a real DLR: the message final status is updated, a failure status reverses the wallet debit, and the DLR is queued in `dlr_forwarding_queue` for forwarding to the SP.

The segment update and the forwarding job are written in one SQL statement, so a DLR is never applied without also being queued for the customer.

## Which segments are affected

A segment is waiting for a final DLR when the MNO accepted it (`mno_message_id` is set) and its `dlr_status` is empty, `no_dlr`, or an intermediate status such as `ENROUTE` or `ACCEPTD`. Only those segments are updated. These are left alone:

* segments with a final status (`DELIVRD`, `UNDELIV`, `EXPIRED`, `REJECTD`, `DELETED`)
* segments that never reached the MNO
* messages already closed as `failed`, `rejected` or `expired`, or already refunded, so the SP never receives a DLR that contradicts a refund

Updated segments get `dlr_source = 'manual'` or `'auto'`. A real DLR that arrives later overwrites the status, clears `dlr_source`, and is forwarded to the SP as usual, so the SP can receive a second DLR. Manual and auto DLRs do not feed sender or route health.

A message's debit is reversed at most once, however many failure DLRs it receives (real, manual or auto).

## Forwarding to the customer

The gateway's DLR forwarding worker sends queued DLRs to the SP (HTTP callback or SMPP `deliver_sm`):

* Each job gets 10 attempts with exponential backoff (5s, 10s, 20s ... capped at 5 minutes), about 20 minutes in total, so a short customer outage does not lose the DLR.
* Jobs left in `processing` for over 5 minutes (worker crash, or the result could not be recorded) are picked up again. The SP may then receive the DLR twice, but never zero times.
* A DLR with no error code is forwarded without one (`0` over HTTP, `000` over SMPP).
* Jobs that used all their attempts are marked `failed`. Retry them after fixing the cause with `POST /api/v1/dlr-forwarding/retry` or the portal DLR Queue page.

If the SP has no HTTP callback URL (or SMPP system ID), the manual DLR response includes `forward_warning`: the DLR is queued but cannot be delivered until the SP is configured and the job is retried.

## Manager API

### Send a manual DLR for one message

`POST /api/v1/messages/:id/dlr`

```json
{ "status": "DELIVRD", "error_code": "000" }
```

`status` is one of `DELIVRD`, `UNDELIV`, `EXPIRED`, `REJECTD`, `DELETED` (case insensitive). `error_code` is optional, max 10 characters.

| Response | Meaning |
|----------|---------|
| 200 `{"message_id", "segments_updated", "forwarded": true, "queue_id", "forward_warning"?}` | DLR applied and queued for the customer |
| 400 | Invalid status or body |
| 404 | Message not found |
| 409 | No segments waiting for a final DLR (or message closed/refunded) |
| 500 | Nothing was changed |

### Send a manual DLR for many messages

`POST /api/v1/messages/dlr/bulk`

```json
{ "message_ids": [101, 102], "status": "DELIVRD", "error_code": "" }
```

Up to 200 IDs per request. The response lists each message with `forwarded`, or an `error` when it was skipped, plus `forwarded`, `skipped`, `failed` and `warnings` totals.

### Retry failed DLR forwarding

`POST /api/v1/dlr-forwarding/retry`

```json
{ "ids": [12, 13] }
```

Puts `failed` jobs back in the queue with fresh attempts. Up to 200 IDs. Returns `requeued` IDs and how many were `skipped` because they were not failed.

### Auto DLR config per service provider

`GET /api/v1/service-providers/:id/auto-dlr`

`PUT /api/v1/service-providers/:id/auto-dlr`

```json
{ "enabled": true, "status": "DELIVRD", "delay_seconds": 600 }
```

`delay_seconds` is how long to wait after the segment was sent to the MNO, between 60 and 172800 (48 hours). The response includes `enabled_at`.

## Auto DLR worker

`AutoDLRWorker` runs in the smpp-gateway every minute. For each SP with auto DLR enabled, it completes segments that have waited longer than the SP's delay. It only considers:

* segments sent after auto DLR was last switched on (`auto_dlr_enabled_at`), so turning it on does not fire DLRs for an old backlog (use the bulk endpoint for those)
* segments sent within the last 7 days

Changing the status or delay keeps `auto_dlr_enabled_at`; turning auto DLR off clears it.

## Database

Migration `00007_manual_and_auto_dlr.sql` adds `auto_dlr_enabled`, `auto_dlr_status`, `auto_dlr_delay_seconds` and `auto_dlr_enabled_at` to `service_providers`, `dlr_source` to `message_segments`, and the partial index `idx_message_segments_awaiting_final_dlr` used by the worker.

## Portal

The admin portal calls these endpoints through `MANAGER_API_URL`:

* Message detail page: **Send DLR**
* Messages list: select messages, then **Send DLR**
* Service provider page: **DLR** tab for auto DLR settings
* DLR Queue page: retry failed jobs (per job, or all failed on the page)
