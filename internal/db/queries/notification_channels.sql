-- name: CreateNotificationChannel :one
INSERT INTO notification_channels (id, tenant_id, name, type, config)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, name, type, config, created_at;

-- name: ListNotificationChannels :many
SELECT id, tenant_id, name, type, config, created_at
FROM notification_channels
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: GetNotificationChannelByID :one
SELECT id, tenant_id, name, type, config, created_at
FROM notification_channels
WHERE id = $1;
