-- Step a: deduplicate existing rows non-destructively by renaming duplicates.
-- For each case-insensitive name group the oldest row (by created_at, ties broken
-- by id ASC) keeps its original name; every other row in that group is renamed to
-- "<base>-<first-8-chars-of-id>" where <base> is the lower-cased original name
-- truncated so the full result never exceeds 63 characters.
-- Trailing hyphens are stripped from the base before the suffix is appended, so
-- the resulting name always starts and ends with an alphanumeric character.
-- Re-running this statement is safe (idempotent): rows that were already renamed
-- by a previous run will not match the duplicate-detection predicate again.
WITH winners AS (
    -- One winner per lower(name) group: oldest created_at, smallest id on ties.
    SELECT DISTINCT ON (lower(name)) id
    FROM namespaces
    ORDER BY lower(name), created_at ASC, id ASC
),
duplicates AS (
    -- Every row that is NOT the winner of its group.
    SELECT n.id,
           lower(n.name) AS lower_name
    FROM   namespaces n
    WHERE  n.id NOT IN (SELECT id FROM winners)
)
UPDATE namespaces n
SET name = (
    -- Build "<base>-<suffix>" where suffix = first 8 hex chars of the UUID.
    -- Base = lower(name) truncated to at most 54 chars, then right-stripped of
    -- hyphens so the final name never starts/ends with '-'.
    rtrim(
        left(lower(n.name), 54),
        '-'
    ) || '-' || left(replace(n.id::text, '-', ''), 8)
)
FROM duplicates d
WHERE n.id = d.id;

--bun:split

-- Step b: create a case-insensitive unique index on lower(name).
CREATE UNIQUE INDEX namespaces_name_unique ON namespaces USING btree (lower(name));
