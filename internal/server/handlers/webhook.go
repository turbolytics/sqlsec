package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/turbolytics/sqlsec/internal/auth"
	"github.com/turbolytics/sqlsec/internal/db/queries/events"
	"github.com/turbolytics/sqlsec/internal/db/queries/webhooks"
	"github.com/turbolytics/sqlsec/internal/source"
	"go.uber.org/zap"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/turbolytics/sqlsec/internal"
)

type CreateWebhookRequest struct {
	Name   string   `json:"name"`
	Source string   `json:"source"`
	Events []string `json:"events"`
}

// Removed unused Server type declaration.

// NewWebhook creates a new Webhook handler with the given options.
func NewWebhook(
	dbConn *sql.DB,
	eventQueries *events.Queries,
	webhookQueries *webhooks.Queries,
	logger *zap.Logger,
) *Webhook {

	return &Webhook{
		db:             dbConn,
		eventQueries:   eventQueries,
		webhookQueries: webhookQueries,
		logger:         logger,
	}
}

type Webhook struct {
	eventQueries   *events.Queries
	webhookQueries *webhooks.Queries
	db             *sql.DB
	logger         *zap.Logger
}

func (wh *Webhook) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if !source.DefaultRegistry.IsEnabled(source.Source(req.Source)) {
		http.Error(w, "unsupported source", http.StatusBadRequest)
		return
	}

	// Hardcoded tenant_id for now
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	id := uuid.New()
	createdAt := time.Now().UTC()
	secret, err := auth.GenerateSecret()
	if err != nil {
		wh.logger.Error("failed to generate secret", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	hook, err := wh.webhookQueries.CreateWebhook(r.Context(), webhooks.CreateWebhookParams{
		ID:        id,
		TenantID:  tenantID,
		Name:      req.Name,
		Secret:    secret,
		Source:    req.Source,
		Events:    mustMarshalEvents(req.Events),
		CreatedAt: sql.NullTime{Time: createdAt, Valid: true},
	})
	if err != nil {
		wh.logger.Error("failed to create webhook", zap.Error(err))
		http.Error(w, "failed to create webhook", http.StatusInternalServerError)
		return
	}

	resp := internal.Webhook{
		ID:        hook.ID,
		TenantID:  hook.TenantID,
		Name:      hook.Name,
		Secret:    hook.Secret,
		Source:    hook.Source,
		CreatedAt: hook.CreatedAt.Time,
		Events:    mustUnmarshalEvents(hook.Events),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (wh *Webhook) Get(w http.ResponseWriter, r *http.Request) {
	webhookID := chi.URLParam(r, "id")
	wh.logger.Info("Get webhook called", zap.String("id", webhookID))
	wh.logger.Debug("Get webhook called with ID", zap.String("webhookID", webhookID))
	id, err := uuid.Parse(webhookID)
	if err != nil {
		http.Error(w, "invalid webhook id", http.StatusBadRequest)
		return
	}
	// Hardcoded tenant_id for now
	tid := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	webhook, err := wh.webhookQueries.GetWebhook(r.Context(), webhooks.GetWebhookParams{
		ID:       id,
		TenantID: tid,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "webhook not found", http.StatusNotFound)
			return
		}
		wh.logger.Error("failed to get webhook", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(internal.Webhook{
		ID:        webhook.ID,
		TenantID:  webhook.TenantID,
		Name:      webhook.Name,
		Secret:    webhook.Secret,
		Source:    webhook.Source,
		CreatedAt: webhook.CreatedAt.Time,
		Events:    mustUnmarshalEvents(webhook.Events),
	})
}

func (wh *Webhook) Event(w http.ResponseWriter, r *http.Request) {
	webhookID := chi.URLParam(r, "webhook_id")
	id, err := uuid.Parse(webhookID)
	if err != nil {
		http.Error(w, "invalid webhook id", http.StatusBadRequest)
		return
	}
	// Hardcoded tenant_id for now
	tid := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	webhook, err := wh.webhookQueries.GetWebhook(r.Context(), webhooks.GetWebhookParams{
		ID:       id,
		TenantID: tid,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "webhook not found", http.StatusNotFound)
			return
		}
		wh.logger.Error("failed to get webhook", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	validator := source.DefaultRegistry.GetValidator(source.Source(webhook.Source))
	if validator == nil {
		http.Error(w, "unsupported source", http.StatusBadRequest)
		return
	}
	if err := validator.Validate(r, webhook.Secret); err != nil {
		wh.logger.Warn("validation failed", zap.Error(err))
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	parser := source.DefaultRegistry.GetParser(source.Source(webhook.Source))
	if parser == nil {
		http.Error(w, "unsupported source", http.StatusBadRequest)
		return
	}
	payload, err := parser.Parse(r)
	if err != nil {
		wh.logger.Warn("failed to parse payload", zap.Error(err))
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	// Insert event and queue in a transaction
	tx, err := wh.db.BeginTx(r.Context(), nil)
	if err != nil {
		wh.logger.Error("failed to begin transaction", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	q := wh.eventQueries.WithTx(tx)
	eventType, err := parser.Type(r)
	if err != nil {
		tx.Rollback()
		wh.logger.Warn("failed to get event type", zap.Error(err))
		http.Error(w, "invalid event type", http.StatusBadRequest)
		return
	}
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		tx.Rollback()
		wh.logger.Error("failed to marshal payload", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	event, err := q.InsertEvent(r.Context(), events.InsertEventParams{
		TenantID:   webhook.TenantID,
		WebhookID:  webhook.ID,
		Source:     webhook.Source,
		EventType:  eventType,
		Action:     sql.NullString{}, // Set this if you can extract from payload
		RawPayload: rawPayload,
		DedupHash:  sql.NullString{}, // Set this if you want to deduplicate
	})
	if err != nil {
		tx.Rollback()
		wh.logger.Error("failed to insert event", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	_, err = q.InsertEventProcessingQueue(r.Context(), event.ID)
	if err != nil {
		tx.Rollback()
		wh.logger.Error("failed to queue event for processing", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		wh.logger.Error("failed to commit transaction", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	eventLog := struct {
		Time    time.Time
		Payload map[string]any
	}{
		Time:    time.Now().UTC(),
		Payload: payload,
	}
	wh.logger.Info("event received", zap.Any("event", eventLog))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("event received and queued"))
}

func mustMarshalEvents(events []string) []byte {
	data, err := json.Marshal(events)
	if err != nil {
		panic(err)
	}
	return data
}

func mustUnmarshalEvents(data []byte) []string {
	var events []string
	if err := json.Unmarshal(data, &events); err != nil {
		panic(err)
	}
	return events
}
