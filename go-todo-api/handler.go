package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// TodoHandler はHTTPハンドラが使うTodoListへの参照を持つ構造体
type TodoHandler struct {
	todoList *TodoList
	filePath string
}

// NewTodoHandler はハンドラを初期化する（DI）
func NewTodoHandler(tl *TodoList, filePath string) *TodoHandler {
	return &TodoHandler{
		todoList: tl,
		filePath: filePath,
	}
}

// ============================================
// GET /todos - タスク一覧取得
// ============================================
func (h *TodoHandler) ListTodos(w http.ResponseWriter, r *http.Request) {
	todos := h.todoList.List()
	writeJSON(w, http.StatusOK, todos)
}

// ============================================
// POST /todos - タスク追加
// ============================================
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	// リクエストボディを受け取る構造体を定義（この関数内でしか使わない）
	var req struct {
		Title string `json:"title"`
	}

	// JSONボディを構造体にデコード
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "リクエストボディが不正です")
		return
	}

	// バリデーション
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "titleは必須です")
		return
	}

	// ビジネスロジックを呼び出す
	todo := h.todoList.Add(req.Title)

	// ファイルに保存
	if err := Save(h.todoList, h.filePath); err != nil {
		writeError(w, http.StatusInternalServerError, "保存に失敗しました")
		return
	}

	// 作成したTodoを返す（201 Created）
	writeJSON(w, http.StatusCreated, todo)
}

// ============================================
// PATCH /todos/{id}/done - タスク完了
// ============================================
func (h *TodoHandler) DoneTodo(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "IDが不正です")
		return
	}

	if err := h.todoList.Done(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if err := Save(h.todoList, h.filePath); err != nil {
		writeError(w, http.StatusInternalServerError, "保存に失敗しました")
		return
	}

	// 成功したがボディは返さない場合は 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// ============================================
// DELETE /todos/{id} - タスク削除
// ============================================
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "IDが不正です")
		return
	}

	if err := h.todoList.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if err := Save(h.todoList, h.filePath); err != nil {
		writeError(w, http.StatusInternalServerError, "保存に失敗しました")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================
// ヘルパー関数
// ============================================

// URLパラメータからIDを取り出して数値に変換する
// 例: /todos/1 → 1
func parseIDFromURL(r *http.Request) (int, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.Atoi(idStr)
}

// JSONレスポンスを書き込む共通関数
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// エラーレスポンスを書き込む共通関数
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
