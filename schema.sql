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
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX rules_tenant_idx ON rules (tenant_id);

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

-- Rule <-> NotificationChannel mapping
CREATE TABLE rule_destinations
(
    rule_id    UUID REFERENCES rules (id) ON DELETE CASCADE,
    channel_id UUID REFERENCES notification_channels (id) ON DELETE CASCADE,
    PRIMARY KEY (rule_id, channel_id)
);

-- Alerts: when a rule is triggered on a specific event
CREATE TABLE alerts
(
    id           UUID PRIMARY KEY,
    tenant_id    UUID  NOT NULL,
    rule_id      UUID REFERENCES rules (id),
    event        JSONB NOT NULL,
    triggered_at TIMESTAMP DEFAULT now(),
    notified     BOOLEAN   DEFAULT FALSE
);
