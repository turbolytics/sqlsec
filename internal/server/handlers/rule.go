package handlers

import (
	"database/sql"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/turbolytics/sqlsec/internal"
	"github.com/turbolytics/sqlsec/internal/db/queries/events"
	"github.com/turbolytics/sqlsec/internal/db/queries/rules"
	"github.com/turbolytics/sqlsec/internal/engine/sandbox"
	"github.com/turbolytics/sqlsec/internal/source"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type RuleHandlers struct {
	logger *zap.Logger

	ruleQueries *rules.Queries
}

func NewRuleHandlers(l *zap.Logger, ruleQueries *rules.Queries) *RuleHandlers {
	return &RuleHandlers{
		logger:      l,
		ruleQueries: ruleQueries,
	}
}

type RuleCreateRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Source         string `json:"source"`
	EventType      string `json:"event_type"`
	Condition      string `json:"condition"`
	EvaluationType string `json:"evaluation_type"`
	AlertLevel     string `json:"alert_level"`
	Active         *bool  `json:"active,omitempty"`
}

// PATCH request for toggling active flag
// Only supports toggling 'active' for now
// Extendable for other fields

type RulePatchRequest struct {
	Active *bool `json:"active"`
}

type TestRuleRequest struct {
	Event map[string]interface{} `json:"event"`
}

type TestRuleResponse struct {
	Match      string                 `json:"match"`
	Error      string                 `json:"error,omitempty"`
	AlertLevel string                 `json:"alert_level"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

func (h *RuleHandlers) Create(w http.ResponseWriter, r *http.Request) {
	var req RuleCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	// Validate source
	if !source.DefaultRegistry.IsEnabled(source.Source(req.Source)) {
		http.Error(w, "unsupported source", http.StatusBadRequest)
		return
	}
	id := uuid.New()
	// TODO: get tenant_id from context/session
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	createdAt := time.Now().UTC()
	active := false
	if req.Active != nil {
		active = *req.Active
	}
	// Insert into DB
	dbRule, err := h.ruleQueries.CreateRule(r.Context(), rules.CreateRuleParams{
		ID:          id,
		TenantID:    tenantID,
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Source:      req.Source,
		EventType:   req.EventType,
		Sql:         req.Condition,
		EvalType:    req.EvaluationType,
		AlertLevel:  req.AlertLevel,
		CreatedAt:   createdAt,
		Active:      active,
	})
	if err != nil {
		h.logger.Error("failed to create rule", zap.Error(err))
		http.Error(w, "failed to create rule", http.StatusInternalServerError)
		return
	}
	resp := internal.Rule{
		ID:             dbRule.ID.String(),
		Name:           dbRule.Name,
		Description:    dbRule.Description.String,
		EvaluationType: internal.EvaluationType(dbRule.EvalType),
		EventSource:    dbRule.Source,
		EventType:      dbRule.EventType,
		SQL:            dbRule.Sql,
		CreatedAt:      dbRule.CreatedAt,
		AlertLevel:     internal.AlertLevel(dbRule.AlertLevel),
		Active:         dbRule.Active,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *RuleHandlers) List(w http.ResponseWriter, r *http.Request) {
	// TODO: get tenant_id from context/session
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	limit := int32(50)
	offset := int32(0)
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = int32(v)
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = int32(v)
		}
	}
	rules, err := h.ruleQueries.ListRules(r.Context(), rules.ListRulesParams{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		http.Error(w, "failed to list rules", http.StatusInternalServerError)
		return
	}
	resp := make([]map[string]interface{}, 0, len(rules))
	for _, rule := range rules {
		resp = append(resp, map[string]interface{}{
			"id":              rule.ID,
			"tenant_id":       rule.TenantID,
			"name":            rule.Name,
			"description":     rule.Description.String,
			"source":          rule.Source,
			"event_type":      rule.EventType,
			"condition":       rule.Sql,
			"evaluation_type": rule.EvalType,
			"alert_level":     rule.AlertLevel,
			"created_at":      rule.CreatedAt,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *RuleHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	// TODO: get tenant_id from context/session
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	err = h.ruleQueries.DeleteRule(r.Context(), rules.DeleteRuleParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		http.Error(w, "rule not found or failed to delete", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RuleHandlers) Test(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	_, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req TestRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	rule, err := h.ruleQueries.GetRuleByID(ctx, rules.GetRuleByIDParams{
		ID: uuid.MustParse(idStr),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get rule by ID", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	box, err := sandbox.New(
		ctx,
		sandbox.WithLogger(h.logger),
		sandbox.WithDuckDBMemoryConnection(),
	)

	if err != nil {
		h.logger.Error("failed to create sandbox", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(req.Event)
	if err != nil {
		h.logger.Error("failed to marshal event", zap.Error(err))
		http.Error(w, "invalid event format", http.StatusBadRequest)
		return
	}

	testEvent := &events.Event{
		RawPayload: b,
	}

	if err = box.AddEvent(ctx, testEvent); err != nil {
		h.logger.Error("failed to add event to sandbox", zap.Error(err))
		http.Error(w, "failed to add event", http.StatusInternalServerError)
		return
	}

	i, err := box.ExecuteRule(ctx, rule)
	if err != nil {
		h.logger.Error("failed to execute rule in sandbox", zap.Error(err))
	}

	// TODO: fetch rule, evaluate condition, return result
	resp := TestRuleResponse{
		Match:      strconv.FormatBool(i > 0),
		AlertLevel: rule.AlertLevel,
		Details: map[string]any{
			"rule_id":   rule.ID.String(),
			"rule_name": rule.Name,
			"sql":       rule.Sql,
			"event":     testEvent,
		},
	}

	if err != nil {
		resp.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *RuleHandlers) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	// TODO: get tenant_id from context/session
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	dbRule, err := h.ruleQueries.GetRuleByID(r.Context(), rules.GetRuleByIDParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		http.Error(w, "rule not found", http.StatusNotFound)
		return
	}
	resp := internal.Rule{
		ID:             dbRule.ID.String(),
		Name:           dbRule.Name,
		Description:    dbRule.Description.String,
		EvaluationType: internal.EvaluationType(dbRule.EvalType),
		EventSource:    dbRule.Source,
		EventType:      dbRule.EventType,
		SQL:            dbRule.Sql,
		CreatedAt:      dbRule.CreatedAt,
		AlertLevel:     internal.AlertLevel(dbRule.AlertLevel),
		Active:         dbRule.Active,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *RuleHandlers) Patch(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req RulePatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Active == nil {
		http.Error(w, "missing active field", http.StatusBadRequest)
		return
	}
	// TODO: get tenant_id from context/session
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	// Update rule
	dbRule, err := h.ruleQueries.UpdateRuleActive(r.Context(), rules.UpdateRuleActiveParams{
		ID:       id,
		TenantID: tenantID,
		Active:   *req.Active,
	})
	if err != nil {
		http.Error(w, "failed to update rule", http.StatusInternalServerError)
		return
	}
	resp := internal.Rule{
		ID:             dbRule.ID.String(),
		Name:           dbRule.Name,
		Description:    dbRule.Description.String,
		EvaluationType: internal.EvaluationType(dbRule.EvalType),
		EventSource:    dbRule.Source,
		EventType:      dbRule.EventType,
		SQL:            dbRule.Sql,
		CreatedAt:      dbRule.CreatedAt,
		AlertLevel:     internal.AlertLevel(dbRule.AlertLevel),
		Active:         dbRule.Active,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
