package main

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

// ============================================
// Storage interface - 永続化の契約
// ============================================
// この interface を満たせば、ファイルでもDBでもメモリでも使える。
// ハンドラはこの interface にだけ依存する。
type Storage interface {
	GetAll() ([]Todo, error)
	Create(title string) (Todo, error)
	Done(id int) error
	Delete(id int) error
}

// ============================================
// sentinel error - エラーの種類を識別する
// ============================================
// errors.Is() で判定できるように、パッケージ変数として定義する。
// 文字列比較でエラー判別すると、多言語化した瞬間に壊れるのでNG。
var (
	ErrTodoNotFound = errors.New("todo not found")
	ErrAlreadyDone  = errors.New("todo already done")
)

// ============================================
// FileStorage - Storage interface のファイル実装
// ============================================
// ファイル版の永続化ロジック。Storage interface を満たすために
// GetAll / Create / Done / Delete を実装する。
type FileStorage struct {
	mu       sync.Mutex
	filePath string
	todos    []Todo
	nextID   int
}

// NewFileStorage は FileStorage を初期化する。
// 起動時にファイルを読み込んでメモリに乗せる。
func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath: filePath,
		todos:    []Todo{},
		nextID:   1,
	}

	if err := fs.load(); err != nil {
		return nil, err
	}
	return fs, nil
}

// load はファイルからデータを読み込む（内部用）
func (fs *FileStorage) load() error {
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // ファイルが無い = 初回起動。そのまま空状態
		}
		return err
	}

	var state struct {
		Todos  []Todo `json:"todos"`
		NextID int    `json:"next_id"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	fs.todos = state.Todos
	if state.NextID > 0 {
		fs.nextID = state.NextID
	}
	return nil
}

// save は現在のメモリ内容をファイルに保存する（内部用）
// アトミックな書き込み：一時ファイルに書いてから rename する
func (fs *FileStorage) save() error {
	state := struct {
		Todos  []Todo `json:"todos"`
		NextID int    `json:"next_id"`
	}{
		Todos:  fs.todos,
		NextID: fs.nextID,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := fs.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, fs.filePath)
}

// ============================================
// Storage interface の実装
// ============================================

// GetAll は全Todoを返す（防御的コピー）
func (fs *FileStorage) GetAll() ([]Todo, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	result := make([]Todo, len(fs.todos))
	copy(result, fs.todos)
	return result, nil
}

// Create は新しいTodoを作成する
func (fs *FileStorage) Create(title string) (Todo, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	todo := Todo{
		ID:        fs.nextID,
		Title:     title,
		Done:      false,
		CreatedAt: time.Now(),
	}
	fs.todos = append(fs.todos, todo)
	fs.nextID++

	if err := fs.save(); err != nil {
		return Todo{}, err
	}
	return todo, nil
}

// Done は指定IDのTodoを完了にする
func (fs *FileStorage) Done(id int) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for i := range fs.todos {
		if fs.todos[i].ID == id {
			if fs.todos[i].Done {
				return ErrAlreadyDone
			}
			fs.todos[i].Done = true
			return fs.save()
		}
	}
	return ErrTodoNotFound
}

// Delete は指定IDのTodoを削除する
func (fs *FileStorage) Delete(id int) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for i := range fs.todos {
		if fs.todos[i].ID == id {
			fs.todos = append(fs.todos[:i], fs.todos[i+1:]...)
			return fs.save()
		}
	}
	return ErrTodoNotFound
}
