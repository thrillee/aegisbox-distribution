-- Jennie-SMS (SP ID=6) / ICN(PRO-1) (MNO ID=9) OTP Alternative Senders Seed
-- Generated: 2026-04-28
-- DO NOT RUN THIS SCRIPT MORE THAN ONCE

BEGIN;

DO $$
DECLARE
    v_sender_id INT;
    v_tmpl_digits_id INT;
    v_tmpl_words_id INT;
BEGIN

    -- === AAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AAI - OTP Digits', 'AAI: Kindly go ahead with [OTP]', 'AAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AAI - OTP Words', 'AAI: Kindly go ahead with {{otp_words}}', 'AAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ABase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ABase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ABase - OTP Digits', 'ABase: Kindly go ahead with [OTP]', 'ABase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ABase - OTP Words', 'ABase: Kindly go ahead with {{otp_words}}', 'ABase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ACloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ACloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ACloud - OTP Digits', 'ACloud: Kindly go ahead with [OTP]', 'ACloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ACloud - OTP Words', 'ACloud: Kindly go ahead with {{otp_words}}', 'ACloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ACore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ACore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ACore - OTP Digits', 'ACore: Kindly go ahead with [OTP]', 'ACore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ACore - OTP Words', 'ACore: Kindly go ahead with {{otp_words}}', 'ACore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ADrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ADrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ADrive - OTP Digits', 'ADrive: Kindly go ahead with [OTP]', 'ADrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ADrive - OTP Words', 'ADrive: Kindly go ahead with {{otp_words}}', 'ADrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AEdge - OTP Digits', 'AEdge: Kindly go ahead with [OTP]', 'AEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AEdge - OTP Words', 'AEdge: Kindly go ahead with {{otp_words}}', 'AEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AFlow - OTP Digits', 'AFlow: Kindly go ahead with [OTP]', 'AFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AFlow - OTP Words', 'AFlow: Kindly go ahead with {{otp_words}}', 'AFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AGo - OTP Digits', 'AGo: Kindly go ahead with [OTP]', 'AGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AGo - OTP Words', 'AGo: Kindly go ahead with {{otp_words}}', 'AGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AHQ - OTP Digits', 'AHQ: Kindly go ahead with [OTP]', 'AHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AHQ - OTP Words', 'AHQ: Kindly go ahead with {{otp_words}}', 'AHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AHub - OTP Digits', 'AHub: Kindly go ahead with [OTP]', 'AHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AHub - OTP Words', 'AHub: Kindly go ahead with {{otp_words}}', 'AHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AIBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AIBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIBase - OTP Digits', 'AIBase: Kindly go ahead with [OTP]', 'AIBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIBase - OTP Words', 'AIBase: Kindly go ahead with {{otp_words}}', 'AIBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AICloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AICloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AICloud - OTP Digits', 'AICloud: Kindly go ahead with [OTP]', 'AICloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AICloud - OTP Words', 'AICloud: Kindly go ahead with {{otp_words}}', 'AICloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AICore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AICore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AICore - OTP Digits', 'AICore: Kindly go ahead with [OTP]', 'AICore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AICore - OTP Words', 'AICore: Kindly go ahead with {{otp_words}}', 'AICore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AIDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AIDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIDrive - OTP Digits', 'AIDrive: Kindly go ahead with [OTP]', 'AIDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIDrive - OTP Words', 'AIDrive: Kindly go ahead with {{otp_words}}', 'AIDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AIEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AIEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIEdge - OTP Digits', 'AIEdge: Kindly go ahead with [OTP]', 'AIEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIEdge - OTP Words', 'AIEdge: Kindly go ahead with {{otp_words}}', 'AIEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AIFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AIFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIFlow - OTP Digits', 'AIFlow: Kindly go ahead with [OTP]', 'AIFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIFlow - OTP Words', 'AIFlow: Kindly go ahead with {{otp_words}}', 'AIFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AIGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AIGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIGo - OTP Digits', 'AIGo: Kindly go ahead with [OTP]', 'AIGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIGo - OTP Words', 'AIGo: Kindly go ahead with {{otp_words}}', 'AIGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AIHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AIHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIHQ - OTP Digits', 'AIHQ: Kindly go ahead with [OTP]', 'AIHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIHQ - OTP Words', 'AIHQ: Kindly go ahead with {{otp_words}}', 'AIHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AIHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AIHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIHub - OTP Digits', 'AIHub: Kindly go ahead with [OTP]', 'AIHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIHub - OTP Words', 'AIHub: Kindly go ahead with {{otp_words}}', 'AIHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AILabs ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AILabs', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AILabs - OTP Digits', 'AILabs: Kindly go ahead with [OTP]', 'AILabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AILabs - OTP Words', 'AILabs: Kindly go ahead with {{otp_words}}', 'AILabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AILink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AILink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AILink - OTP Digits', 'AILink: Kindly go ahead with [OTP]', 'AILink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AILink - OTP Words', 'AILink: Kindly go ahead with {{otp_words}}', 'AILink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AINet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AINet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AINet - OTP Digits', 'AINet: Kindly go ahead with [OTP]', 'AINet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AINet - OTP Words', 'AINet: Kindly go ahead with {{otp_words}}', 'AINet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AINode ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AINode', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AINode - OTP Digits', 'AINode: Kindly go ahead with [OTP]', 'AINode', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AINode - OTP Words', 'AINode: Kindly go ahead with {{otp_words}}', 'AINode', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AIOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AIOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIOne - OTP Digits', 'AIOne: Kindly go ahead with [OTP]', 'AIOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIOne - OTP Words', 'AIOne: Kindly go ahead with {{otp_words}}', 'AIOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AIOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AIOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIOps - OTP Digits', 'AIOps: Kindly go ahead with [OTP]', 'AIOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIOps - OTP Words', 'AIOps: Kindly go ahead with {{otp_words}}', 'AIOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AIPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AIPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIPro - OTP Digits', 'AIPro: Kindly go ahead with [OTP]', 'AIPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIPro - OTP Words', 'AIPro: Kindly go ahead with {{otp_words}}', 'AIPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AIStack ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AIStack', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIStack - OTP Digits', 'AIStack: Kindly go ahead with [OTP]', 'AIStack', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIStack - OTP Words', 'AIStack: Kindly go ahead with {{otp_words}}', 'AIStack', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AISys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AISys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AISys - OTP Digits', 'AISys: Kindly go ahead with [OTP]', 'AISys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AISys - OTP Words', 'AISys: Kindly go ahead with {{otp_words}}', 'AISys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AIX ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AIX', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIX - OTP Digits', 'AIX: Kindly go ahead with [OTP]', 'AIX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIX - OTP Words', 'AIX: Kindly go ahead with {{otp_words}}', 'AIX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AIZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AIZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIZone - OTP Digits', 'AIZone: Kindly go ahead with [OTP]', 'AIZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AIZone - OTP Words', 'AIZone: Kindly go ahead with {{otp_words}}', 'AIZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ALabs ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ALabs', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ALabs - OTP Digits', 'ALabs: Kindly go ahead with [OTP]', 'ALabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ALabs - OTP Words', 'ALabs: Kindly go ahead with {{otp_words}}', 'ALabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ALink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ALink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ALink - OTP Digits', 'ALink: Kindly go ahead with [OTP]', 'ALink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ALink - OTP Words', 'ALink: Kindly go ahead with {{otp_words}}', 'ALink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ANet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ANet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ANet - OTP Digits', 'ANet: Kindly go ahead with [OTP]', 'ANet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ANet - OTP Words', 'ANet: Kindly go ahead with {{otp_words}}', 'ANet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AOne - OTP Digits', 'AOne: Kindly go ahead with [OTP]', 'AOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AOne - OTP Words', 'AOne: Kindly go ahead with {{otp_words}}', 'AOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AOps - OTP Digits', 'AOps: Kindly go ahead with [OTP]', 'AOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AOps - OTP Words', 'AOps: Kindly go ahead with {{otp_words}}', 'AOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === APro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('APro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: APro - OTP Digits', 'APro: Kindly go ahead with [OTP]', 'APro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: APro - OTP Words', 'APro: Kindly go ahead with {{otp_words}}', 'APro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === Astra ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('Astra', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: Astra - OTP Digits', 'Astra: Kindly go ahead with [OTP]', 'Astra', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: Astra - OTP Words', 'Astra: Kindly go ahead with {{otp_words}}', 'Astra', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraAI - OTP Digits', 'AstraAI: Kindly go ahead with [OTP]', 'AstraAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraAI - OTP Words', 'AstraAI: Kindly go ahead with {{otp_words}}', 'AstraAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraCloud - OTP Digits', 'AstraCloud: Kindly go ahead with [OTP]', 'AstraCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraCloud - OTP Words', 'AstraCloud: Kindly go ahead with {{otp_words}}', 'AstraCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraCore - OTP Digits', 'AstraCore: Kindly go ahead with [OTP]', 'AstraCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraCore - OTP Words', 'AstraCore: Kindly go ahead with {{otp_words}}', 'AstraCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraDrive - OTP Digits', 'AstraDrive: Kindly go ahead with [OTP]', 'AstraDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraDrive - OTP Words', 'AstraDrive: Kindly go ahead with {{otp_words}}', 'AstraDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraEdge - OTP Digits', 'AstraEdge: Kindly go ahead with [OTP]', 'AstraEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraEdge - OTP Words', 'AstraEdge: Kindly go ahead with {{otp_words}}', 'AstraEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraFlow - OTP Digits', 'AstraFlow: Kindly go ahead with [OTP]', 'AstraFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraFlow - OTP Words', 'AstraFlow: Kindly go ahead with {{otp_words}}', 'AstraFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraGo - OTP Digits', 'AstraGo: Kindly go ahead with [OTP]', 'AstraGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraGo - OTP Words', 'AstraGo: Kindly go ahead with {{otp_words}}', 'AstraGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraHQ - OTP Digits', 'AstraHQ: Kindly go ahead with [OTP]', 'AstraHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraHQ - OTP Words', 'AstraHQ: Kindly go ahead with {{otp_words}}', 'AstraHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraHub - OTP Digits', 'AstraHub: Kindly go ahead with [OTP]', 'AstraHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraHub - OTP Words', 'AstraHub: Kindly go ahead with {{otp_words}}', 'AstraHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraLab ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraLab', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraLab - OTP Digits', 'AstraLab: Kindly go ahead with [OTP]', 'AstraLab', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraLab - OTP Words', 'AstraLab: Kindly go ahead with {{otp_words}}', 'AstraLab', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraLink - OTP Digits', 'AstraLink: Kindly go ahead with [OTP]', 'AstraLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraLink - OTP Words', 'AstraLink: Kindly go ahead with {{otp_words}}', 'AstraLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraNet - OTP Digits', 'AstraNet: Kindly go ahead with [OTP]', 'AstraNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraNet - OTP Words', 'AstraNet: Kindly go ahead with {{otp_words}}', 'AstraNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraOne - OTP Digits', 'AstraOne: Kindly go ahead with [OTP]', 'AstraOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraOne - OTP Words', 'AstraOne: Kindly go ahead with {{otp_words}}', 'AstraOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraOps - OTP Digits', 'AstraOps: Kindly go ahead with [OTP]', 'AstraOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraOps - OTP Words', 'AstraOps: Kindly go ahead with {{otp_words}}', 'AstraOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraPro - OTP Digits', 'AstraPro: Kindly go ahead with [OTP]', 'AstraPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraPro - OTP Words', 'AstraPro: Kindly go ahead with {{otp_words}}', 'AstraPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraSys - OTP Digits', 'AstraSys: Kindly go ahead with [OTP]', 'AstraSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraSys - OTP Words', 'AstraSys: Kindly go ahead with {{otp_words}}', 'AstraSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraTech - OTP Digits', 'AstraTech: Kindly go ahead with [OTP]', 'AstraTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraTech - OTP Words', 'AstraTech: Kindly go ahead with {{otp_words}}', 'AstraTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraX ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraX', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraX - OTP Digits', 'AstraX: Kindly go ahead with [OTP]', 'AstraX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraX - OTP Words', 'AstraX: Kindly go ahead with {{otp_words}}', 'AstraX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstraZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstraZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraZone - OTP Digits', 'AstraZone: Kindly go ahead with [OTP]', 'AstraZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstraZone - OTP Words', 'AstraZone: Kindly go ahead with {{otp_words}}', 'AstraZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === Astro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('Astro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: Astro - OTP Digits', 'Astro: Kindly go ahead with [OTP]', 'Astro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: Astro - OTP Words', 'Astro: Kindly go ahead with {{otp_words}}', 'Astro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroAI - OTP Digits', 'AstroAI: Kindly go ahead with [OTP]', 'AstroAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroAI - OTP Words', 'AstroAI: Kindly go ahead with {{otp_words}}', 'AstroAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroBase - OTP Digits', 'AstroBase: Kindly go ahead with [OTP]', 'AstroBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroBase - OTP Words', 'AstroBase: Kindly go ahead with {{otp_words}}', 'AstroBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroCloud - OTP Digits', 'AstroCloud: Kindly go ahead with [OTP]', 'AstroCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroCloud - OTP Words', 'AstroCloud: Kindly go ahead with {{otp_words}}', 'AstroCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroCore - OTP Digits', 'AstroCore: Kindly go ahead with [OTP]', 'AstroCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroCore - OTP Words', 'AstroCore: Kindly go ahead with {{otp_words}}', 'AstroCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroDrive - OTP Digits', 'AstroDrive: Kindly go ahead with [OTP]', 'AstroDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroDrive - OTP Words', 'AstroDrive: Kindly go ahead with {{otp_words}}', 'AstroDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroEdge - OTP Digits', 'AstroEdge: Kindly go ahead with [OTP]', 'AstroEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroEdge - OTP Words', 'AstroEdge: Kindly go ahead with {{otp_words}}', 'AstroEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroFlow - OTP Digits', 'AstroFlow: Kindly go ahead with [OTP]', 'AstroFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroFlow - OTP Words', 'AstroFlow: Kindly go ahead with {{otp_words}}', 'AstroFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroGo - OTP Digits', 'AstroGo: Kindly go ahead with [OTP]', 'AstroGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroGo - OTP Words', 'AstroGo: Kindly go ahead with {{otp_words}}', 'AstroGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroHub - OTP Digits', 'AstroHub: Kindly go ahead with [OTP]', 'AstroHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroHub - OTP Words', 'AstroHub: Kindly go ahead with {{otp_words}}', 'AstroHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroLab ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroLab', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroLab - OTP Digits', 'AstroLab: Kindly go ahead with [OTP]', 'AstroLab', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroLab - OTP Words', 'AstroLab: Kindly go ahead with {{otp_words}}', 'AstroLab', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroLink - OTP Digits', 'AstroLink: Kindly go ahead with [OTP]', 'AstroLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroLink - OTP Words', 'AstroLink: Kindly go ahead with {{otp_words}}', 'AstroLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroNet - OTP Digits', 'AstroNet: Kindly go ahead with [OTP]', 'AstroNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroNet - OTP Words', 'AstroNet: Kindly go ahead with {{otp_words}}', 'AstroNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroOne - OTP Digits', 'AstroOne: Kindly go ahead with [OTP]', 'AstroOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroOne - OTP Words', 'AstroOne: Kindly go ahead with {{otp_words}}', 'AstroOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroOps - OTP Digits', 'AstroOps: Kindly go ahead with [OTP]', 'AstroOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroOps - OTP Words', 'AstroOps: Kindly go ahead with {{otp_words}}', 'AstroOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroPro - OTP Digits', 'AstroPro: Kindly go ahead with [OTP]', 'AstroPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroPro - OTP Words', 'AstroPro: Kindly go ahead with {{otp_words}}', 'AstroPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroSys - OTP Digits', 'AstroSys: Kindly go ahead with [OTP]', 'AstroSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroSys - OTP Words', 'AstroSys: Kindly go ahead with {{otp_words}}', 'AstroSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroTech - OTP Digits', 'AstroTech: Kindly go ahead with [OTP]', 'AstroTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroTech - OTP Words', 'AstroTech: Kindly go ahead with {{otp_words}}', 'AstroTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroX ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroX', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroX - OTP Digits', 'AstroX: Kindly go ahead with [OTP]', 'AstroX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroX - OTP Words', 'AstroX: Kindly go ahead with {{otp_words}}', 'AstroX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AstroZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AstroZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroZone - OTP Digits', 'AstroZone: Kindly go ahead with [OTP]', 'AstroZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AstroZone - OTP Words', 'AstroZone: Kindly go ahead with {{otp_words}}', 'AstroZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ASys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ASys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ASys - OTP Digits', 'ASys: Kindly go ahead with [OTP]', 'ASys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ASys - OTP Words', 'ASys: Kindly go ahead with {{otp_words}}', 'ASys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ATech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ATech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ATech - OTP Digits', 'ATech: Kindly go ahead with [OTP]', 'ATech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ATech - OTP Words', 'ATech: Kindly go ahead with {{otp_words}}', 'ATech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === AZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('AZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AZone - OTP Digits', 'AZone: Kindly go ahead with [OTP]', 'AZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: AZone - OTP Words', 'AZone: Kindly go ahead with {{otp_words}}', 'AZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CAIpack ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CAIpack', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CAIpack - OTP Digits', 'CAIpack: Kindly go ahead with [OTP]', 'CAIpack', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CAIpack - OTP Words', 'CAIpack: Kindly go ahead with {{otp_words}}', 'CAIpack', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CBase - OTP Digits', 'CBase: Kindly go ahead with [OTP]', 'CBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CBase - OTP Words', 'CBase: Kindly go ahead with {{otp_words}}', 'CBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CCloud - OTP Digits', 'CCloud: Kindly go ahead with [OTP]', 'CCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CCloud - OTP Words', 'CCloud: Kindly go ahead with {{otp_words}}', 'CCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CCore - OTP Digits', 'CCore: Kindly go ahead with [OTP]', 'CCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CCore - OTP Words', 'CCore: Kindly go ahead with {{otp_words}}', 'CCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CDrive - OTP Digits', 'CDrive: Kindly go ahead with [OTP]', 'CDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CDrive - OTP Words', 'CDrive: Kindly go ahead with {{otp_words}}', 'CDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CEdge - OTP Digits', 'CEdge: Kindly go ahead with [OTP]', 'CEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CEdge - OTP Words', 'CEdge: Kindly go ahead with {{otp_words}}', 'CEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CFlow - OTP Digits', 'CFlow: Kindly go ahead with [OTP]', 'CFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CFlow - OTP Words', 'CFlow: Kindly go ahead with {{otp_words}}', 'CFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CGo - OTP Digits', 'CGo: Kindly go ahead with [OTP]', 'CGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CGo - OTP Words', 'CGo: Kindly go ahead with {{otp_words}}', 'CGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CHQ - OTP Digits', 'CHQ: Kindly go ahead with [OTP]', 'CHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CHQ - OTP Words', 'CHQ: Kindly go ahead with {{otp_words}}', 'CHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CHub - OTP Digits', 'CHub: Kindly go ahead with [OTP]', 'CHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CHub - OTP Words', 'CHub: Kindly go ahead with {{otp_words}}', 'CHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CLabs ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CLabs', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CLabs - OTP Digits', 'CLabs: Kindly go ahead with [OTP]', 'CLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CLabs - OTP Words', 'CLabs: Kindly go ahead with {{otp_words}}', 'CLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CLink - OTP Digits', 'CLink: Kindly go ahead with [OTP]', 'CLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CLink - OTP Words', 'CLink: Kindly go ahead with {{otp_words}}', 'CLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudAI - OTP Digits', 'CloudAI: Kindly go ahead with [OTP]', 'CloudAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudAI - OTP Words', 'CloudAI: Kindly go ahead with {{otp_words}}', 'CloudAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudBase - OTP Digits', 'CloudBase: Kindly go ahead with [OTP]', 'CloudBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudBase - OTP Words', 'CloudBase: Kindly go ahead with {{otp_words}}', 'CloudBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudCore - OTP Digits', 'CloudCore: Kindly go ahead with [OTP]', 'CloudCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudCore - OTP Words', 'CloudCore: Kindly go ahead with {{otp_words}}', 'CloudCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudDrive - OTP Digits', 'CloudDrive: Kindly go ahead with [OTP]', 'CloudDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudDrive - OTP Words', 'CloudDrive: Kindly go ahead with {{otp_words}}', 'CloudDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudEdge - OTP Digits', 'CloudEdge: Kindly go ahead with [OTP]', 'CloudEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudEdge - OTP Words', 'CloudEdge: Kindly go ahead with {{otp_words}}', 'CloudEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudFlow - OTP Digits', 'CloudFlow: Kindly go ahead with [OTP]', 'CloudFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudFlow - OTP Words', 'CloudFlow: Kindly go ahead with {{otp_words}}', 'CloudFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudGo - OTP Digits', 'CloudGo: Kindly go ahead with [OTP]', 'CloudGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudGo - OTP Words', 'CloudGo: Kindly go ahead with {{otp_words}}', 'CloudGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudHQ - OTP Digits', 'CloudHQ: Kindly go ahead with [OTP]', 'CloudHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudHQ - OTP Words', 'CloudHQ: Kindly go ahead with {{otp_words}}', 'CloudHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudHub - OTP Digits', 'CloudHub: Kindly go ahead with [OTP]', 'CloudHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudHub - OTP Words', 'CloudHub: Kindly go ahead with {{otp_words}}', 'CloudHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudLab ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudLab', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudLab - OTP Digits', 'CloudLab: Kindly go ahead with [OTP]', 'CloudLab', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudLab - OTP Words', 'CloudLab: Kindly go ahead with {{otp_words}}', 'CloudLab', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudLink - OTP Digits', 'CloudLink: Kindly go ahead with [OTP]', 'CloudLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudLink - OTP Words', 'CloudLink: Kindly go ahead with {{otp_words}}', 'CloudLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudNet - OTP Digits', 'CloudNet: Kindly go ahead with [OTP]', 'CloudNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudNet - OTP Words', 'CloudNet: Kindly go ahead with {{otp_words}}', 'CloudNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudNode ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudNode', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudNode - OTP Digits', 'CloudNode: Kindly go ahead with [OTP]', 'CloudNode', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudNode - OTP Words', 'CloudNode: Kindly go ahead with {{otp_words}}', 'CloudNode', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudOne - OTP Digits', 'CloudOne: Kindly go ahead with [OTP]', 'CloudOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudOne - OTP Words', 'CloudOne: Kindly go ahead with {{otp_words}}', 'CloudOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudOps - OTP Digits', 'CloudOps: Kindly go ahead with [OTP]', 'CloudOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudOps - OTP Words', 'CloudOps: Kindly go ahead with {{otp_words}}', 'CloudOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudPro - OTP Digits', 'CloudPro: Kindly go ahead with [OTP]', 'CloudPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudPro - OTP Words', 'CloudPro: Kindly go ahead with {{otp_words}}', 'CloudPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudStack ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudStack', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudStack - OTP Digits', 'CloudStack: Kindly go ahead with [OTP]', 'CloudStack', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudStack - OTP Words', 'CloudStack: Kindly go ahead with {{otp_words}}', 'CloudStack', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudSys - OTP Digits', 'CloudSys: Kindly go ahead with [OTP]', 'CloudSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudSys - OTP Words', 'CloudSys: Kindly go ahead with {{otp_words}}', 'CloudSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudX ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudX', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudX - OTP Digits', 'CloudX: Kindly go ahead with [OTP]', 'CloudX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudX - OTP Words', 'CloudX: Kindly go ahead with {{otp_words}}', 'CloudX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CloudZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CloudZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudZone - OTP Digits', 'CloudZone: Kindly go ahead with [OTP]', 'CloudZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CloudZone - OTP Words', 'CloudZone: Kindly go ahead with {{otp_words}}', 'CloudZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CNet - OTP Digits', 'CNet: Kindly go ahead with [OTP]', 'CNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CNet - OTP Words', 'CNet: Kindly go ahead with {{otp_words}}', 'CNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === COne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('COne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: COne - OTP Digits', 'COne: Kindly go ahead with [OTP]', 'COne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: COne - OTP Words', 'COne: Kindly go ahead with {{otp_words}}', 'COne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === COps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('COps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: COps - OTP Digits', 'COps: Kindly go ahead with [OTP]', 'COps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: COps - OTP Words', 'COps: Kindly go ahead with {{otp_words}}', 'COps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoAI - OTP Digits', 'CosmoAI: Kindly go ahead with [OTP]', 'CosmoAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoAI - OTP Words', 'CosmoAI: Kindly go ahead with {{otp_words}}', 'CosmoAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoBase - OTP Digits', 'CosmoBase: Kindly go ahead with [OTP]', 'CosmoBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoBase - OTP Words', 'CosmoBase: Kindly go ahead with {{otp_words}}', 'CosmoBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoCloud - OTP Digits', 'CosmoCloud: Kindly go ahead with [OTP]', 'CosmoCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoCloud - OTP Words', 'CosmoCloud: Kindly go ahead with {{otp_words}}', 'CosmoCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoCore - OTP Digits', 'CosmoCore: Kindly go ahead with [OTP]', 'CosmoCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoCore - OTP Words', 'CosmoCore: Kindly go ahead with {{otp_words}}', 'CosmoCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoDrive - OTP Digits', 'CosmoDrive: Kindly go ahead with [OTP]', 'CosmoDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoDrive - OTP Words', 'CosmoDrive: Kindly go ahead with {{otp_words}}', 'CosmoDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoEdge - OTP Digits', 'CosmoEdge: Kindly go ahead with [OTP]', 'CosmoEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoEdge - OTP Words', 'CosmoEdge: Kindly go ahead with {{otp_words}}', 'CosmoEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoFlow - OTP Digits', 'CosmoFlow: Kindly go ahead with [OTP]', 'CosmoFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoFlow - OTP Words', 'CosmoFlow: Kindly go ahead with {{otp_words}}', 'CosmoFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoGo - OTP Digits', 'CosmoGo: Kindly go ahead with [OTP]', 'CosmoGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoGo - OTP Words', 'CosmoGo: Kindly go ahead with {{otp_words}}', 'CosmoGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoHQ - OTP Digits', 'CosmoHQ: Kindly go ahead with [OTP]', 'CosmoHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoHQ - OTP Words', 'CosmoHQ: Kindly go ahead with {{otp_words}}', 'CosmoHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoHub - OTP Digits', 'CosmoHub: Kindly go ahead with [OTP]', 'CosmoHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoHub - OTP Words', 'CosmoHub: Kindly go ahead with {{otp_words}}', 'CosmoHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoLabs ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoLabs', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoLabs - OTP Digits', 'CosmoLabs: Kindly go ahead with [OTP]', 'CosmoLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoLabs - OTP Words', 'CosmoLabs: Kindly go ahead with {{otp_words}}', 'CosmoLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoLink - OTP Digits', 'CosmoLink: Kindly go ahead with [OTP]', 'CosmoLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoLink - OTP Words', 'CosmoLink: Kindly go ahead with {{otp_words}}', 'CosmoLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoNet - OTP Digits', 'CosmoNet: Kindly go ahead with [OTP]', 'CosmoNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoNet - OTP Words', 'CosmoNet: Kindly go ahead with {{otp_words}}', 'CosmoNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoOne - OTP Digits', 'CosmoOne: Kindly go ahead with [OTP]', 'CosmoOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoOne - OTP Words', 'CosmoOne: Kindly go ahead with {{otp_words}}', 'CosmoOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoOps - OTP Digits', 'CosmoOps: Kindly go ahead with [OTP]', 'CosmoOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoOps - OTP Words', 'CosmoOps: Kindly go ahead with {{otp_words}}', 'CosmoOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoPro - OTP Digits', 'CosmoPro: Kindly go ahead with [OTP]', 'CosmoPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoPro - OTP Words', 'CosmoPro: Kindly go ahead with {{otp_words}}', 'CosmoPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoSys - OTP Digits', 'CosmoSys: Kindly go ahead with [OTP]', 'CosmoSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoSys - OTP Words', 'CosmoSys: Kindly go ahead with {{otp_words}}', 'CosmoSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoTech - OTP Digits', 'CosmoTech: Kindly go ahead with [OTP]', 'CosmoTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoTech - OTP Words', 'CosmoTech: Kindly go ahead with {{otp_words}}', 'CosmoTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoX ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoX', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoX - OTP Digits', 'CosmoX: Kindly go ahead with [OTP]', 'CosmoX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoX - OTP Words', 'CosmoX: Kindly go ahead with {{otp_words}}', 'CosmoX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CosmoZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CosmoZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoZone - OTP Digits', 'CosmoZone: Kindly go ahead with [OTP]', 'CosmoZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CosmoZone - OTP Words', 'CosmoZone: Kindly go ahead with {{otp_words}}', 'CosmoZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CPro - OTP Digits', 'CPro: Kindly go ahead with [OTP]', 'CPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CPro - OTP Words', 'CPro: Kindly go ahead with {{otp_words}}', 'CPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CSys - OTP Digits', 'CSys: Kindly go ahead with [OTP]', 'CSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CSys - OTP Words', 'CSys: Kindly go ahead with {{otp_words}}', 'CSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CTech - OTP Digits', 'CTech: Kindly go ahead with [OTP]', 'CTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CTech - OTP Words', 'CTech: Kindly go ahead with {{otp_words}}', 'CTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === CZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('CZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CZone - OTP Digits', 'CZone: Kindly go ahead with [OTP]', 'CZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: CZone - OTP Words', 'CZone: Kindly go ahead with {{otp_words}}', 'CZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraBase - OTP Digits', 'InfraBase: Kindly go ahead with [OTP]', 'InfraBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraBase - OTP Words', 'InfraBase: Kindly go ahead with {{otp_words}}', 'InfraBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraCloud - OTP Digits', 'InfraCloud: Kindly go ahead with [OTP]', 'InfraCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraCloud - OTP Words', 'InfraCloud: Kindly go ahead with {{otp_words}}', 'InfraCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraCore - OTP Digits', 'InfraCore: Kindly go ahead with [OTP]', 'InfraCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraCore - OTP Words', 'InfraCore: Kindly go ahead with {{otp_words}}', 'InfraCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraDrive - OTP Digits', 'InfraDrive: Kindly go ahead with [OTP]', 'InfraDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraDrive - OTP Words', 'InfraDrive: Kindly go ahead with {{otp_words}}', 'InfraDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraEdge - OTP Digits', 'InfraEdge: Kindly go ahead with [OTP]', 'InfraEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraEdge - OTP Words', 'InfraEdge: Kindly go ahead with {{otp_words}}', 'InfraEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraFlow - OTP Digits', 'InfraFlow: Kindly go ahead with [OTP]', 'InfraFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraFlow - OTP Words', 'InfraFlow: Kindly go ahead with {{otp_words}}', 'InfraFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraGo - OTP Digits', 'InfraGo: Kindly go ahead with [OTP]', 'InfraGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraGo - OTP Words', 'InfraGo: Kindly go ahead with {{otp_words}}', 'InfraGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraHQ - OTP Digits', 'InfraHQ: Kindly go ahead with [OTP]', 'InfraHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraHQ - OTP Words', 'InfraHQ: Kindly go ahead with {{otp_words}}', 'InfraHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraHub - OTP Digits', 'InfraHub: Kindly go ahead with [OTP]', 'InfraHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraHub - OTP Words', 'InfraHub: Kindly go ahead with {{otp_words}}', 'InfraHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraLab ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraLab', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraLab - OTP Digits', 'InfraLab: Kindly go ahead with [OTP]', 'InfraLab', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraLab - OTP Words', 'InfraLab: Kindly go ahead with {{otp_words}}', 'InfraLab', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraLink - OTP Digits', 'InfraLink: Kindly go ahead with [OTP]', 'InfraLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraLink - OTP Words', 'InfraLink: Kindly go ahead with {{otp_words}}', 'InfraLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraNet - OTP Digits', 'InfraNet: Kindly go ahead with [OTP]', 'InfraNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraNet - OTP Words', 'InfraNet: Kindly go ahead with {{otp_words}}', 'InfraNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraNode ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraNode', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraNode - OTP Digits', 'InfraNode: Kindly go ahead with [OTP]', 'InfraNode', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraNode - OTP Words', 'InfraNode: Kindly go ahead with {{otp_words}}', 'InfraNode', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraOne - OTP Digits', 'InfraOne: Kindly go ahead with [OTP]', 'InfraOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraOne - OTP Words', 'InfraOne: Kindly go ahead with {{otp_words}}', 'InfraOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraOps - OTP Digits', 'InfraOps: Kindly go ahead with [OTP]', 'InfraOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraOps - OTP Words', 'InfraOps: Kindly go ahead with {{otp_words}}', 'InfraOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraPro - OTP Digits', 'InfraPro: Kindly go ahead with [OTP]', 'InfraPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraPro - OTP Words', 'InfraPro: Kindly go ahead with {{otp_words}}', 'InfraPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraStack ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraStack', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraStack - OTP Digits', 'InfraStack: Kindly go ahead with [OTP]', 'InfraStack', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraStack - OTP Words', 'InfraStack: Kindly go ahead with {{otp_words}}', 'InfraStack', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraSys - OTP Digits', 'InfraSys: Kindly go ahead with [OTP]', 'InfraSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraSys - OTP Words', 'InfraSys: Kindly go ahead with {{otp_words}}', 'InfraSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraX ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraX', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraX - OTP Digits', 'InfraX: Kindly go ahead with [OTP]', 'InfraX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraX - OTP Words', 'InfraX: Kindly go ahead with {{otp_words}}', 'InfraX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === InfraZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('InfraZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraZone - OTP Digits', 'InfraZone: Kindly go ahead with [OTP]', 'InfraZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: InfraZone - OTP Words', 'InfraZone: Kindly go ahead with {{otp_words}}', 'InfraZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NAI - OTP Digits', 'NAI: Kindly go ahead with [OTP]', 'NAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NAI - OTP Words', 'NAI: Kindly go ahead with {{otp_words}}', 'NAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NBase - OTP Digits', 'NBase: Kindly go ahead with [OTP]', 'NBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NBase - OTP Words', 'NBase: Kindly go ahead with {{otp_words}}', 'NBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NCloud - OTP Digits', 'NCloud: Kindly go ahead with [OTP]', 'NCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NCloud - OTP Words', 'NCloud: Kindly go ahead with {{otp_words}}', 'NCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NCore - OTP Digits', 'NCore: Kindly go ahead with [OTP]', 'NCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NCore - OTP Words', 'NCore: Kindly go ahead with {{otp_words}}', 'NCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NDrive - OTP Digits', 'NDrive: Kindly go ahead with [OTP]', 'NDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NDrive - OTP Words', 'NDrive: Kindly go ahead with {{otp_words}}', 'NDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NEdge - OTP Digits', 'NEdge: Kindly go ahead with [OTP]', 'NEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NEdge - OTP Words', 'NEdge: Kindly go ahead with {{otp_words}}', 'NEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraAI - OTP Digits', 'NeuraAI: Kindly go ahead with [OTP]', 'NeuraAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraAI - OTP Words', 'NeuraAI: Kindly go ahead with {{otp_words}}', 'NeuraAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraBase - OTP Digits', 'NeuraBase: Kindly go ahead with [OTP]', 'NeuraBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraBase - OTP Words', 'NeuraBase: Kindly go ahead with {{otp_words}}', 'NeuraBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraCloud - OTP Digits', 'NeuraCloud: Kindly go ahead with [OTP]', 'NeuraCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraCloud - OTP Words', 'NeuraCloud: Kindly go ahead with {{otp_words}}', 'NeuraCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraCore - OTP Digits', 'NeuraCore: Kindly go ahead with [OTP]', 'NeuraCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraCore - OTP Words', 'NeuraCore: Kindly go ahead with {{otp_words}}', 'NeuraCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraDrive - OTP Digits', 'NeuraDrive: Kindly go ahead with [OTP]', 'NeuraDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraDrive - OTP Words', 'NeuraDrive: Kindly go ahead with {{otp_words}}', 'NeuraDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraEdge - OTP Digits', 'NeuraEdge: Kindly go ahead with [OTP]', 'NeuraEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraEdge - OTP Words', 'NeuraEdge: Kindly go ahead with {{otp_words}}', 'NeuraEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraFlow - OTP Digits', 'NeuraFlow: Kindly go ahead with [OTP]', 'NeuraFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraFlow - OTP Words', 'NeuraFlow: Kindly go ahead with {{otp_words}}', 'NeuraFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraGo - OTP Digits', 'NeuraGo: Kindly go ahead with [OTP]', 'NeuraGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraGo - OTP Words', 'NeuraGo: Kindly go ahead with {{otp_words}}', 'NeuraGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraHQ - OTP Digits', 'NeuraHQ: Kindly go ahead with [OTP]', 'NeuraHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraHQ - OTP Words', 'NeuraHQ: Kindly go ahead with {{otp_words}}', 'NeuraHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraHub - OTP Digits', 'NeuraHub: Kindly go ahead with [OTP]', 'NeuraHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraHub - OTP Words', 'NeuraHub: Kindly go ahead with {{otp_words}}', 'NeuraHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraLabs ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraLabs', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraLabs - OTP Digits', 'NeuraLabs: Kindly go ahead with [OTP]', 'NeuraLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraLabs - OTP Words', 'NeuraLabs: Kindly go ahead with {{otp_words}}', 'NeuraLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraLink - OTP Digits', 'NeuraLink: Kindly go ahead with [OTP]', 'NeuraLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraLink - OTP Words', 'NeuraLink: Kindly go ahead with {{otp_words}}', 'NeuraLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraNet - OTP Digits', 'NeuraNet: Kindly go ahead with [OTP]', 'NeuraNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraNet - OTP Words', 'NeuraNet: Kindly go ahead with {{otp_words}}', 'NeuraNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraOne - OTP Digits', 'NeuraOne: Kindly go ahead with [OTP]', 'NeuraOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraOne - OTP Words', 'NeuraOne: Kindly go ahead with {{otp_words}}', 'NeuraOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraOps - OTP Digits', 'NeuraOps: Kindly go ahead with [OTP]', 'NeuraOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraOps - OTP Words', 'NeuraOps: Kindly go ahead with {{otp_words}}', 'NeuraOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraPro - OTP Digits', 'NeuraPro: Kindly go ahead with [OTP]', 'NeuraPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraPro - OTP Words', 'NeuraPro: Kindly go ahead with {{otp_words}}', 'NeuraPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraSys - OTP Digits', 'NeuraSys: Kindly go ahead with [OTP]', 'NeuraSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraSys - OTP Words', 'NeuraSys: Kindly go ahead with {{otp_words}}', 'NeuraSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraTech - OTP Digits', 'NeuraTech: Kindly go ahead with [OTP]', 'NeuraTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraTech - OTP Words', 'NeuraTech: Kindly go ahead with {{otp_words}}', 'NeuraTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraX ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraX', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraX - OTP Digits', 'NeuraX: Kindly go ahead with [OTP]', 'NeuraX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraX - OTP Words', 'NeuraX: Kindly go ahead with {{otp_words}}', 'NeuraX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NeuraZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NeuraZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraZone - OTP Digits', 'NeuraZone: Kindly go ahead with [OTP]', 'NeuraZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NeuraZone - OTP Words', 'NeuraZone: Kindly go ahead with {{otp_words}}', 'NeuraZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NFlow - OTP Digits', 'NFlow: Kindly go ahead with [OTP]', 'NFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NFlow - OTP Words', 'NFlow: Kindly go ahead with {{otp_words}}', 'NFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NGo - OTP Digits', 'NGo: Kindly go ahead with [OTP]', 'NGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NGo - OTP Words', 'NGo: Kindly go ahead with {{otp_words}}', 'NGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NHQ - OTP Digits', 'NHQ: Kindly go ahead with [OTP]', 'NHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NHQ - OTP Words', 'NHQ: Kindly go ahead with {{otp_words}}', 'NHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NHub - OTP Digits', 'NHub: Kindly go ahead with [OTP]', 'NHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NHub - OTP Words', 'NHub: Kindly go ahead with {{otp_words}}', 'NHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NLabs ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NLabs', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NLabs - OTP Digits', 'NLabs: Kindly go ahead with [OTP]', 'NLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NLabs - OTP Words', 'NLabs: Kindly go ahead with {{otp_words}}', 'NLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NLink - OTP Digits', 'NLink: Kindly go ahead with [OTP]', 'NLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NLink - OTP Words', 'NLink: Kindly go ahead with {{otp_words}}', 'NLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NNet - OTP Digits', 'NNet: Kindly go ahead with [OTP]', 'NNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NNet - OTP Words', 'NNet: Kindly go ahead with {{otp_words}}', 'NNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NOne - OTP Digits', 'NOne: Kindly go ahead with [OTP]', 'NOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NOne - OTP Words', 'NOne: Kindly go ahead with {{otp_words}}', 'NOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NOps - OTP Digits', 'NOps: Kindly go ahead with [OTP]', 'NOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NOps - OTP Words', 'NOps: Kindly go ahead with {{otp_words}}', 'NOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaAI - OTP Digits', 'NovaAI: Kindly go ahead with [OTP]', 'NovaAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaAI - OTP Words', 'NovaAI: Kindly go ahead with {{otp_words}}', 'NovaAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaBase - OTP Digits', 'NovaBase: Kindly go ahead with [OTP]', 'NovaBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaBase - OTP Words', 'NovaBase: Kindly go ahead with {{otp_words}}', 'NovaBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaCloud - OTP Digits', 'NovaCloud: Kindly go ahead with [OTP]', 'NovaCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaCloud - OTP Words', 'NovaCloud: Kindly go ahead with {{otp_words}}', 'NovaCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaCore - OTP Digits', 'NovaCore: Kindly go ahead with [OTP]', 'NovaCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaCore - OTP Words', 'NovaCore: Kindly go ahead with {{otp_words}}', 'NovaCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaDrive - OTP Digits', 'NovaDrive: Kindly go ahead with [OTP]', 'NovaDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaDrive - OTP Words', 'NovaDrive: Kindly go ahead with {{otp_words}}', 'NovaDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaEdge - OTP Digits', 'NovaEdge: Kindly go ahead with [OTP]', 'NovaEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaEdge - OTP Words', 'NovaEdge: Kindly go ahead with {{otp_words}}', 'NovaEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaFlow - OTP Digits', 'NovaFlow: Kindly go ahead with [OTP]', 'NovaFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaFlow - OTP Words', 'NovaFlow: Kindly go ahead with {{otp_words}}', 'NovaFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaGo - OTP Digits', 'NovaGo: Kindly go ahead with [OTP]', 'NovaGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaGo - OTP Words', 'NovaGo: Kindly go ahead with {{otp_words}}', 'NovaGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaHQ - OTP Digits', 'NovaHQ: Kindly go ahead with [OTP]', 'NovaHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaHQ - OTP Words', 'NovaHQ: Kindly go ahead with {{otp_words}}', 'NovaHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaHub - OTP Digits', 'NovaHub: Kindly go ahead with [OTP]', 'NovaHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaHub - OTP Words', 'NovaHub: Kindly go ahead with {{otp_words}}', 'NovaHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaLabs ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaLabs', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaLabs - OTP Digits', 'NovaLabs: Kindly go ahead with [OTP]', 'NovaLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaLabs - OTP Words', 'NovaLabs: Kindly go ahead with {{otp_words}}', 'NovaLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaLink - OTP Digits', 'NovaLink: Kindly go ahead with [OTP]', 'NovaLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaLink - OTP Words', 'NovaLink: Kindly go ahead with {{otp_words}}', 'NovaLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaNet - OTP Digits', 'NovaNet: Kindly go ahead with [OTP]', 'NovaNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaNet - OTP Words', 'NovaNet: Kindly go ahead with {{otp_words}}', 'NovaNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaOne - OTP Digits', 'NovaOne: Kindly go ahead with [OTP]', 'NovaOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaOne - OTP Words', 'NovaOne: Kindly go ahead with {{otp_words}}', 'NovaOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaOps - OTP Digits', 'NovaOps: Kindly go ahead with [OTP]', 'NovaOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaOps - OTP Words', 'NovaOps: Kindly go ahead with {{otp_words}}', 'NovaOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaPro - OTP Digits', 'NovaPro: Kindly go ahead with [OTP]', 'NovaPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaPro - OTP Words', 'NovaPro: Kindly go ahead with {{otp_words}}', 'NovaPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaSys - OTP Digits', 'NovaSys: Kindly go ahead with [OTP]', 'NovaSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaSys - OTP Words', 'NovaSys: Kindly go ahead with {{otp_words}}', 'NovaSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaTech - OTP Digits', 'NovaTech: Kindly go ahead with [OTP]', 'NovaTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaTech - OTP Words', 'NovaTech: Kindly go ahead with {{otp_words}}', 'NovaTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaX ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaX', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaX - OTP Digits', 'NovaX: Kindly go ahead with [OTP]', 'NovaX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaX - OTP Words', 'NovaX: Kindly go ahead with {{otp_words}}', 'NovaX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NovaZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NovaZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaZone - OTP Digits', 'NovaZone: Kindly go ahead with [OTP]', 'NovaZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NovaZone - OTP Words', 'NovaZone: Kindly go ahead with {{otp_words}}', 'NovaZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NPro - OTP Digits', 'NPro: Kindly go ahead with [OTP]', 'NPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NPro - OTP Words', 'NPro: Kindly go ahead with {{otp_words}}', 'NPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NSys - OTP Digits', 'NSys: Kindly go ahead with [OTP]', 'NSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NSys - OTP Words', 'NSys: Kindly go ahead with {{otp_words}}', 'NSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NTech - OTP Digits', 'NTech: Kindly go ahead with [OTP]', 'NTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NTech - OTP Words', 'NTech: Kindly go ahead with {{otp_words}}', 'NTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === NZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('NZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NZone - OTP Digits', 'NZone: Kindly go ahead with [OTP]', 'NZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: NZone - OTP Words', 'NZone: Kindly go ahead with {{otp_words}}', 'NZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PAI - OTP Digits', 'PAI: Kindly go ahead with [OTP]', 'PAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PAI - OTP Words', 'PAI: Kindly go ahead with {{otp_words}}', 'PAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PBase - OTP Digits', 'PBase: Kindly go ahead with [OTP]', 'PBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PBase - OTP Words', 'PBase: Kindly go ahead with {{otp_words}}', 'PBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PCloud - OTP Digits', 'PCloud: Kindly go ahead with [OTP]', 'PCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PCloud - OTP Words', 'PCloud: Kindly go ahead with {{otp_words}}', 'PCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PCore - OTP Digits', 'PCore: Kindly go ahead with [OTP]', 'PCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PCore - OTP Words', 'PCore: Kindly go ahead with {{otp_words}}', 'PCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PDrive - OTP Digits', 'PDrive: Kindly go ahead with [OTP]', 'PDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PDrive - OTP Words', 'PDrive: Kindly go ahead with {{otp_words}}', 'PDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PEdge - OTP Digits', 'PEdge: Kindly go ahead with [OTP]', 'PEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PEdge - OTP Words', 'PEdge: Kindly go ahead with {{otp_words}}', 'PEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PFlow - OTP Digits', 'PFlow: Kindly go ahead with [OTP]', 'PFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PFlow - OTP Words', 'PFlow: Kindly go ahead with {{otp_words}}', 'PFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PGo - OTP Digits', 'PGo: Kindly go ahead with [OTP]', 'PGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PGo - OTP Words', 'PGo: Kindly go ahead with {{otp_words}}', 'PGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PHQ - OTP Digits', 'PHQ: Kindly go ahead with [OTP]', 'PHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PHQ - OTP Words', 'PHQ: Kindly go ahead with {{otp_words}}', 'PHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PHub - OTP Digits', 'PHub: Kindly go ahead with [OTP]', 'PHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PHub - OTP Words', 'PHub: Kindly go ahead with {{otp_words}}', 'PHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PLabs ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PLabs', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PLabs - OTP Digits', 'PLabs: Kindly go ahead with [OTP]', 'PLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PLabs - OTP Words', 'PLabs: Kindly go ahead with {{otp_words}}', 'PLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PLink - OTP Digits', 'PLink: Kindly go ahead with [OTP]', 'PLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PLink - OTP Words', 'PLink: Kindly go ahead with {{otp_words}}', 'PLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PNet - OTP Digits', 'PNet: Kindly go ahead with [OTP]', 'PNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PNet - OTP Words', 'PNet: Kindly go ahead with {{otp_words}}', 'PNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === POne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('POne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: POne - OTP Digits', 'POne: Kindly go ahead with [OTP]', 'POne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: POne - OTP Words', 'POne: Kindly go ahead with {{otp_words}}', 'POne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === POps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('POps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: POps - OTP Digits', 'POps: Kindly go ahead with [OTP]', 'POps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: POps - OTP Words', 'POps: Kindly go ahead with {{otp_words}}', 'POps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PPro - OTP Digits', 'PPro: Kindly go ahead with [OTP]', 'PPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PPro - OTP Words', 'PPro: Kindly go ahead with {{otp_words}}', 'PPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PSys - OTP Digits', 'PSys: Kindly go ahead with [OTP]', 'PSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PSys - OTP Words', 'PSys: Kindly go ahead with {{otp_words}}', 'PSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PTech - OTP Digits', 'PTech: Kindly go ahead with [OTP]', 'PTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PTech - OTP Words', 'PTech: Kindly go ahead with {{otp_words}}', 'PTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === PZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('PZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PZone - OTP Digits', 'PZone: Kindly go ahead with [OTP]', 'PZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: PZone - OTP Words', 'PZone: Kindly go ahead with {{otp_words}}', 'PZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QAI - OTP Digits', 'QAI: Kindly go ahead with [OTP]', 'QAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QAI - OTP Words', 'QAI: Kindly go ahead with {{otp_words}}', 'QAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QBase - OTP Digits', 'QBase: Kindly go ahead with [OTP]', 'QBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QBase - OTP Words', 'QBase: Kindly go ahead with {{otp_words}}', 'QBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QCloud - OTP Digits', 'QCloud: Kindly go ahead with [OTP]', 'QCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QCloud - OTP Words', 'QCloud: Kindly go ahead with {{otp_words}}', 'QCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QCore - OTP Digits', 'QCore: Kindly go ahead with [OTP]', 'QCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QCore - OTP Words', 'QCore: Kindly go ahead with {{otp_words}}', 'QCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QDrive - OTP Digits', 'QDrive: Kindly go ahead with [OTP]', 'QDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QDrive - OTP Words', 'QDrive: Kindly go ahead with {{otp_words}}', 'QDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QEdge - OTP Digits', 'QEdge: Kindly go ahead with [OTP]', 'QEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QEdge - OTP Words', 'QEdge: Kindly go ahead with {{otp_words}}', 'QEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QFlow - OTP Digits', 'QFlow: Kindly go ahead with [OTP]', 'QFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QFlow - OTP Words', 'QFlow: Kindly go ahead with {{otp_words}}', 'QFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QGo - OTP Digits', 'QGo: Kindly go ahead with [OTP]', 'QGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QGo - OTP Words', 'QGo: Kindly go ahead with {{otp_words}}', 'QGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QHQ - OTP Digits', 'QHQ: Kindly go ahead with [OTP]', 'QHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QHQ - OTP Words', 'QHQ: Kindly go ahead with {{otp_words}}', 'QHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QHub - OTP Digits', 'QHub: Kindly go ahead with [OTP]', 'QHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QHub - OTP Words', 'QHub: Kindly go ahead with {{otp_words}}', 'QHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QLabs ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QLabs', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QLabs - OTP Digits', 'QLabs: Kindly go ahead with [OTP]', 'QLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QLabs - OTP Words', 'QLabs: Kindly go ahead with {{otp_words}}', 'QLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QLink - OTP Digits', 'QLink: Kindly go ahead with [OTP]', 'QLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QLink - OTP Words', 'QLink: Kindly go ahead with {{otp_words}}', 'QLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QNet - OTP Digits', 'QNet: Kindly go ahead with [OTP]', 'QNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QNet - OTP Words', 'QNet: Kindly go ahead with {{otp_words}}', 'QNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QOne - OTP Digits', 'QOne: Kindly go ahead with [OTP]', 'QOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QOne - OTP Words', 'QOne: Kindly go ahead with {{otp_words}}', 'QOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QOps - OTP Digits', 'QOps: Kindly go ahead with [OTP]', 'QOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QOps - OTP Words', 'QOps: Kindly go ahead with {{otp_words}}', 'QOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QPro - OTP Digits', 'QPro: Kindly go ahead with [OTP]', 'QPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QPro - OTP Words', 'QPro: Kindly go ahead with {{otp_words}}', 'QPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QSys - OTP Digits', 'QSys: Kindly go ahead with [OTP]', 'QSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QSys - OTP Words', 'QSys: Kindly go ahead with {{otp_words}}', 'QSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QTech - OTP Digits', 'QTech: Kindly go ahead with [OTP]', 'QTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QTech - OTP Words', 'QTech: Kindly go ahead with {{otp_words}}', 'QTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === QZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('QZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QZone - OTP Digits', 'QZone: Kindly go ahead with [OTP]', 'QZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: QZone - OTP Words', 'QZone: Kindly go ahead with {{otp_words}}', 'QZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RAI - OTP Digits', 'RAI: Kindly go ahead with [OTP]', 'RAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RAI - OTP Words', 'RAI: Kindly go ahead with {{otp_words}}', 'RAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RBase - OTP Digits', 'RBase: Kindly go ahead with [OTP]', 'RBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RBase - OTP Words', 'RBase: Kindly go ahead with {{otp_words}}', 'RBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RCloud - OTP Digits', 'RCloud: Kindly go ahead with [OTP]', 'RCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RCloud - OTP Words', 'RCloud: Kindly go ahead with {{otp_words}}', 'RCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RCore - OTP Digits', 'RCore: Kindly go ahead with [OTP]', 'RCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RCore - OTP Words', 'RCore: Kindly go ahead with {{otp_words}}', 'RCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RDrive - OTP Digits', 'RDrive: Kindly go ahead with [OTP]', 'RDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RDrive - OTP Words', 'RDrive: Kindly go ahead with {{otp_words}}', 'RDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === REdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('REdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: REdge - OTP Digits', 'REdge: Kindly go ahead with [OTP]', 'REdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: REdge - OTP Words', 'REdge: Kindly go ahead with {{otp_words}}', 'REdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RFlow - OTP Digits', 'RFlow: Kindly go ahead with [OTP]', 'RFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RFlow - OTP Words', 'RFlow: Kindly go ahead with {{otp_words}}', 'RFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RGo - OTP Digits', 'RGo: Kindly go ahead with [OTP]', 'RGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RGo - OTP Words', 'RGo: Kindly go ahead with {{otp_words}}', 'RGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RHQ - OTP Digits', 'RHQ: Kindly go ahead with [OTP]', 'RHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RHQ - OTP Words', 'RHQ: Kindly go ahead with {{otp_words}}', 'RHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RHub - OTP Digits', 'RHub: Kindly go ahead with [OTP]', 'RHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RHub - OTP Words', 'RHub: Kindly go ahead with {{otp_words}}', 'RHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RLabs ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RLabs', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RLabs - OTP Digits', 'RLabs: Kindly go ahead with [OTP]', 'RLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RLabs - OTP Words', 'RLabs: Kindly go ahead with {{otp_words}}', 'RLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RLink - OTP Digits', 'RLink: Kindly go ahead with [OTP]', 'RLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RLink - OTP Words', 'RLink: Kindly go ahead with {{otp_words}}', 'RLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RNet - OTP Digits', 'RNet: Kindly go ahead with [OTP]', 'RNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RNet - OTP Words', 'RNet: Kindly go ahead with {{otp_words}}', 'RNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ROne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ROne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ROne - OTP Digits', 'ROne: Kindly go ahead with [OTP]', 'ROne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ROne - OTP Words', 'ROne: Kindly go ahead with {{otp_words}}', 'ROne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ROps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ROps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ROps - OTP Digits', 'ROps: Kindly go ahead with [OTP]', 'ROps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ROps - OTP Words', 'ROps: Kindly go ahead with {{otp_words}}', 'ROps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RPro - OTP Digits', 'RPro: Kindly go ahead with [OTP]', 'RPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RPro - OTP Words', 'RPro: Kindly go ahead with {{otp_words}}', 'RPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RSys - OTP Digits', 'RSys: Kindly go ahead with [OTP]', 'RSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RSys - OTP Words', 'RSys: Kindly go ahead with {{otp_words}}', 'RSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RTech - OTP Digits', 'RTech: Kindly go ahead with [OTP]', 'RTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RTech - OTP Words', 'RTech: Kindly go ahead with {{otp_words}}', 'RTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === RZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('RZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RZone - OTP Digits', 'RZone: Kindly go ahead with [OTP]', 'RZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: RZone - OTP Words', 'RZone: Kindly go ahead with {{otp_words}}', 'RZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === sender_id ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('sender_id', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: sender_id - OTP Digits', 'sender_id: Kindly go ahead with [OTP]', 'sender_id', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: sender_id - OTP Words', 'sender_id: Kindly go ahead with {{otp_words}}', 'sender_id', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TAI - OTP Digits', 'TAI: Kindly go ahead with [OTP]', 'TAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TAI - OTP Words', 'TAI: Kindly go ahead with {{otp_words}}', 'TAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TBase - OTP Digits', 'TBase: Kindly go ahead with [OTP]', 'TBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TBase - OTP Words', 'TBase: Kindly go ahead with {{otp_words}}', 'TBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TCloud - OTP Digits', 'TCloud: Kindly go ahead with [OTP]', 'TCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TCloud - OTP Words', 'TCloud: Kindly go ahead with {{otp_words}}', 'TCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TCore - OTP Digits', 'TCore: Kindly go ahead with [OTP]', 'TCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TCore - OTP Words', 'TCore: Kindly go ahead with {{otp_words}}', 'TCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TDrive - OTP Digits', 'TDrive: Kindly go ahead with [OTP]', 'TDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TDrive - OTP Words', 'TDrive: Kindly go ahead with {{otp_words}}', 'TDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TEdge - OTP Digits', 'TEdge: Kindly go ahead with [OTP]', 'TEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TEdge - OTP Words', 'TEdge: Kindly go ahead with {{otp_words}}', 'TEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TFlow - OTP Digits', 'TFlow: Kindly go ahead with [OTP]', 'TFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TFlow - OTP Words', 'TFlow: Kindly go ahead with {{otp_words}}', 'TFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TGo - OTP Digits', 'TGo: Kindly go ahead with [OTP]', 'TGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TGo - OTP Words', 'TGo: Kindly go ahead with {{otp_words}}', 'TGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === THQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('THQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: THQ - OTP Digits', 'THQ: Kindly go ahead with [OTP]', 'THQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: THQ - OTP Words', 'THQ: Kindly go ahead with {{otp_words}}', 'THQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === THub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('THub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: THub - OTP Digits', 'THub: Kindly go ahead with [OTP]', 'THub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: THub - OTP Words', 'THub: Kindly go ahead with {{otp_words}}', 'THub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TLabs ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TLabs', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TLabs - OTP Digits', 'TLabs: Kindly go ahead with [OTP]', 'TLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TLabs - OTP Words', 'TLabs: Kindly go ahead with {{otp_words}}', 'TLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TLink - OTP Digits', 'TLink: Kindly go ahead with [OTP]', 'TLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TLink - OTP Words', 'TLink: Kindly go ahead with {{otp_words}}', 'TLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TNet - OTP Digits', 'TNet: Kindly go ahead with [OTP]', 'TNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TNet - OTP Words', 'TNet: Kindly go ahead with {{otp_words}}', 'TNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TOne - OTP Digits', 'TOne: Kindly go ahead with [OTP]', 'TOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TOne - OTP Words', 'TOne: Kindly go ahead with {{otp_words}}', 'TOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TOps - OTP Digits', 'TOps: Kindly go ahead with [OTP]', 'TOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TOps - OTP Words', 'TOps: Kindly go ahead with {{otp_words}}', 'TOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TPro - OTP Digits', 'TPro: Kindly go ahead with [OTP]', 'TPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TPro - OTP Words', 'TPro: Kindly go ahead with {{otp_words}}', 'TPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TSys - OTP Digits', 'TSys: Kindly go ahead with [OTP]', 'TSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TSys - OTP Words', 'TSys: Kindly go ahead with {{otp_words}}', 'TSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TTech - OTP Digits', 'TTech: Kindly go ahead with [OTP]', 'TTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TTech - OTP Words', 'TTech: Kindly go ahead with {{otp_words}}', 'TTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === TZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('TZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TZone - OTP Digits', 'TZone: Kindly go ahead with [OTP]', 'TZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: TZone - OTP Words', 'TZone: Kindly go ahead with {{otp_words}}', 'TZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VAI - OTP Digits', 'VAI: Kindly go ahead with [OTP]', 'VAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VAI - OTP Words', 'VAI: Kindly go ahead with {{otp_words}}', 'VAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VBase - OTP Digits', 'VBase: Kindly go ahead with [OTP]', 'VBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VBase - OTP Words', 'VBase: Kindly go ahead with {{otp_words}}', 'VBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VCloud - OTP Digits', 'VCloud: Kindly go ahead with [OTP]', 'VCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VCloud - OTP Words', 'VCloud: Kindly go ahead with {{otp_words}}', 'VCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VCore - OTP Digits', 'VCore: Kindly go ahead with [OTP]', 'VCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VCore - OTP Words', 'VCore: Kindly go ahead with {{otp_words}}', 'VCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VDrive - OTP Digits', 'VDrive: Kindly go ahead with [OTP]', 'VDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VDrive - OTP Words', 'VDrive: Kindly go ahead with {{otp_words}}', 'VDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VEdge - OTP Digits', 'VEdge: Kindly go ahead with [OTP]', 'VEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VEdge - OTP Words', 'VEdge: Kindly go ahead with {{otp_words}}', 'VEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VFlow - OTP Digits', 'VFlow: Kindly go ahead with [OTP]', 'VFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VFlow - OTP Words', 'VFlow: Kindly go ahead with {{otp_words}}', 'VFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VGo - OTP Digits', 'VGo: Kindly go ahead with [OTP]', 'VGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VGo - OTP Words', 'VGo: Kindly go ahead with {{otp_words}}', 'VGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VHQ - OTP Digits', 'VHQ: Kindly go ahead with [OTP]', 'VHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VHQ - OTP Words', 'VHQ: Kindly go ahead with {{otp_words}}', 'VHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VHub - OTP Digits', 'VHub: Kindly go ahead with [OTP]', 'VHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VHub - OTP Words', 'VHub: Kindly go ahead with {{otp_words}}', 'VHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VLabs ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VLabs', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VLabs - OTP Digits', 'VLabs: Kindly go ahead with [OTP]', 'VLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VLabs - OTP Words', 'VLabs: Kindly go ahead with {{otp_words}}', 'VLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VLink - OTP Digits', 'VLink: Kindly go ahead with [OTP]', 'VLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VLink - OTP Words', 'VLink: Kindly go ahead with {{otp_words}}', 'VLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VNet - OTP Digits', 'VNet: Kindly go ahead with [OTP]', 'VNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VNet - OTP Words', 'VNet: Kindly go ahead with {{otp_words}}', 'VNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VOne - OTP Digits', 'VOne: Kindly go ahead with [OTP]', 'VOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VOne - OTP Words', 'VOne: Kindly go ahead with {{otp_words}}', 'VOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VOps - OTP Digits', 'VOps: Kindly go ahead with [OTP]', 'VOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VOps - OTP Words', 'VOps: Kindly go ahead with {{otp_words}}', 'VOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VPro - OTP Digits', 'VPro: Kindly go ahead with [OTP]', 'VPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VPro - OTP Words', 'VPro: Kindly go ahead with {{otp_words}}', 'VPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VSys - OTP Digits', 'VSys: Kindly go ahead with [OTP]', 'VSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VSys - OTP Words', 'VSys: Kindly go ahead with {{otp_words}}', 'VSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VTech - OTP Digits', 'VTech: Kindly go ahead with [OTP]', 'VTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VTech - OTP Words', 'VTech: Kindly go ahead with {{otp_words}}', 'VTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === VZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('VZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VZone - OTP Digits', 'VZone: Kindly go ahead with [OTP]', 'VZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: VZone - OTP Words', 'VZone: Kindly go ahead with {{otp_words}}', 'VZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XAI - OTP Digits', 'XAI: Kindly go ahead with [OTP]', 'XAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XAI - OTP Words', 'XAI: Kindly go ahead with {{otp_words}}', 'XAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XBase - OTP Digits', 'XBase: Kindly go ahead with [OTP]', 'XBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XBase - OTP Words', 'XBase: Kindly go ahead with {{otp_words}}', 'XBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XCloud - OTP Digits', 'XCloud: Kindly go ahead with [OTP]', 'XCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XCloud - OTP Words', 'XCloud: Kindly go ahead with {{otp_words}}', 'XCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XCore - OTP Digits', 'XCore: Kindly go ahead with [OTP]', 'XCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XCore - OTP Words', 'XCore: Kindly go ahead with {{otp_words}}', 'XCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XDrive - OTP Digits', 'XDrive: Kindly go ahead with [OTP]', 'XDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XDrive - OTP Words', 'XDrive: Kindly go ahead with {{otp_words}}', 'XDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XEdge - OTP Digits', 'XEdge: Kindly go ahead with [OTP]', 'XEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XEdge - OTP Words', 'XEdge: Kindly go ahead with {{otp_words}}', 'XEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XFlow - OTP Digits', 'XFlow: Kindly go ahead with [OTP]', 'XFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XFlow - OTP Words', 'XFlow: Kindly go ahead with {{otp_words}}', 'XFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XGo - OTP Digits', 'XGo: Kindly go ahead with [OTP]', 'XGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XGo - OTP Words', 'XGo: Kindly go ahead with {{otp_words}}', 'XGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XHQ - OTP Digits', 'XHQ: Kindly go ahead with [OTP]', 'XHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XHQ - OTP Words', 'XHQ: Kindly go ahead with {{otp_words}}', 'XHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XHub - OTP Digits', 'XHub: Kindly go ahead with [OTP]', 'XHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XHub - OTP Words', 'XHub: Kindly go ahead with {{otp_words}}', 'XHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XLabs ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XLabs', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XLabs - OTP Digits', 'XLabs: Kindly go ahead with [OTP]', 'XLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XLabs - OTP Words', 'XLabs: Kindly go ahead with {{otp_words}}', 'XLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XLink - OTP Digits', 'XLink: Kindly go ahead with [OTP]', 'XLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XLink - OTP Words', 'XLink: Kindly go ahead with {{otp_words}}', 'XLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XNet - OTP Digits', 'XNet: Kindly go ahead with [OTP]', 'XNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XNet - OTP Words', 'XNet: Kindly go ahead with {{otp_words}}', 'XNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XOne - OTP Digits', 'XOne: Kindly go ahead with [OTP]', 'XOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XOne - OTP Words', 'XOne: Kindly go ahead with {{otp_words}}', 'XOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XOps - OTP Digits', 'XOps: Kindly go ahead with [OTP]', 'XOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XOps - OTP Words', 'XOps: Kindly go ahead with {{otp_words}}', 'XOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XPro - OTP Digits', 'XPro: Kindly go ahead with [OTP]', 'XPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XPro - OTP Words', 'XPro: Kindly go ahead with {{otp_words}}', 'XPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XSys - OTP Digits', 'XSys: Kindly go ahead with [OTP]', 'XSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XSys - OTP Words', 'XSys: Kindly go ahead with {{otp_words}}', 'XSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XTech - OTP Digits', 'XTech: Kindly go ahead with [OTP]', 'XTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XTech - OTP Words', 'XTech: Kindly go ahead with {{otp_words}}', 'XTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === XZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('XZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XZone - OTP Digits', 'XZone: Kindly go ahead with [OTP]', 'XZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: XZone - OTP Words', 'XZone: Kindly go ahead with {{otp_words}}', 'XZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZAI - OTP Digits', 'ZAI: Kindly go ahead with [OTP]', 'ZAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZAI - OTP Words', 'ZAI: Kindly go ahead with {{otp_words}}', 'ZAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZBase ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZBase', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZBase - OTP Digits', 'ZBase: Kindly go ahead with [OTP]', 'ZBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZBase - OTP Words', 'ZBase: Kindly go ahead with {{otp_words}}', 'ZBase', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZCloud - OTP Digits', 'ZCloud: Kindly go ahead with [OTP]', 'ZCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZCloud - OTP Words', 'ZCloud: Kindly go ahead with {{otp_words}}', 'ZCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZCore - OTP Digits', 'ZCore: Kindly go ahead with [OTP]', 'ZCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZCore - OTP Words', 'ZCore: Kindly go ahead with {{otp_words}}', 'ZCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZDrive - OTP Digits', 'ZDrive: Kindly go ahead with [OTP]', 'ZDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZDrive - OTP Words', 'ZDrive: Kindly go ahead with {{otp_words}}', 'ZDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZEdge - OTP Digits', 'ZEdge: Kindly go ahead with [OTP]', 'ZEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZEdge - OTP Words', 'ZEdge: Kindly go ahead with {{otp_words}}', 'ZEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZFlow - OTP Digits', 'ZFlow: Kindly go ahead with [OTP]', 'ZFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZFlow - OTP Words', 'ZFlow: Kindly go ahead with {{otp_words}}', 'ZFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZGo - OTP Digits', 'ZGo: Kindly go ahead with [OTP]', 'ZGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZGo - OTP Words', 'ZGo: Kindly go ahead with {{otp_words}}', 'ZGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZHQ - OTP Digits', 'ZHQ: Kindly go ahead with [OTP]', 'ZHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZHQ - OTP Words', 'ZHQ: Kindly go ahead with {{otp_words}}', 'ZHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZHub - OTP Digits', 'ZHub: Kindly go ahead with [OTP]', 'ZHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZHub - OTP Words', 'ZHub: Kindly go ahead with {{otp_words}}', 'ZHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZLabs ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZLabs', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZLabs - OTP Digits', 'ZLabs: Kindly go ahead with [OTP]', 'ZLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZLabs - OTP Words', 'ZLabs: Kindly go ahead with {{otp_words}}', 'ZLabs', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZLink - OTP Digits', 'ZLink: Kindly go ahead with [OTP]', 'ZLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZLink - OTP Words', 'ZLink: Kindly go ahead with {{otp_words}}', 'ZLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZNet - OTP Digits', 'ZNet: Kindly go ahead with [OTP]', 'ZNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZNet - OTP Words', 'ZNet: Kindly go ahead with {{otp_words}}', 'ZNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZOne - OTP Digits', 'ZOne: Kindly go ahead with [OTP]', 'ZOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZOne - OTP Words', 'ZOne: Kindly go ahead with {{otp_words}}', 'ZOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZOps - OTP Digits', 'ZOps: Kindly go ahead with [OTP]', 'ZOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZOps - OTP Words', 'ZOps: Kindly go ahead with {{otp_words}}', 'ZOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === Zpace ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('Zpace', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: Zpace - OTP Digits', 'Zpace: Kindly go ahead with [OTP]', 'Zpace', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: Zpace - OTP Words', 'Zpace: Kindly go ahead with {{otp_words}}', 'Zpace', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPace ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPace', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPace - OTP Digits', 'ZPace: Kindly go ahead with [OTP]', 'ZPace', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPace - OTP Words', 'ZPace: Kindly go ahead with {{otp_words}}', 'ZPace', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceAI - OTP Digits', 'ZpaceAI: Kindly go ahead with [OTP]', 'ZpaceAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceAI - OTP Words', 'ZpaceAI: Kindly go ahead with {{otp_words}}', 'ZpaceAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceAI ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceAI', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceAI - OTP Digits', 'ZPaceAI: Kindly go ahead with [OTP]', 'ZPaceAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceAI - OTP Words', 'ZPaceAI: Kindly go ahead with {{otp_words}}', 'ZPaceAI', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceCloud - OTP Digits', 'ZpaceCloud: Kindly go ahead with [OTP]', 'ZpaceCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceCloud - OTP Words', 'ZpaceCloud: Kindly go ahead with {{otp_words}}', 'ZpaceCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceCloud ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceCloud', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceCloud - OTP Digits', 'ZPaceCloud: Kindly go ahead with [OTP]', 'ZPaceCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceCloud - OTP Words', 'ZPaceCloud: Kindly go ahead with {{otp_words}}', 'ZPaceCloud', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceCore - OTP Digits', 'ZpaceCore: Kindly go ahead with [OTP]', 'ZpaceCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceCore - OTP Words', 'ZpaceCore: Kindly go ahead with {{otp_words}}', 'ZpaceCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceCore ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceCore', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceCore - OTP Digits', 'ZPaceCore: Kindly go ahead with [OTP]', 'ZPaceCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceCore - OTP Words', 'ZPaceCore: Kindly go ahead with {{otp_words}}', 'ZPaceCore', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceDrive - OTP Digits', 'ZpaceDrive: Kindly go ahead with [OTP]', 'ZpaceDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceDrive - OTP Words', 'ZpaceDrive: Kindly go ahead with {{otp_words}}', 'ZpaceDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceDrive ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceDrive', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceDrive - OTP Digits', 'ZPaceDrive: Kindly go ahead with [OTP]', 'ZPaceDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceDrive - OTP Words', 'ZPaceDrive: Kindly go ahead with {{otp_words}}', 'ZPaceDrive', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceEdge - OTP Digits', 'ZpaceEdge: Kindly go ahead with [OTP]', 'ZpaceEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceEdge - OTP Words', 'ZpaceEdge: Kindly go ahead with {{otp_words}}', 'ZpaceEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceEdge ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceEdge', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceEdge - OTP Digits', 'ZPaceEdge: Kindly go ahead with [OTP]', 'ZPaceEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceEdge - OTP Words', 'ZPaceEdge: Kindly go ahead with {{otp_words}}', 'ZPaceEdge', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceFlow - OTP Digits', 'ZpaceFlow: Kindly go ahead with [OTP]', 'ZpaceFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceFlow - OTP Words', 'ZpaceFlow: Kindly go ahead with {{otp_words}}', 'ZpaceFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceFlow ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceFlow', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceFlow - OTP Digits', 'ZPaceFlow: Kindly go ahead with [OTP]', 'ZPaceFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceFlow - OTP Words', 'ZPaceFlow: Kindly go ahead with {{otp_words}}', 'ZPaceFlow', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceGo - OTP Digits', 'ZpaceGo: Kindly go ahead with [OTP]', 'ZpaceGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceGo - OTP Words', 'ZpaceGo: Kindly go ahead with {{otp_words}}', 'ZpaceGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceGo ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceGo', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceGo - OTP Digits', 'ZPaceGo: Kindly go ahead with [OTP]', 'ZPaceGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceGo - OTP Words', 'ZPaceGo: Kindly go ahead with {{otp_words}}', 'ZPaceGo', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceHQ - OTP Digits', 'ZpaceHQ: Kindly go ahead with [OTP]', 'ZpaceHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceHQ - OTP Words', 'ZpaceHQ: Kindly go ahead with {{otp_words}}', 'ZpaceHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceHQ ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceHQ', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceHQ - OTP Digits', 'ZPaceHQ: Kindly go ahead with [OTP]', 'ZPaceHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceHQ - OTP Words', 'ZPaceHQ: Kindly go ahead with {{otp_words}}', 'ZPaceHQ', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceHub ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceHub', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceHub - OTP Digits', 'ZPaceHub: Kindly go ahead with [OTP]', 'ZPaceHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceHub - OTP Words', 'ZPaceHub: Kindly go ahead with {{otp_words}}', 'ZPaceHub', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceLab ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceLab', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceLab - OTP Digits', 'ZpaceLab: Kindly go ahead with [OTP]', 'ZpaceLab', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceLab - OTP Words', 'ZpaceLab: Kindly go ahead with {{otp_words}}', 'ZpaceLab', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceLab ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceLab', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceLab - OTP Digits', 'ZPaceLab: Kindly go ahead with [OTP]', 'ZPaceLab', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceLab - OTP Words', 'ZPaceLab: Kindly go ahead with {{otp_words}}', 'ZPaceLab', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceLink - OTP Digits', 'ZpaceLink: Kindly go ahead with [OTP]', 'ZpaceLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceLink - OTP Words', 'ZpaceLink: Kindly go ahead with {{otp_words}}', 'ZpaceLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceLink ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceLink', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceLink - OTP Digits', 'ZPaceLink: Kindly go ahead with [OTP]', 'ZPaceLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceLink - OTP Words', 'ZPaceLink: Kindly go ahead with {{otp_words}}', 'ZPaceLink', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceNet - OTP Digits', 'ZpaceNet: Kindly go ahead with [OTP]', 'ZpaceNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceNet - OTP Words', 'ZpaceNet: Kindly go ahead with {{otp_words}}', 'ZpaceNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceNet ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceNet', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceNet - OTP Digits', 'ZPaceNet: Kindly go ahead with [OTP]', 'ZPaceNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceNet - OTP Words', 'ZPaceNet: Kindly go ahead with {{otp_words}}', 'ZPaceNet', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceOne - OTP Digits', 'ZpaceOne: Kindly go ahead with [OTP]', 'ZpaceOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceOne - OTP Words', 'ZpaceOne: Kindly go ahead with {{otp_words}}', 'ZpaceOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceOne ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceOne', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceOne - OTP Digits', 'ZPaceOne: Kindly go ahead with [OTP]', 'ZPaceOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceOne - OTP Words', 'ZPaceOne: Kindly go ahead with {{otp_words}}', 'ZPaceOne', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceOps - OTP Digits', 'ZpaceOps: Kindly go ahead with [OTP]', 'ZpaceOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceOps - OTP Words', 'ZpaceOps: Kindly go ahead with {{otp_words}}', 'ZpaceOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceOps ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceOps', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceOps - OTP Digits', 'ZPaceOps: Kindly go ahead with [OTP]', 'ZPaceOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceOps - OTP Words', 'ZPaceOps: Kindly go ahead with {{otp_words}}', 'ZPaceOps', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpacePro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpacePro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpacePro - OTP Digits', 'ZpacePro: Kindly go ahead with [OTP]', 'ZpacePro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpacePro - OTP Words', 'ZpacePro: Kindly go ahead with {{otp_words}}', 'ZpacePro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPacePro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPacePro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPacePro - OTP Digits', 'ZPacePro: Kindly go ahead with [OTP]', 'ZPacePro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPacePro - OTP Words', 'ZPacePro: Kindly go ahead with {{otp_words}}', 'ZPacePro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceSys - OTP Digits', 'ZpaceSys: Kindly go ahead with [OTP]', 'ZpaceSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceSys - OTP Words', 'ZpaceSys: Kindly go ahead with {{otp_words}}', 'ZpaceSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceSys - OTP Digits', 'ZPaceSys: Kindly go ahead with [OTP]', 'ZPaceSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceSys - OTP Words', 'ZPaceSys: Kindly go ahead with {{otp_words}}', 'ZPaceSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === Zpacetech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('Zpacetech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: Zpacetech - OTP Digits', 'Zpacetech: Kindly go ahead with [OTP]', 'Zpacetech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: Zpacetech - OTP Words', 'Zpacetech: Kindly go ahead with {{otp_words}}', 'Zpacetech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceTech - OTP Digits', 'ZPaceTech: Kindly go ahead with [OTP]', 'ZPaceTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceTech - OTP Words', 'ZPaceTech: Kindly go ahead with {{otp_words}}', 'ZPaceTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === Zpacex ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('Zpacex', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: Zpacex - OTP Digits', 'Zpacex: Kindly go ahead with [OTP]', 'Zpacex', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: Zpacex - OTP Words', 'Zpacex: Kindly go ahead with {{otp_words}}', 'Zpacex', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceX ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceX', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceX - OTP Digits', 'ZpaceX: Kindly go ahead with [OTP]', 'ZpaceX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceX - OTP Words', 'ZpaceX: Kindly go ahead with {{otp_words}}', 'ZpaceX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceX ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceX', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceX - OTP Digits', 'ZPaceX: Kindly go ahead with [OTP]', 'ZPaceX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceX - OTP Words', 'ZPaceX: Kindly go ahead with {{otp_words}}', 'ZPaceX', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZpaceZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZpaceZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceZone - OTP Digits', 'ZpaceZone: Kindly go ahead with [OTP]', 'ZpaceZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZpaceZone - OTP Words', 'ZpaceZone: Kindly go ahead with {{otp_words}}', 'ZpaceZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPaceZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPaceZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceZone - OTP Digits', 'ZPaceZone: Kindly go ahead with [OTP]', 'ZPaceZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPaceZone - OTP Words', 'ZPaceZone: Kindly go ahead with {{otp_words}}', 'ZPaceZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZPro ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZPro', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPro - OTP Digits', 'ZPro: Kindly go ahead with [OTP]', 'ZPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZPro - OTP Words', 'ZPro: Kindly go ahead with {{otp_words}}', 'ZPro', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZSys ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZSys', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZSys - OTP Digits', 'ZSys: Kindly go ahead with [OTP]', 'ZSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZSys - OTP Words', 'ZSys: Kindly go ahead with {{otp_words}}', 'ZSys', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZTech ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZTech', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZTech - OTP Digits', 'ZTech: Kindly go ahead with [OTP]', 'ZTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZTech - OTP Words', 'ZTech: Kindly go ahead with {{otp_words}}', 'ZTech', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

    -- === ZZone ===
    INSERT INTO otp_alternative_senders
        (sender_id_string, service_provider_id, mno_id, status, max_usage_count, reset_interval_hours, current_usage_count, last_reset_at)
    VALUES
        ('ZZone', 6, 9, 'active', 2000, 24, 0, NOW())
    ON CONFLICT (service_provider_id, sender_id_string, mno_id) DO NOTHING
    RETURNING id INTO v_sender_id;

    IF v_sender_id IS NOT NULL THEN
        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZZone - OTP Digits', 'ZZone: Kindly go ahead with [OTP]', 'ZZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_digits_id;

        INSERT INTO otp_message_templates (name, content_template, default_brand_name, status)
        VALUES ('Jennie: ZZone - OTP Words', 'ZZone: Kindly go ahead with {{otp_words}}', 'ZZone', 'active')
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO v_tmpl_words_id;

        IF v_tmpl_digits_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_digits_id, NULL, 1, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;

        IF v_tmpl_words_id IS NOT NULL THEN
            INSERT INTO otp_sender_template_assignments
                (otp_alternative_sender_id, otp_message_template_id, mno_id, priority, status)
            VALUES (v_sender_id, v_tmpl_words_id, NULL, 2, 'active')
            ON CONFLICT (otp_alternative_sender_id, otp_message_template_id, mno_id) DO NOTHING;
        END IF;
    END IF;

END $$;

COMMIT;
