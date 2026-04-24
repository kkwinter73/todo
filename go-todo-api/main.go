package main

import (
	"log"
	"net/http"
)

const (
	serverAddr      = ":8080"
	defaultFilePath = "todos.json"
)

func main() {
	// 1. ストレージを初期化（ファイル実装）
	storage, err := NewFileStorage(defaultFilePath)
	if err != nil {
		log.Fatalf("ストレージの初期化に失敗しました: %v", err)
	}

	// 2. ハンドラを作成（Storage interface を渡す = 真のDI）
	handler := NewTodoHandler(storage)

	// 3. ルーターを作成
	router := NewRouter(handler)

	// 4. HTTPサーバーを起動
	log.Printf("🚀 サーバーを起動しました: http://localhost%s", serverAddr)
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}
}
