-- Send DELIVRD for a customer's messages in a period that have no final DLR, including those the
-- upstream never answered, and queue the DLR to the customer. SMPP logins only. Same changes as
-- the gateway's manual DLR. One transaction; safe to run twice (a second run finds nothing).
-- Variables: customer_email, day_start, day_end. Skips messages sent in the last 10 minutes,
-- anything that failed to send, refunded messages, and messages with any failed part.
BEGIN;

CREATE TEMP TABLE dlr_targets ON COMMIT DROP AS
SELECT DISTINCT m.id
FROM messages m
JOIN service_providers sp ON sp.id = m.service_provider_id
JOIN message_segments ms ON ms.message_id = m.id
JOIN sp_credentials spc ON spc.id = m.sp_credential_id AND spc.protocol = 'smpp'
WHERE sp.email = :'customer_email'
  AND m.submitted_at >= :'day_start'::timestamptz
  AND m.submitted_at <  :'day_end'::timestamptz
  AND m.final_status NOT IN ('delivered', 'failed', 'rejected', 'expired')
  AND ms.sent_to_mno_at IS NOT NULL
  AND ms.sent_to_mno_at < now() - interval '10 minutes'
  AND (ms.mno_message_id IS NOT NULL
       OR ((ms.error_code IS NULL OR ms.error_code = 'TIMEOUT') AND ms.error_description IS NULL)
       OR ms.error_code = 'MNO_TIMEOUT')
  AND (ms.dlr_status IS NULL OR upper(ms.dlr_status) NOT IN ('DELIVRD','UNDELIV','EXPIRED','REJECTD','DELETED'))
  AND NOT EXISTS (SELECT 1 FROM wallet_transactions wt WHERE wt.message_id = m.id AND wt.transaction_type = 'reversal');

-- Only messages whose every waiting part can be completed.
DELETE FROM dlr_targets t
WHERE EXISTS (
  SELECT 1 FROM message_segments ms
  WHERE ms.message_id = t.id
    AND (ms.dlr_status IS NULL OR upper(ms.dlr_status) NOT IN ('DELIVRD','UNDELIV','EXPIRED','REJECTD','DELETED'))
    AND NOT (ms.sent_to_mno_at IS NOT NULL AND (ms.mno_message_id IS NOT NULL
             OR ((ms.error_code IS NULL OR ms.error_code = 'TIMEOUT') AND ms.error_description IS NULL)
             OR ms.error_code = 'MNO_TIMEOUT'))
);

UPDATE message_segments ms
SET dlr_status = 'DELIVRD', dlr_received_at = NOW(), error_code = NULL, dlr_source = 'manual'
WHERE ms.message_id IN (SELECT id FROM dlr_targets)
  AND (ms.dlr_status IS NULL OR upper(ms.dlr_status) NOT IN ('DELIVRD','UNDELIV','EXPIRED','REJECTD','DELETED'));

UPDATE messages m
SET final_status = 'delivered', error_code = NULL, error_description = NULL, completed_at = NOW()
WHERE m.id IN (SELECT id FROM dlr_targets)
  AND NOT EXISTS (SELECT 1 FROM message_segments ms WHERE ms.message_id = m.id AND upper(coalesce(ms.dlr_status,'')) <> 'DELIVRD');

-- The forwarding job, in the JSON the gateway writes (Go field names). ExtraClientMessageIDs makes
-- a joined long SMS send the DLR to every part ID.
INSERT INTO dlr_forwarding_queue (message_id, payload, max_attempts, created_at, status)
SELECT m.id,
       json_build_object(
         'sp_details', json_build_object(
           'ServiceProviderID', m.service_provider_id,
           'Protocol',          spc.protocol,
           'SMPPSystemID',      coalesce(spc.system_id, ''),
           'HTTPCallbackURL',   '',
           'HTTPAuthConfig',    ''),
         'dlr_info', json_build_object(
           'InternalMessageID',     m.id,
           'ClientMessageRef',      m.client_ref,
           'ClientMessageID',       m.client_message_id,
           'ExtraClientMessageIDs', m.extra_client_message_ids,
           'SourceAddr',            m.original_destination_addr,
           'DestAddr',              m.original_source_addr,
           'SubmitDate',            m.submitted_at,
           'DoneDate',              now(),
           'Status',                'DELIVRD',
           'ErrorCode',             NULL,
           'NetworkCode',           NULL,
           'TotalSegments',         m.total_segments)
       )::jsonb,
       10, NOW(), 'pending'
FROM messages m
JOIN sp_credentials spc ON spc.id = m.sp_credential_id
WHERE m.id IN (SELECT id FROM dlr_targets) AND m.final_status = 'delivered';

SELECT (SELECT count(*) FROM dlr_targets) AS messages_completed,
       (SELECT count(*) FROM dlr_forwarding_queue q WHERE q.message_id IN (SELECT id FROM dlr_targets) AND q.status = 'pending') AS dlrs_queued;

COMMIT;
