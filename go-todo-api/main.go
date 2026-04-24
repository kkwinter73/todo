package main

import (
	"log"
	"net/http"
)

const (
	serverAddr = ":8080"

	// PostgreSQL 接続文字列
	// 書式: postgres://ユーザー:パスワード@ホスト:ポート/DB名?オプション
	connString = "postgres://todouser:todopassword@localhost:5432/tododb?sslmode=disable"
)

func main() {
	// 1. ストレージを初期化（PostgreSQL実装に変更）
	storage, err := NewPostgresStorage(connString)
	if err != nil {
		log.Fatalf("ストレージの初期化に失敗しました: %v", err)
	}
	defer storage.Close()

	// 2. ハンドラを作成（interfaceとして渡す＝コード変更なし）
	handler := NewTodoHandler(storage)

	// 3. ルーターを作成
	router := NewRouter(handler)

	// 4. HTTPサーバーを起動
	log.Printf("🚀 サーバーを起動しました: http://localhost%s", serverAddr)
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}
}
