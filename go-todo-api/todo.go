package main

import "time"

// ============================================
// Todo - 1件のタスクを表す構造体
// ============================================
// データの定義だけ。操作ロジックは storage.go の FileStorage 側に移動した。
type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}
