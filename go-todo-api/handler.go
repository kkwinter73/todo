package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type TodoHandler struct {
	storage Storage
}

func NewTodoHandler(s Storage) *TodoHandler {
	return &TodoHandler{
		storage: s,
	}
}

// GET /todos
func (h *TodoHandler) ListTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := h.storage.GetAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "データ取得に失敗しました")
		return
	}
	writeJSON(w, http.StatusOK, todos)
}

// POST /todos
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "リクエストボディが不正です")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "titleは必須です")
		return
	}

	todo, err := h.storage.Create(r.Context(), req.Title)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "作成に失敗しました")
		return
	}

	writeJSON(w, http.StatusCreated, todo)
}

// PATCH /todos/{id}/done
func (h *TodoHandler) DoneTodo(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "IDが不正です")
		return
	}

	err = h.storage.Done(r.Context(), id)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrTodoNotFound):
		writeError(w, http.StatusNotFound, "タスクが見つかりません")
	case errors.Is(err, ErrAlreadyDone):
		writeError(w, http.StatusConflict, "タスクは既に完了しています")
	default:
		writeError(w, http.StatusInternalServerError, "処理に失敗しました")
	}
}

// DELETE /todos/{id}
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "IDが不正です")
		return
	}

	err = h.storage.Delete(r.Context(), id)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrTodoNotFound):
		writeError(w, http.StatusNotFound, "タスクが見つかりません")
	default:
		writeError(w, http.StatusInternalServerError, "処理に失敗しました")
	}
}

// parseIDFromURL はパスパラメータ {id} を整数として取り出す。
// Go 1.22+ の r.PathValue を使うので chi.URLParam は不要になった。
func parseIDFromURL(r *http.Request) (int, error) {
	idStr := r.PathValue("id")
	return strconv.Atoi(idStr)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
