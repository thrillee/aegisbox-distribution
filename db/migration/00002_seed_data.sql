-- +goose Up
-- +goose StatementBegin

-- Seed initial MNOs
INSERT INTO mnos (id, name, country_code, status) VALUES
(1, 'Nigeria MNOs (Aggregated)', '234', 'active'),
(2, 'Egypt MNOs (Aggregated)', '20', 'active'),
(3, 'Ghana MNOs (Aggregated)', '233', 'active')
ON CONFLICT (id) DO NOTHING;

-- Seed HTTP MNO Connections (from talk-thrill-http providers)
-- SIMPU Provider
INSERT INTO mno_connections (id, mno_id, protocol, status, name, base_url, api_key, rate_limit, supports_webhook) VALUES
(1, 1, 'http', 'active', 'SIMPU', 'https://api.simpu.co', '', 100, true)
ON CONFLICT (id) DO NOTHING;

-- TERMII Provider
INSERT INTO mno_connections (id, mno_id, protocol, status, name, base_url, api_key, rate_limit, supports_webhook) VALUES
(2, 1, 'http', 'active', 'TERMII', 'https://api.termii.com', '', 150, true)
ON CONFLICT (id) DO NOTHING;

-- MULTITEXTER Provider
INSERT INTO mno_connections (id, mno_id, protocol, status, name, base_url, username, password, rate_limit, supports_webhook) VALUES
(3, 1, 'http', 'active', 'MULTITEXTER', 'https://app.multitexter.com/v2', '', '', 100, false)
ON CONFLICT (id) DO NOTHING;

-- AFRICATALKING Provider (restricted networks)
INSERT INTO mno_connections (id, mno_id, protocol, status, name, base_url, api_key, rate_limit, supports_webhook) VALUES
(4, 1, 'http', 'active', 'AFRICATALKING', 'https://api.africastalking.com', '', 120, true)
ON CONFLICT (id) DO NOTHING;

-- EBULKSMS Provider
INSERT INTO mno_connections (id, mno_id, protocol, status, name, base_url, rate_limit, supports_webhook) VALUES
(5, 1, 'http', 'active', 'EBULKSMS', 'https://api.ebulksms.com', 200, true)
ON CONFLICT (id) DO NOTHING;

-- Seed Sender IDs (from talk-thrill-http senders)
-- Get provider IDs for reference
DO $$
DECLARE
    multitexter_id INT := 3;
    termii_id INT := 2;
    africatalking_id INT := 4;
    ebulksms_id INT := 5;
BEGIN
    -- MULTITEXTER Senders
    INSERT INTO sender_ids (sender_id_string, provider_id, rate_limit, otp_template, networks, status) VALUES
    ('EmiProTech', multitexter_id, 100, 'Your verification code is %s.', '{}', 'active'),
    ('EminencePro', multitexter_id, 100, 'Your verification code is %s.', '{}', 'active'),
    ('EProTech', multitexter_id, 100, 'Your verification code is %s.', '{}', 'active'),
    ('Protech', multitexter_id, 100, 'Your verification code is %s.', '{}', 'active'),
    ('MineTech', multitexter_id, 100, 'Your verification code is %s.', '{}', 'active'),
    ('MineProTech', multitexter_id, 100, 'Your verification code is %s.', '{}', 'active'),
    ('EmiCoTech', multitexter_id, 100, 'Your verification code is %s.', '{}', 'active'),
    ('MyProTech', multitexter_id, 100, 'Your verification code is %s.', '{}', 'active'),
    ('MyTechPro', multitexter_id, 100, 'Your verification code is %s.', '{}', 'active'),
    ('Econsult', multitexter_id, 100, 'Your verification code is %s.', '{}', 'active')
    ON CONFLICT (sender_id_string) DO NOTHING;

    -- TERMII Senders
    INSERT INTO sender_ids (sender_id_string, provider_id, rate_limit, otp_template, networks, status) VALUES
    ('EconTech', termii_id, 150, 'Dear customer, your verification code is %s. Use this code to complete the authentication process', '{}', 'active'),
    ('MineConsult', termii_id, 150, 'Dear customer, your verification code is %s. Use this code to complete the authentication process', '{}', 'active'),
    ('EProTech', termii_id, 150, 'Dear customer, your verification code is %s. Use this code to complete the authentication process', '{}', 'active'),
    ('EmiProTech', termii_id, 150, 'Dear customer, your verification code is %s. Use this code to complete the authentication process', '{}', 'active'),
    ('MineProTech', termii_id, 150, 'Dear customer, your verification code is %s. Use this code to complete the authentication process', '{}', 'active'),
    ('MineTech', termii_id, 150, 'Dear customer, your verification code is %s. Use this code to complete the authentication process', '{}', 'active'),
    ('ProTech', termii_id, 150, 'Dear customer, your verification code is %s. Use this code to complete the authentication process', '{}', 'active'),
    ('EminencePro', termii_id, 150, 'Dear customer, your verification code is %s. Use this code to complete the authentication process', '{}', 'active'),
    ('EPConsult', termii_id, 150, 'Dear customer, your verification code is %s. Use this code to complete the authentication process', '{}', 'active')
    ON CONFLICT (sender_id_string) DO NOTHING;

    -- AFRICATALKING Senders (restricted to AIRTEL, GLO)
    INSERT INTO sender_ids (sender_id_string, provider_id, rate_limit, otp_template, networks, status) VALUES
    ('maitechotp', africatalking_id, 120, 'Please use maitechotp: %s to continue.', '{"AIRTEL","GLO"}', 'active'),
    ('maiauth', africatalking_id, 120, 'Please use maiauth: %s to continue.', '{"AIRTEL","GLO"}', 'active'),
    ('maiotp', africatalking_id, 120, 'Please use maiotp: %s to continue.', '{"AIRTEL","GLO"}', 'active')
    ON CONFLICT (sender_id_string) DO NOTHING;

    -- EBULKSMS Senders
    INSERT INTO sender_ids (sender_id_string, provider_id, rate_limit, otp_template, networks, status) VALUES
    ('EminencePro', ebulksms_id, 200, 'Dear customer, your confirmation code is %s valid for 5minutes. Please enter to complete your verification process.', '{}', 'active')
    ON CONFLICT (sender_id_string) DO NOTHING;
END $$;

-- Create System Default Routing Group
INSERT INTO routing_groups (name, description, reference) VALUES
('System Default Route', 'System-wide default routing group for unassigned credentials or general fallback.', 'SYS_DEFAULT_ROUTE')
ON CONFLICT (reference) DO NOTHING;

-- Nigeria: Prefix Groups
INSERT INTO msisdn_prefix_groups (name, description, reference) VALUES
('MTN Nigeria Prefixes', 'All known MSISDN prefixes for MTN Nigeria', 'MTN_NG_PFX'),
('Glo Nigeria Prefixes', 'All known MSISDN prefixes for Globacom Nigeria', 'GLO_NG_PFX'),
('Airtel Nigeria Prefixes', 'All known MSISDN prefixes for Airtel Nigeria', 'AIRTEL_NG_PFX'),
('9mobile Nigeria Prefixes', 'All known MSISDN prefixes for 9mobile Nigeria', '9MOBILE_NG_PFX')
ON CONFLICT (reference) DO NOTHING;

-- Nigeria: MTN Prefixes
INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY[
    '234703', '234704', '234706', '234803', '234806', '234810', '234813', '234814', '234816',
    '234903', '234906', '234913', '234916', '2347025', '2347026'
]) FROM msisdn_prefix_groups WHERE reference = 'MTN_NG_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

-- Nigeria: Glo Prefixes
INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY[
    '234705', '234805', '234807', '234811', '234815', '234905', '234915'
]) FROM msisdn_prefix_groups WHERE reference = 'GLO_NG_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

-- Nigeria: Airtel Prefixes
INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY[
    '234701', '234708', '234802', '234808', '234812', '234901', '234902', '234904', '234907', '234912'
]) FROM msisdn_prefix_groups WHERE reference = 'AIRTEL_NG_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

-- Nigeria: 9mobile Prefixes
INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY[
    '234809', '234817', '234818', '234909', '234908'
]) FROM msisdn_prefix_groups WHERE reference = '9MOBILE_NG_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

-- Ghana: Prefix Groups
INSERT INTO msisdn_prefix_groups (name, description, reference) VALUES
('MTN Ghana Prefixes', 'MSISDN prefixes for MTN Ghana', 'MTN_GH_PFX'),
('Vodafone Ghana Prefixes', 'MSISDN prefixes for Vodafone Ghana', 'VODA_GH_PFX'),
('AirtelTigo Ghana Prefixes', 'MSISDN prefixes for AirtelTigo Ghana', 'AIRTELTIGO_GH_PFX')
ON CONFLICT (reference) DO NOTHING;

-- Ghana: MTN Prefixes
INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY['23324', '23353', '23354', '23355', '23359', '23325'])
FROM msisdn_prefix_groups WHERE reference = 'MTN_GH_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

-- Ghana: Vodafone Prefixes
INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY['23320', '23350'])
FROM msisdn_prefix_groups WHERE reference = 'VODA_GH_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

-- Ghana: AirtelTigo Prefixes
INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY['23327', '23357', '23326', '23356'])
FROM msisdn_prefix_groups WHERE reference = 'AIRTELTIGO_GH_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

-- Egypt: Prefix Groups
INSERT INTO msisdn_prefix_groups (name, description, reference) VALUES
('Vodafone Egypt Prefixes', 'MSISDN prefixes for Vodafone Egypt', 'VODA_EG_PFX'),
('Orange Egypt Prefixes', 'MSISDN prefixes for Orange Egypt', 'ORANGE_EG_PFX'),
('Etisalat Egypt Prefixes', 'MSISDN prefixes for Etisalat Misr Egypt', 'ETISALAT_EG_PFX'),
('WE Egypt Prefixes', 'MSISDN prefixes for Telecom Egypt (WE)', 'WE_EG_PFX')
ON CONFLICT (reference) DO NOTHING;

-- Egypt: Prefix Entries
INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY['2010']) FROM msisdn_prefix_groups WHERE reference = 'VODA_EG_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY['2012']) FROM msisdn_prefix_groups WHERE reference = 'ORANGE_EG_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY['2011']) FROM msisdn_prefix_groups WHERE reference = 'ETISALAT_EG_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY['2015']) FROM msisdn_prefix_groups WHERE reference = 'WE_EG_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

-- Seed OTP Templates
INSERT INTO otp_message_templates (name, content_template, default_brand_name, status) VALUES
('Standard OTP English v1', 'Your verification code for [APP_NAME] is [OTP]. Do not share it. This code is valid for [VALIDITY_MINUTES] minutes. [BRAND_NAME]', 'Aegisbox', 'active'),
('Simple OTP', 'Your OTP is [OTP]', 'Aegisbox', 'active'),
('Secure OTP', 'Security Code: [OTP]. Do not share this code with anyone.', 'Aegisbox', 'active')
ON CONFLICT (name) DO NOTHING;

-- Seed Routing Assignments (System Default)
INSERT INTO routing_assignments (routing_group_id, msisdn_prefix_group_id, mno_id, priority)
SELECT rg.id, mpg.id, 1, 0
FROM routing_groups rg, msisdn_prefix_groups mpg
WHERE rg.reference = 'SYS_DEFAULT_ROUTE' AND mpg.reference IN ('MTN_NG_PFX', 'GLO_NG_PFX', 'AIRTEL_NG_PFX', '9MOBILE_NG_PFX')
ON CONFLICT (routing_group_id, msisdn_prefix_group_id) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove seeded data
DELETE FROM routing_assignments WHERE routing_group_id IN (
    SELECT id FROM routing_groups WHERE reference = 'SYS_DEFAULT_ROUTE'
);

DELETE FROM otp_sender_template_assignments;
DELETE FROM otp_message_templates;
DELETE FROM otp_alternative_senders;

DELETE FROM msisdn_prefix_entries WHERE msisdn_prefix_group_id IN (
    SELECT id FROM msisdn_prefix_groups WHERE reference IN (
        'MTN_NG_PFX', 'GLO_NG_PFX', 'AIRTEL_NG_PFX', '9MOBILE_NG_PFX',
        'MTN_GH_PFX', 'VODA_GH_PFX', 'AIRTELTIGO_GH_PFX',
        'VODA_EG_PFX', 'ORANGE_EG_PFX', 'ETISALAT_EG_PFX', 'WE_EG_PFX'
    )
);

DELETE FROM msisdn_prefix_groups WHERE reference IN (
    'MTN_NG_PFX', 'GLO_NG_PFX', 'AIRTEL_NG_PFX', '9MOBILE_NG_PFX',
    'MTN_GH_PFX', 'VODA_GH_PFX', 'AIRTELTIGO_GH_PFX',
    'VODA_EG_PFX', 'ORANGE_EG_PFX', 'ETISALAT_EG_PFX', 'WE_EG_PFX'
);

DELETE FROM routing_groups WHERE reference = 'SYS_DEFAULT_ROUTE';

DELETE FROM sender_ids WHERE provider_id IN (1, 2, 3, 4, 5);
DELETE FROM mno_connections WHERE id IN (1, 2, 3, 4, 5);
DELETE FROM mnos WHERE id IN (1, 2, 3);

-- +goose StatementEnd
