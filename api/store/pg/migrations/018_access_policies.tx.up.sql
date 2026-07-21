-- Access Policies: namespace-scoped, default-deny authorization grants for the
-- identity-based SSH access mode. A policy grants a subject (user, role, or all
-- members) access to the devices selected by its filter, as the listed logins.
CREATE TABLE access_policies (
    id uuid NOT NULL,
    namespace_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name character varying NOT NULL,
    subject_type character varying NOT NULL,
    subject_value character varying DEFAULT ''::character varying NOT NULL,
    filter_hostname character varying DEFAULT ''::character varying,
    logins text[] NOT NULL DEFAULT '{}'::text[],
    effect character varying NOT NULL DEFAULT 'allow',
    source_ip text[] NOT NULL DEFAULT '{}'::text[],
    PRIMARY KEY (id),
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE
);

--bun:split

CREATE INDEX access_policies_namespace_id ON access_policies USING btree (namespace_id);

--bun:split

CREATE TABLE access_policy_tags (
    access_policy_id uuid NOT NULL,
    tag_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    PRIMARY KEY (access_policy_id, tag_id),
    FOREIGN KEY (access_policy_id) REFERENCES access_policies(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

--bun:split

CREATE INDEX access_policy_tags_access_policy_id ON access_policy_tags USING btree (access_policy_id);

--bun:split

CREATE INDEX access_policy_tags_tag_id ON access_policy_tags USING btree (tag_id);
