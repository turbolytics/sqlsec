CREATE
EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tenants
CREATE TABLE tenants
(
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ      DEFAULT now()
);

-- Webhooks (stream source)
CREATE TABLE webhooks
(
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id  UUID  NOT NULL REFERENCES tenants (id),
    name       TEXT  NOT NULL,
    secret     TEXT  NOT NULL,
    source     TEXT  NOT NULL, -- e.g., 'github'
    events     JSONB NOT NULL   DEFAULT '[]',
    created_at TIMESTAMPTZ      DEFAULT now()
);

insert into tenants (id, name)
VALUES ('00000000-0000-0000-0000-000000000000', 'root');

-- Rules: User-defined SQL checks on live events
CREATE TABLE rules
(
    id          UUID PRIMARY KEY,
    tenant_id   UUID        NOT NULL,
    name        TEXT        NOT NULL,
    description TEXT,
    source      TEXT        NOT NULL, -- e.g. 'github'
    event_type  TEXT        NOT NULL, -- e.g. 'pull_request'
    sql         TEXT        NOT NULL, -- just the WHERE clause
    eval_type   TEXT        NOT NULL DEFAULT 'LIVE_TRIGGER'
        CHECK (eval_type IN ('LIVE_TRIGGER')),
    alert_level TEXT        NOT NULL DEFAULT 'MEDIUM'
        CHECK (alert_level IN ('LOW', 'MEDIUM', 'HIGH')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    active      BOOLEAN     NOT NULL DEFAULT false
);

CREATE INDEX rules_tenant_idx ON rules (tenant_id);
CREATE UNIQUE INDEX rules_tenant_source_name_idx ON rules (tenant_id, source, name);

-- Notification Channels: Slack, Webhook, Email, etc.
CREATE TABLE notification_channels
(
    id         UUID PRIMARY KEY,
    tenant_id  UUID  NOT NULL,
    name       TEXT  NOT NULL, -- Human-readable name for identification
    type       TEXT  NOT NULL, -- 'slack', 'webhook', 'email'
    config     JSONB NOT NULL, -- token, URL, etc.
    created_at TIMESTAMP DEFAULT now()
);

CREATE UNIQUE INDEX notification_channels_tenant_name_idx ON notification_channels (tenant_id, name);

-- Rule <-> NotificationChannel mapping
CREATE TABLE rule_destinations
(
    rule_id    UUID REFERENCES rules (id) ON DELETE CASCADE,
    channel_id UUID REFERENCES notification_channels (id) ON DELETE CASCADE,
    PRIMARY KEY (rule_id, channel_id)
);

-- Ingested Events (raw and parsed)
CREATE TABLE events
(
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID  NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    webhook_id  UUID  NOT NULL REFERENCES webhooks (id) ON DELETE CASCADE,

    -- Metadata for routing/evaluation
    source      TEXT  NOT NULL, -- e.g. 'github'
    event_type  TEXT  NOT NULL, -- e.g. 'pull_request'
    action      TEXT,           -- e.g. 'opened', 'closed' (nullable for flexibility)

    -- Actual event payload
    raw_payload JSONB NOT NULL,

    -- Optional: deduplication or trace
    dedup_hash  TEXT UNIQUE,    -- SHA256 hash of raw_payload, optional
    received_at TIMESTAMPTZ      DEFAULT now()
);

-- Alerts: when a rule is triggered on a specific event
CREATE TABLE alerts
(
    id           UUID PRIMARY KEY,
    tenant_id    UUID NOT NULL,
    rule_id      UUID NOT NULL REFERENCES rules (id),
    event_id     UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    triggered_at TIMESTAMP DEFAULT now(),
    notified     BOOLEAN   DEFAULT FALSE
);

-- Indexes for fast rule lookup

CREATE TABLE event_processing_queue
(
    id           SERIAL PRIMARY KEY,
    event_id     UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'done', 'failed')),
    locked_at    TIMESTAMPTZ,
    locked_by    TEXT,
    processed_at TIMESTAMPTZ,
    error        TEXT
);

-- Alert Processing Queue: same pattern as event_processing_queue
CREATE TABLE alert_processing_queue
(
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_id     UUID NOT NULL REFERENCES alerts (id),
    status       TEXT NOT NULL    DEFAULT 'pending',
    locked_at    TIMESTAMPTZ,
    locked_by    TEXT,
    processed_at TIMESTAMPTZ,
    error        TEXT
);

CREATE INDEX alert_processing_queue_status_idx ON alert_processing_queue (status);
CREATE INDEX alert_processing_queue_alert_id_idx ON alert_processing_queue (alert_id);

CREATE TABLE alert_deliveries
(
    alert_id   UUID NOT NULL REFERENCES alerts (id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES notification_channels (id) ON DELETE CASCADE,
    status     TEXT NOT NULL CHECK (status IN ('pending', 'delivered', 'failed')),
    attempt_at TIMESTAMPTZ DEFAULT now(),
    error      TEXT,
    PRIMARY KEY (alert_id, channel_id)
);