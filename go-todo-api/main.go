package main

import (
	"log"
	"net/http"
)

const serverAddr = ":8080"

func main() {
	// 1. 起動時にファイルからデータを読み込む
	todoList, err := Load(defaultFilePath)
	if err != nil {
		log.Fatalf("データの読み込みに失敗しました: %v", err)
	}

	// 2. ハンドラを作成（TodoList と ファイルパスを渡す = DI）
	handler := NewTodoHandler(todoList, defaultFilePath)

	// 3. ルーターを作成（ハンドラを渡す = DI）
	router := NewRouter(handler)

	// 4. HTTPサーバーを起動
	log.Printf("🚀 サーバーを起動しました: http://localhost%s", serverAddr)
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}
}
