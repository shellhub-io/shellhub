-- PostgreSQL cannot drop a value from an enum, and rebuilding user_origin and user_auth_method
-- would mean rewriting every row of users to a value that carries different meaning. Rolling back
-- leaves the values in place: they are inert unless a user actually holds one.
SELECT 1;
