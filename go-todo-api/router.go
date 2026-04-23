package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter は chi.Router を作って、ルートを登録して返す。
// ハンドラを引数で受け取るのは、ルーターがハンドラに依存するが
// ハンドラはルーターを知らない、という一方向の依存を守るため。
func NewRouter(h *TodoHandler) *chi.Mux {
	r := chi.NewRouter()

	// chi 標準のミドルウェア
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// ルート定義
	r.Route("/todos", func(r chi.Router) {
		r.Get("/", h.ListTodos)
		r.Post("/", h.CreateTodo)
		r.Patch("/{id}/done", h.DoneTodo)
		r.Delete("/{id}", h.DeleteTodo)
	})

	return r
}
