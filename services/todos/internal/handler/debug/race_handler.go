package debug

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/chienha0903/Todo_App/services/todos/internal/domain/service"
)

// RaceHandler phục vụ debug HTTP endpoint cho race condition demo.
// Chỉ active khi ENABLE_DEBUG_RACE=true.
type RaceHandler struct {
	debugger *service.RaceDebugger
}

func NewRaceHandler(debugger *service.RaceDebugger) *RaceHandler {
	return &RaceHandler{debugger: debugger}
}

// RegisterRoutes đăng ký 2 route vào mux (Go 1.22+ với method+path pattern).
//
//	POST /debug/race/{id}         → không lock (demo lost update)
//	POST /debug/race/{id}/locked  → SELECT FOR UPDATE (demo fix)
func (h *RaceHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /debug/race/{id}", h.handleNoLock)
	mux.HandleFunc("POST /debug/race/{id}/locked", h.handleLocked)
}

func (h *RaceHandler) handleNoLock(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, false)
}

func (h *RaceHandler) handleLocked(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, true)
}

func (h *RaceHandler) handle(w http.ResponseWriter, r *http.Request, withLock bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid todo id", http.StatusBadRequest)
		return
	}

	var result *service.RaceResult
	if withLock {
		result, err = h.debugger.RunWithLock(r.Context(), id)
	} else {
		result, err = h.debugger.RunNoLock(r.Context(), id)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
