package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/turbolytics/sqlsec/internal/auth"
	"github.com/turbolytics/sqlsec/internal/source"
	"go.uber.org/zap"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/turbolytics/sqlsec/internal"
	"github.com/turbolytics/sqlsec/internal/db"
)

type CreateWebhookRequest struct {
	Name   string   `json:"name"`
	Source string   `json:"source"`
	Events []string `json:"events"`
}

// Removed unused Server type declaration.

// NewWebhook creates a new Webhook handler with the given options.
func NewWebhook(queries *db.Queries, logger *zap.Logger) *Webhook {
	return &Webhook{
		queries: queries,
		logger:  logger,
	}
}

type Webhook struct {
	queries *db.Queries
	logger  *zap.Logger
}

func (wh *Webhook) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if !source.DefaultRegistry.IsEnabled(req.Source) {
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

	hook, err := wh.queries.CreateWebhook(r.Context(), db.CreateWebhookParams{
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

	webhook, err := wh.queries.GetWebhook(r.Context(), db.GetWebhookParams{
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
	webhook, err := wh.queries.GetWebhook(r.Context(), db.GetWebhookParams{
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

	validator := source.DefaultRegistry.GetValidator(webhook.Source)
	if validator == nil {
		http.Error(w, "unsupported source", http.StatusBadRequest)
		return
	}
	if err := validator.Validate(r, webhook.Secret); err != nil {
		wh.logger.Warn("validation failed", zap.Error(err))
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	parser := source.DefaultRegistry.GetParser(webhook.Source)
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
	event := struct {
		Time    time.Time
		Payload map[string]any
	}{
		Time:    time.Now().UTC(),
		Payload: payload,
	}
	wh.logger.Info("event received", zap.Any("event", event))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("event received"))
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
