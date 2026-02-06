-- +goose Up
-- +goose StatementBegin

-- Create a System Default Routing Group (if it doesn't exist)
INSERT INTO routing_groups (name, description, reference) VALUES
('System Default Route', 'System-wide default routing group for unassigned credentials or general fallback.', 'SYS_DEFAULT_ROUTE')
ON CONFLICT (reference) DO NOTHING; -- Avoids error if run multiple times or if it exists

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
ON CONFLICT (msisdn_prefix) DO NOTHING; -- Or ON CONFLICT (msisdn_prefix_group_id, msisdn_prefix) DO NOTHING

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


-- Ghana: Prefix Groups (Country Code 233)
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
SELECT id, unnest(ARRAY['23327', '23357', '23326', '23356']) -- Combining former Tigo and Airtel
FROM msisdn_prefix_groups WHERE reference = 'AIRTELTIGO_GH_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;


-- Egypt: Prefix Groups (Country Code 20)
INSERT INTO msisdn_prefix_groups (name, description, reference) VALUES
('Vodafone Egypt Prefixes', 'MSISDN prefixes for Vodafone Egypt', 'VODA_EG_PFX'),
('Orange Egypt Prefixes', 'MSISDN prefixes for Orange Egypt', 'ORANGE_EG_PFX'),
('Etisalat Egypt Prefixes', 'MSISDN prefixes for Etisalat Misr Egypt', 'ETISALAT_EG_PFX'),
('WE Egypt Prefixes', 'MSISDN prefixes for Telecom Egypt (WE)', 'WE_EG_PFX')
ON CONFLICT (reference) DO NOTHING;

-- Egypt: Vodafone Prefixes
INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY['2010']) FROM msisdn_prefix_groups WHERE reference = 'VODA_EG_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

-- Egypt: Orange Prefixes
INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY['2012']) FROM msisdn_prefix_groups WHERE reference = 'ORANGE_EG_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

-- Egypt: Etisalat Prefixes
INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY['2011']) FROM msisdn_prefix_groups WHERE reference = 'ETISALAT_EG_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

-- Egypt: WE Prefixes
INSERT INTO msisdn_prefix_entries (msisdn_prefix_group_id, msisdn_prefix)
SELECT id, unnest(ARRAY['2015']) FROM msisdn_prefix_groups WHERE reference = 'WE_EG_PFX'
ON CONFLICT (msisdn_prefix) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Deleting seed data can be selective or complete.
-- Here's a selective approach based on references.
DELETE FROM routing_assignments WHERE msisdn_prefix_group_id IN (
    SELECT id from msisdn_prefix_groups WHERE reference IN (
        'MTN_NG_PFX', 'GLO_NG_PFX', 'AIRTEL_NG_PFX', '9MOBILE_NG_PFX',
        'MTN_GH_PFX', 'VODA_GH_PFX', 'AIRTELTIGO_GH_PFX',
        'VODA_EG_PFX', 'ORANGE_EG_PFX', 'ETISALAT_EG_PFX', 'WE_EG_PFX'
    )
);

DELETE FROM msisdn_prefix_entries WHERE msisdn_prefix_group_id IN (
    SELECT id from msisdn_prefix_groups WHERE reference IN (
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

-- +goose StatementEnd
