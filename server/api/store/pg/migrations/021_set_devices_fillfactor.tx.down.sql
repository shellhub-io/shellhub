-- Back to the default 100. Existing pages keep whatever slack they already have; only pages
-- allocated after this are packed full.
ALTER TABLE devices RESET (fillfactor);
