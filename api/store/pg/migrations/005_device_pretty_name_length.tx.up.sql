-- Widen pretty_name to fit longer OS-supplied PRETTY_NAME values
-- (/etc/os-release); varchar(64) overflowed on some distros.
ALTER TABLE devices ALTER COLUMN pretty_name TYPE character varying(256);
