// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: webhooks.sql

package webhooks

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
)

const createWebhook = `-- name: CreateWebhook :one
INSERT INTO webhooks (id, tenant_id, name, secret, source, events, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING id, tenant_id, name, secret, source, events, created_at
`

type CreateWebhookParams struct {
	ID        uuid.UUID       `json:"id"`
	TenantID  uuid.UUID       `json:"tenant_id"`
	Name      string          `json:"name"`
	Secret    string          `json:"secret"`
	Source    string          `json:"source"`
	Events    json.RawMessage `json:"events"`
	CreatedAt sql.NullTime    `json:"created_at"`
}

func (q *Queries) CreateWebhook(ctx context.Context, arg CreateWebhookParams) (Webhook, error) {
	row := q.db.QueryRowContext(ctx, createWebhook,
		arg.ID,
		arg.TenantID,
		arg.Name,
		arg.Secret,
		arg.Source,
		arg.Events,
		arg.CreatedAt,
	)
	var i Webhook
	err := row.Scan(
		&i.ID,
		&i.TenantID,
		&i.Name,
		&i.Secret,
		&i.Source,
		&i.Events,
		&i.CreatedAt,
	)
	return i, err
}

const getWebhook = `-- name: GetWebhook :one
SELECT id, tenant_id, name, secret, source, events, created_at FROM webhooks WHERE id = $1 AND tenant_id = $2
`

type GetWebhookParams struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`
}

func (q *Queries) GetWebhook(ctx context.Context, arg GetWebhookParams) (Webhook, error) {
	row := q.db.QueryRowContext(ctx, getWebhook, arg.ID, arg.TenantID)
	var i Webhook
	err := row.Scan(
		&i.ID,
		&i.TenantID,
		&i.Name,
		&i.Secret,
		&i.Source,
		&i.Events,
		&i.CreatedAt,
	)
	return i, err
}
