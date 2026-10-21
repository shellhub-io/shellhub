CREATE TABLE webhooks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    namespace_id uuid NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    url text NOT NULL,
    secret_encrypted text NOT NULL,
    events text[] NOT NULL,
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);

--bun:split

CREATE INDEX webhooks_namespace_id_idx ON webhooks USING btree (namespace_id);

--bun:split

CREATE TABLE webhook_events (
    id uuid PRIMARY KEY,
    namespace_id uuid NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    type text NOT NULL,
    data jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL
);

--bun:split

CREATE TABLE webhook_deliveries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES webhook_events(id) ON DELETE CASCADE,
    webhook_id uuid NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    namespace_id uuid NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    status text NOT NULL DEFAULT 'pending',
    attempts integer NOT NULL DEFAULT 0,
    next_attempt_at timestamp with time zone NOT NULL,
    lease_until timestamp with time zone,
    last_status_code integer,
    last_error text,
    delivered_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT webhook_deliveries_event_webhook_unique UNIQUE (event_id, webhook_id),
    CONSTRAINT webhook_deliveries_status_check CHECK (status IN ('pending', 'delivered', 'failed'))
);

--bun:split

CREATE INDEX webhook_deliveries_due_idx ON webhook_deliveries USING btree (next_attempt_at) WHERE status = 'pending';

--bun:split

CREATE INDEX webhook_deliveries_webhook_id_created_at_idx ON webhook_deliveries USING btree (webhook_id, created_at DESC);
