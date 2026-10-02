DO $$
DECLARE
    duplicate record;
    suffix_length integer;
    candidate text;
BEGIN
    FOR duplicate IN
        SELECT id, namespace_id, name
        FROM (
            SELECT id, namespace_id, name,
                   row_number() OVER (PARTITION BY namespace_id, name ORDER BY last_seen DESC, created_at ASC, id ASC) AS position
            FROM devices
            WHERE status = 'accepted'
        ) ranked
        WHERE position > 1
        ORDER BY namespace_id, name, position
    LOOP
        suffix_length := 8;

        LOOP
            IF suffix_length >= length(duplicate.id) THEN
                candidate := duplicate.id;
            ELSE
                candidate := rtrim(left(duplicate.name, 63 - suffix_length), '-') || '-' || left(duplicate.id, suffix_length);
            END IF;

            EXIT WHEN NOT EXISTS (
                SELECT 1
                FROM devices
                WHERE namespace_id = duplicate.namespace_id
                  AND status = 'accepted'
                  AND name = candidate
            );

            IF suffix_length >= length(duplicate.id) THEN
                RAISE EXCEPTION 'no free name for accepted device %', duplicate.id;
            END IF;

            suffix_length := suffix_length + 8;
        END LOOP;

        UPDATE devices SET name = candidate, updated_at = now() WHERE id = duplicate.id;
    END LOOP;
END
$$;

--bun:split

CREATE UNIQUE INDEX devices_accepted_name_unique
    ON devices USING btree (namespace_id, name)
    WHERE status = 'accepted';
