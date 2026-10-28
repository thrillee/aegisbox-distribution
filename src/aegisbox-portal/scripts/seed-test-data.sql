-- Insert test MNOs
INSERT INTO mnos (name, code, slug) VALUES
('MTN Nigeria', 'MTN', 'mtn-ng'),
('Airtel Nigeria', 'AIRTEL', 'airtel-ng'),
('Glo Mobile', 'GLO', 'glo-ng'),
('9mobile', '9MOBILE', '9mobile-ng')
ON CONFLICT DO NOTHING;

-- Insert test admins (you'll need to replace with actual Clerk IDs)
INSERT INTO admins (name, email, clerk_id, status, created_at) VALUES
('Admin User', 'admin@aegisbox.com', 'admin-test-id', 'active', NOW())
ON CONFLICT DO NOTHING;

-- Insert test service providers (you'll need to replace with actual Clerk IDs)
INSERT INTO service_providers (name, email, clerk_id, business_type, wallet_balance, status, created_at) VALUES
('Test Provider 1', 'provider1@test.com', 'provider-test-id-1', 'bulk-sms', 50000, 'active', NOW()),
('Test Provider 2', 'provider2@test.com', 'provider-test-id-2', 'notifications', 25000, 'active', NOW())
ON CONFLICT DO NOTHING;

-- Link providers to MNOs
INSERT INTO provider_mno_mappings (provider_id, mno_id, created_at) 
SELECT sp.id, m.id, NOW() 
FROM service_providers sp, mnos m
WHERE sp.email LIKE 'provider%@test.com'
ON CONFLICT DO NOTHING;

-- Insert test messages
INSERT INTO messages (provider_id, recipient, content, mno_id, status, delivery_status, created_at)
SELECT sp.id, '+2348012345678', 'Test message content', m.id, 'sent', 'delivered', NOW()
FROM service_providers sp, mnos m
WHERE sp.email LIKE 'provider%@test.com'
LIMIT 1
ON CONFLICT DO NOTHING;

-- Insert test transactions
INSERT INTO transactions (provider_id, type, amount, description, status, created_at)
SELECT sp.id, 'debit', 50, 'SMS charge', 'completed', NOW()
FROM service_providers sp
WHERE sp.email = 'provider1@test.com'
LIMIT 1
ON CONFLICT DO NOTHING;
