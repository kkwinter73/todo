package main

import (
	"log"
	"net/http"
	"time"
)

// NewRouter は標準 net/http の ServeMux でルーティングを構築する。
// Go 1.22 から ServeMux がメソッド指定とパスパラメータをサポートしたため、
// 外部ルーターライブラリを使わずに済む。
func NewRouter(h *TodoHandler) http.Handler {
	mux := http.NewServeMux()

	// メソッドとパスを同時に指定できる (Go 1.22+)
	// パスパラメータは {id} で書く
	mux.HandleFunc("GET /todos", h.ListTodos)
	mux.HandleFunc("POST /todos", h.CreateTodo)
	mux.HandleFunc("PATCH /todos/{id}/done", h.DoneTodo)
	mux.HandleFunc("DELETE /todos/{id}", h.DeleteTodo)

	// ミドルウェアは関数合成で適用する。
	// 適用順は外側から内側へ: logger(recoverer(mux))
	return loggerMiddleware(recovererMiddleware(mux))
}

// loggerMiddleware は各リクエストのメソッド・パス・処理時間をログ出力する。
// chi の middleware.Logger と同等の最小実装。
func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// recovererMiddleware はハンドラ内で発生した panic を捕捉し、
// 500 を返してプロセスを継続させる。
// chi の middleware.Recoverer と同等の最小実装。
func recovererMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
