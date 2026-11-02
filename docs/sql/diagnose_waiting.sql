-- Why does a customer have messages without a final DLR? Read only.
-- Variables: customer_email, day_start, day_end.
SELECT m.final_status,
       CASE
         WHEN ms.dlr_status IS NOT NULL AND upper(ms.dlr_status) IN ('DELIVRD','UNDELIV','EXPIRED','REJECTD','DELETED') THEN 'final DLR'
         WHEN ms.sent_to_mno_at IS NULL THEN 'not sent yet'
         WHEN ms.mno_message_id IS NOT NULL THEN 'accepted by upstream, no DLR'
         WHEN ((ms.error_code IS NULL OR ms.error_code = 'TIMEOUT') AND ms.error_description IS NULL)
              OR ms.error_code = 'MNO_TIMEOUT' THEN 'no answer from upstream'
         ELSE 'failed to send: ' || coalesce(ms.error_code, '-') || ' ' || left(coalesce(ms.error_description, ''), 60)
       END AS segment_state,
       count(DISTINCT m.id) AS messages
FROM messages m
JOIN service_providers sp ON sp.id = m.service_provider_id
JOIN message_segments ms ON ms.message_id = m.id
WHERE sp.email = :'customer_email'
  AND m.submitted_at >= :'day_start'::timestamptz
  AND m.submitted_at <  :'day_end'::timestamptz
GROUP BY 1, 2
ORDER BY 3 DESC;
