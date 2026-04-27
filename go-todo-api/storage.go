package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

// ============================================
// Storage interface - 永続化の契約
// ============================================
// 全メソッドが context.Context を第1引数に取る。
// HTTPリクエストのキャンセル・タイムアウトを下流まで伝えるため。
type Storage interface {
	GetAll(ctx context.Context) ([]Todo, error)
	Create(ctx context.Context, title string) (Todo, error)
	Done(ctx context.Context, id int) error
	Delete(ctx context.Context, id int) error
}

// ============================================
// sentinel error
// ============================================
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
// FileStorage はメモリ操作が中心なので ctx での中断は最小限にする。
// 入口で ctx.Err() をチェックすれば、既にキャンセルされたリクエストの
// 処理を始めずに済む。

func (fs *FileStorage) GetAll(ctx context.Context) ([]Todo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	result := make([]Todo, len(fs.todos))
	copy(result, fs.todos)
	return result, nil
}

func (fs *FileStorage) Create(ctx context.Context, title string) (Todo, error) {
	if err := ctx.Err(); err != nil {
		return Todo{}, err
	}

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

func (fs *FileStorage) Done(ctx context.Context, id int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

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

func (fs *FileStorage) Delete(ctx context.Context, id int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

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
