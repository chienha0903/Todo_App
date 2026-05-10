package http

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"io"
	nethttp "net/http"
	"strconv"
	"strings"
	"time"

	todopb "github.com/chienha0903/Todo_App/proto/todo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxRequestBodyBytes = 1 << 20

type TodoHandler struct {
	todoClient todopb.TodoServiceClient
	timeout    time.Duration
}

type createTodoRequest struct {
	UserID      int64  `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	DueDate     string `json:"due_date,omitempty"`
}

type updateTodoRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Priority    string `json:"priority,omitempty"`
	Status      string `json:"status,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
}

type todoResponse struct {
	Todo *todoDTO `json:"todo"`
}

type listTodosResponse struct {
	Todos []*todoDTO `json:"todos"`
}

type todoDTO struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	DueDate     string `json:"due_date,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewTodoHandler(todoClient todopb.TodoServiceClient, timeout time.Duration) *TodoHandler {
	return &TodoHandler{
		todoClient: todoClient,
		timeout:    timeout,
	}
}

func (h *TodoHandler) Routes() *nethttp.ServeMux {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /todos", h.createTodo)
	mux.HandleFunc("GET /todos", h.listTodos)
	mux.HandleFunc("GET /todos/{id}", h.getTodo)
	mux.HandleFunc("PUT /todos/{id}", h.updateTodo)
	mux.HandleFunc("DELETE /todos/{id}", h.deleteTodo)
	return mux
}

func (h *TodoHandler) health(w nethttp.ResponseWriter, r *nethttp.Request) {
	writeJSON(w, nethttp.StatusOK, map[string]string{"status": "ok"})
}

func (h *TodoHandler) createTodo(w nethttp.ResponseWriter, r *nethttp.Request) {
	body, err := decodeCreateTodoRequest(w, r)
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := h.contextWithTimeout(r.Context())
	defer cancel()

	resp, err := h.todoClient.CreateTodo(ctx, createTodoProtoRequest(body))
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, nethttp.StatusCreated, todoResponse{Todo: toTodoDTO(resp.GetTodo())})
}

func (h *TodoHandler) getTodo(w nethttp.ResponseWriter, r *nethttp.Request) {
	id, ok := parsePositiveInt64(r.PathValue("id"))
	if !ok {
		writeError(w, nethttp.StatusBadRequest, "id must be a positive integer")
		return
	}

	ctx, cancel := h.contextWithTimeout(r.Context())
	defer cancel()

	resp, err := h.todoClient.GetTodo(ctx, &todopb.GetTodoRequest{Id: id})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, nethttp.StatusOK, todoResponse{Todo: toTodoDTO(resp.GetTodo())})
}

func (h *TodoHandler) listTodos(w nethttp.ResponseWriter, r *nethttp.Request) {
	userID, ok := parsePositiveInt64(r.URL.Query().Get("user_id"))
	if !ok {
		writeError(w, nethttp.StatusBadRequest, "user_id query parameter must be a positive integer")
		return
	}

	ctx, cancel := h.contextWithTimeout(r.Context())
	defer cancel()

	resp, err := h.todoClient.ListTodos(ctx, &todopb.ListTodosRequest{UserId: userID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, nethttp.StatusOK, listTodosResponse{Todos: toTodoDTOs(resp.GetTodos())})
}

func (h *TodoHandler) updateTodo(w nethttp.ResponseWriter, r *nethttp.Request) {
	id, ok := parsePositiveInt64(r.PathValue("id"))
	if !ok {
		writeError(w, nethttp.StatusBadRequest, "id must be a positive integer")
		return
	}

	body, err := decodeUpdateTodoRequest(w, r)
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := h.contextWithTimeout(r.Context())
	defer cancel()

	resp, err := h.todoClient.UpdateTodo(ctx, updateTodoProtoRequest(id, body))
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, nethttp.StatusOK, todoResponse{Todo: toTodoDTO(resp.GetTodo())})
}

func (h *TodoHandler) deleteTodo(w nethttp.ResponseWriter, r *nethttp.Request) {
	id, ok := parsePositiveInt64(r.PathValue("id"))
	if !ok {
		writeError(w, nethttp.StatusBadRequest, "id must be a positive integer")
		return
	}

	ctx, cancel := h.contextWithTimeout(r.Context())
	defer cancel()

	if _, err := h.todoClient.DeleteTodo(ctx, &todopb.DeleteTodoRequest{Id: id}); err != nil {
		writeGRPCError(w, err)
		return
	}

	w.WriteHeader(nethttp.StatusNoContent)
}

func (h *TodoHandler) contextWithTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, h.timeout)
}

func decodeCreateTodoRequest(w nethttp.ResponseWriter, r *nethttp.Request) (createTodoRequest, error) {
	var body createTodoRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return createTodoRequest{}, err
	}
	if err := validateCreateTodoRequest(body); err != nil {
		return createTodoRequest{}, err
	}
	return body, nil
}

func decodeUpdateTodoRequest(w nethttp.ResponseWriter, r *nethttp.Request) (updateTodoRequest, error) {
	var body updateTodoRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return updateTodoRequest{}, err
	}
	if err := validateUpdateTodoRequest(body); err != nil {
		return updateTodoRequest{}, err
	}
	return body, nil
}

func validateCreateTodoRequest(req createTodoRequest) error {
	if req.UserID <= 0 {
		return stderrors.New("user_id must be a positive integer")
	}
	if strings.TrimSpace(req.Title) == "" {
		return stderrors.New("title is required")
	}
	if strings.TrimSpace(req.Description) == "" {
		return stderrors.New("description is required")
	}
	if strings.TrimSpace(req.Priority) == "" {
		return stderrors.New("priority is required")
	}
	return validateOptionalRFC3339(req.DueDate)
}

func validateUpdateTodoRequest(req updateTodoRequest) error {
	if !hasUpdateTodoField(req) {
		return stderrors.New("at least one field is required")
	}
	return validateOptionalRFC3339(req.DueDate)
}

func hasUpdateTodoField(req updateTodoRequest) bool {
	return strings.TrimSpace(req.Title) != "" ||
		strings.TrimSpace(req.Description) != "" ||
		strings.TrimSpace(req.Priority) != "" ||
		strings.TrimSpace(req.Status) != "" ||
		strings.TrimSpace(req.DueDate) != ""
}

func validateOptionalRFC3339(value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	if _, err := time.Parse(time.RFC3339, value); err != nil {
		return stderrors.New("due_date must be RFC3339 format")
	}
	return nil
}

func parsePositiveInt64(value string) (int64, bool) {
	if strings.TrimSpace(value) == "" {
		return 0, false
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, false
	}
	return parsed, true
}

func decodeJSON(w nethttp.ResponseWriter, r *nethttp.Request, dst any) error {
	r.Body = nethttp.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return stderrors.New("invalid JSON body")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return stderrors.New("request body must contain a single JSON object")
	}
	return nil
}

func writeGRPCError(w nethttp.ResponseWriter, err error) {
	if stderrors.Is(err, context.DeadlineExceeded) {
		writeError(w, nethttp.StatusGatewayTimeout, "request timed out")
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		writeError(w, nethttp.StatusInternalServerError, "internal server error")
		return
	}

	httpStatus := httpStatusFromGRPCCode(st.Code())
	message := publicGRPCMessage(httpStatus, st.Message())

	writeError(w, httpStatus, message)
}

func httpStatusFromGRPCCode(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return nethttp.StatusBadRequest
	case codes.NotFound:
		return nethttp.StatusNotFound
	case codes.Unauthenticated:
		return nethttp.StatusUnauthorized
	case codes.PermissionDenied:
		return nethttp.StatusForbidden
	case codes.DeadlineExceeded:
		return nethttp.StatusGatewayTimeout
	case codes.Unavailable:
		return nethttp.StatusServiceUnavailable
	default:
		return nethttp.StatusInternalServerError
	}
}

func publicGRPCMessage(httpStatus int, message string) string {
	if httpStatus == nethttp.StatusInternalServerError {
		return "internal server error"
	}
	return message
}

func writeError(w nethttp.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, errorResponse{Error: message})
}

func writeJSON(w nethttp.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}

func createTodoProtoRequest(body createTodoRequest) *todopb.CreateTodoRequest {
	return &todopb.CreateTodoRequest{
		UserId:      body.UserID,
		Title:       body.Title,
		Description: body.Description,
		Priority:    body.Priority,
		DueDate:     body.DueDate,
	}
}

func updateTodoProtoRequest(id int64, body updateTodoRequest) *todopb.UpdateTodoRequest {
	return &todopb.UpdateTodoRequest{
		Id:          id,
		Title:       body.Title,
		Description: body.Description,
		Priority:    body.Priority,
		Status:      body.Status,
		DueDate:     body.DueDate,
	}
}

func toTodoDTOs(todos []*todopb.Todo) []*todoDTO {
	items := make([]*todoDTO, 0, len(todos))
	for _, todo := range todos {
		items = append(items, toTodoDTO(todo))
	}
	return items
}

func toTodoDTO(todo *todopb.Todo) *todoDTO {
	if todo == nil {
		return nil
	}
	return &todoDTO{
		ID:          todo.Id,
		UserID:      todo.UserId,
		Title:       todo.Title,
		Description: todo.Description,
		Status:      todo.Status,
		Priority:    todo.Priority,
		DueDate:     todo.DueDate,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}
}
