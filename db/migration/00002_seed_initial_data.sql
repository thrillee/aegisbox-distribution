-- +goose Up
-- +goose StatementBegin

-- Seed initial MNOs
-- Use INSERT ... ON CONFLICT DO NOTHING to avoid errors if run multiple times or if IDs exist.
INSERT INTO mnos (id, name, country_code, status) VALUES (1, 'Nigeria MNOs (Blacksilicon)', '234', 'active') ON CONFLICT (id) DO NOTHING;
INSERT INTO mnos (id, name, country_code, status) VALUES (2, 'Egypt MNOs (Aggregated)', '20', 'active') ON CONFLICT (id) DO NOTHING;
-- Add specific MNOs if needed later, e.g., INSERT INTO mnos (name, country_code, network_code) VALUES ('MTN NG', '234', 'MTN');

-- Seed Example Nigerian Routing Rules (Assign to MNO ID 1)
-- Use INSERT ... ON CONFLICT DO NOTHING or ON CONFLICT (prefix) DO UPDATE (if you want updates)
-- Priorities can be adjusted. Longer prefixes usually get higher priority implicitly via ORDER BY length DESC.
INSERT INTO routing_rules (prefix, mno_id, priority) VALUES
('234803', 1, 0), ('234806', 1, 0), ('234703', 1, 0), ('234706', 1, 0), ('234810', 1, 0), ('234813', 1, 0), ('234814', 1, 0), ('234816', 1, 0), ('234903', 1, 0), ('234906', 1, 0), ('234913', 1, 0), ('234916', 1, 0), -- MTN NG Examples
('234805', 1, 0), ('234807', 1, 0), ('234705', 1, 0), ('234811', 1, 0), ('234815', 1, 0), ('234905', 1, 0), ('234915', 1, 0), -- Glo NG Examples
('234802', 1, 0), ('234808', 1, 0), ('234701', 1, 0), ('234708', 1, 0), ('234812', 1, 0), ('234901', 1, 0), ('234902', 1, 0), ('234904', 1, 0), ('234907', 1, 0), ('234912', 1, 0), -- Airtel NG Examples
('234809', 1, 0), ('234817', 1, 0), ('234818', 1, 0), ('234908', 1, 0), ('234909', 1, 0) -- 9mobile NG Examples
ON CONFLICT (prefix) DO NOTHING; -- Avoid errors if prefixes already exist


-- Seed Example Egypt Routing Rules (Assign to MNO ID 2)
INSERT INTO routing_rules (prefix, mno_id, priority) VALUES
('2010', 2, 0), -- Vodafone EG Examples
('2011', 2, 0), -- Etisalat EG Examples
('2012', 2, 0), -- Orange EG Examples
('2015', 2, 0)  -- WE (Telecom Egypt) Examples
ON CONFLICT (prefix) DO NOTHING;

-- Seed Country-level routes (Lower priority)
INSERT INTO routing_rules (prefix, mno_id, priority) VALUES
('234', 1, 100), -- Route all other 234 to Nigeria Aggregated MNO
('20', 2, 100)   -- Route all other 20 to Egypt Aggregated MNO
ON CONFLICT (prefix) DO NOTHING;


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove seeded data - Be careful with DELETE based on specific IDs/prefixes if other data exists.
-- It's often safer to not include DELETEs in seed down migrations unless absolutely necessary.
-- DELETE FROM routing_rules WHERE mno_id IN (1, 2); -- Example, might delete too much
-- DELETE FROM mnos WHERE id IN (1, 2); -- Example

-- +goose StatementEnd
