package handlers

import (
	"github.com/turbolytics/sqlsec/internal/db"
	"net/http"
)

type DestinationHandlers struct {
	queries *db.Queries
}

func NewDestinationHandlers(queries *db.Queries) *DestinationHandlers {
	return &DestinationHandlers{queries: queries}
}

type DestinationCreateRequest struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type TestDestinationRequest struct {
	Message string `json:"message"`
}

type TestDestinationResponse struct {
	Success bool                   `json:"success"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (h *DestinationHandlers) Create(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse rule id, validate, create destination, attach to rule, return destination
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *DestinationHandlers) List(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse rule id, list all destinations attached to rule
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *DestinationHandlers) Get(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse rule id and dest_id, return destination if attached
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *DestinationHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse rule id and dest_id, detach destination from rule
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *DestinationHandlers) Test(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse rule id and dest_id, send test message to destination
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
