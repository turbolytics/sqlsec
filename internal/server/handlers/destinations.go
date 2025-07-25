package handlers

import (
	"encoding/json"
	"github.com/turbolytics/sqlsec/internal/db/queries/notificationchannels"
	"github.com/turbolytics/sqlsec/internal/db/queries/rules"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/turbolytics/sqlsec/internal"
)

type DestinationHandlers struct {
	ruleQueries                *rules.Queries
	notificationChannelQueries *notificationchannels.Queries
}

func NewDestinationHandlers(
	ruleQueries *rules.Queries,
	notificationChannelQueries *notificationchannels.Queries,
) *DestinationHandlers {
	return &DestinationHandlers{
		ruleQueries:                ruleQueries,
		notificationChannelQueries: notificationChannelQueries,
	}
}

type DestinationCreateRequest struct {
	RuleID    string `json:"rule_id"`
	ChannelID string `json:"channel_id"`
}

type TestDestinationRequest struct {
	Message string `json:"message"`
}

type TestDestinationResponse struct {
	Success bool                   `json:"success"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (h *DestinationHandlers) Create(w http.ResponseWriter, r *http.Request) {
	ruleIDStr := chi.URLParam(r, "id")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		http.Error(w, "invalid rule id", http.StatusBadRequest)
		return
	}
	// TODO: get tenant_id from context/session
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	// Check that rule exists and is owned by tenant
	_, err = h.ruleQueries.GetRuleByID(r.Context(), rules.GetRuleByIDParams{
		ID:       ruleID,
		TenantID: tenantID,
	})
	if err != nil {
		http.Error(w, "rule not found", http.StatusNotFound)
		return
	}
	// Parse request body
	var req DestinationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	channelID, err := uuid.Parse(req.ChannelID)
	if err != nil {
		http.Error(w, "invalid channel_id", http.StatusBadRequest)
		return
	}
	// Check that the channel exists and is owned by tenant
	channel, err := h.notificationChannelQueries.GetNotificationChannelByID(r.Context(), channelID)
	if err != nil || channel.TenantID != tenantID {
		http.Error(w, "notification channel not found or not owned by tenant", http.StatusNotFound)
		return
	}
	// Attach destination to rule
	_, err = h.ruleQueries.CreateRuleDestination(r.Context(), rules.CreateRuleDestinationParams{
		RuleID:    ruleID,
		ChannelID: channelID,
	})
	if err != nil {
		http.Error(w, "failed to attach destination", http.StatusInternalServerError)
		return
	}
	resp := internal.NotificationChannel{
		ID:        channel.ID,
		TenantID:  channel.TenantID,
		Name:      channel.Name,
		Type:      channel.Type,
		Config:    channel.Config,
		CreatedAt: channel.CreatedAt.Time,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *DestinationHandlers) List(w http.ResponseWriter, r *http.Request) {
	ruleIDStr := chi.URLParam(r, "id")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		http.Error(w, "invalid rule id", http.StatusBadRequest)
		return
	}
	// TODO: get tenant_id from context/session
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	// Check that rule exists and is owned by tenant
	_, err = h.ruleQueries.GetRuleByID(r.Context(), rules.GetRuleByIDParams{
		ID:       ruleID,
		TenantID: tenantID,
	})
	if err != nil {
		http.Error(w, "rule not found", http.StatusNotFound)
		return
	}
	channels, err := h.ruleQueries.ListNotificationChannelsForRule(r.Context(), ruleID)
	if err != nil {
		http.Error(w, "failed to list destinations", http.StatusInternalServerError)
		return
	}
	var dests []internal.NotificationChannel
	for _, ch := range channels {
		dests = append(dests, internal.NotificationChannel{
			ID:        ch.ID,
			TenantID:  ch.TenantID,
			Name:      ch.Name,
			Type:      ch.Type,
			Config:    ch.Config,
			CreatedAt: ch.CreatedAt.Time,
		})
	}
	if len(dests) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dests)
}

func (h *DestinationHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	ruleIDStr := chi.URLParam(r, "id")
	destIDStr := chi.URLParam(r, "dest_id")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		http.Error(w, "invalid rule id", http.StatusBadRequest)
		return
	}
	destID, err := uuid.Parse(destIDStr)
	if err != nil {
		http.Error(w, "invalid destination id", http.StatusBadRequest)
		return
	}
	// TODO: get tenant_id from context/session
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	// Check that rule exists and is owned by tenant
	_, err = h.ruleQueries.GetRuleByID(r.Context(), rules.GetRuleByIDParams{
		ID:       ruleID,
		TenantID: tenantID,
	})
	if err != nil {
		http.Error(w, "rule not found", http.StatusNotFound)
		return
	}
	// Delete the rule destination
	err = h.ruleQueries.DeleteRuleDestination(r.Context(), rules.DeleteRuleDestinationParams{
		RuleID:    ruleID,
		ChannelID: destID,
	})
	if err != nil {
		http.Error(w, "failed to detach destination", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DestinationHandlers) Test(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse rule id and dest_id, send test message to destination
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
