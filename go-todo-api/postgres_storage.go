package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// ============================================
// PostgresStorage - Storage interface のPostgreSQL実装
// ============================================
// FileStorage と同じ Storage interface を満たす。
// ハンドラから見れば FileStorage と区別がつかない（DIP）。
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage は PostgreSQL に接続して PostgresStorage を初期化する。
// 引数の connString は PostgreSQL への接続文字列。
// 例: "postgres://todouser:todopass@localhost:5432/tododb"
func NewPostgresStorage(connString string) (*PostgresStorage, error) {
	// sql.Open は「接続情報を保持するオブジェクト」を作るだけで
	// 実際の接続はまだ行われない。遅延接続。
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("DB接続オブジェクトの作成に失敗: %w", err)
	}

	// Ping で実際の接続を確認する。
	// 起動時に接続できなければ fail fast（早期に失敗）させるのが鉄則。
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("DBへの接続確認に失敗: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

// ============================================
// Storage interface の実装
// ============================================

// GetAll は全Todoを取得する
func (ps *PostgresStorage) GetAll() ([]Todo, error) {
	query := `SELECT id, title, done, created_at FROM todos ORDER BY id`

	rows, err := ps.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("SELECT失敗: %w", err)
	}
	defer rows.Close()

	// 結果を []Todo にマッピングする
	todos := []Todo{}
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("行の読み取り失敗: %w", err)
		}
		todos = append(todos, t)
	}

	// イテレーション中のエラーチェック（忘れやすいので注意）
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows走査中のエラー: %w", err)
	}

	return todos, nil
}

// Create は新しいTodoを作成する
// プリペアドステートメント（$1）でSQLインジェクション対策
func (ps *PostgresStorage) Create(title string) (Todo, error) {
	query := `
		INSERT INTO todos (title)
		VALUES ($1)
		RETURNING id, title, done, created_at
	`

	var t Todo
	// QueryRow + Scan で「INSERT後の値を取得」する
	// PostgreSQL の RETURNING 句が便利
	err := ps.db.QueryRow(query, title).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if err != nil {
		return Todo{}, fmt.Errorf("INSERT失敗: %w", err)
	}

	return t, nil
}

// Done は指定IDのTodoを完了にする
func (ps *PostgresStorage) Done(id int) error {
	// まずタスクが存在するか・既に完了済みかをチェックする
	var done bool
	err := ps.db.QueryRow(
		`SELECT done FROM todos WHERE id = $1`,
		id,
	).Scan(&done)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTodoNotFound
		}
		return fmt.Errorf("SELECT失敗: %w", err)
	}

	if done {
		return ErrAlreadyDone
	}

	// 完了にUPDATE
	_, err = ps.db.Exec(
		`UPDATE todos SET done = TRUE WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("UPDATE失敗: %w", err)
	}
	return nil
}

// Delete は指定IDのTodoを削除する
func (ps *PostgresStorage) Delete(id int) error {
	result, err := ps.db.Exec(
		`DELETE FROM todos WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("DELETE失敗: %w", err)
	}

	// 削除された行数で「存在したか」を判定する
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("RowsAffected取得失敗: %w", err)
	}

	if rowsAffected == 0 {
		return ErrTodoNotFound
	}
	return nil
}

// Close はDB接続を閉じる
// アプリ終了時に呼ぶ
func (ps *PostgresStorage) Close() error {
	return ps.db.Close()
}
