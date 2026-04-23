// TodoListの保存と読み込みを担当する
package main

import (
	"encoding/json"
	"errors"
	"os"
)

const defaultFilePath = "todos.json"

// ToListをjson変換してファイルに保存する
func Save(tl *TodoList, filePath string) error {
	// ロック取得中の短時間でスナップショットを取り、それをJSON化する
	snapshot := tl.Snapshot()
	data, err := json.MarshalIndent(&snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// jsonファイル読み込みして返す
// そのために必要なものは？　ファイルパス
func Load(filePath string) (*TodoList, error) {

	// ファイル読み込み
	data, err := os.ReadFile(filePath)

	// ファイルが存在しないかどうかを判定。ファイルが無い場合→初回起動と判定
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewTodoList(), nil
		}

		return nil, err
	}

	// 空のTodoList変数を宣言
	var tl TodoList

	// jsonテキストをToListの構造体に変換
	if err := json.Unmarshal(data, &tl); err != nil {
		return nil, err
	}

	// 変換した構造体のポインタを返す（　呼び出し元がこのデータを使う）
	return &tl, nil
}
