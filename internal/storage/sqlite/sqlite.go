package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

var (
	ErrUrlNotFound = errors.New("url not found")
	ErrUrlExists   = errors.New("url exists")
)

type storageSqlite struct {
	db *sql.DB
}

func New(storagePath string) (*storageSqlite, error) {
	const op = "storage.sqlite.New"
	db, err := sql.Open("sqlite3", "./../../storage/url-shorten.db")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Создаем таблицу, если ее еще нет
	stmt, err := db.Prepare(`
    CREATE TABLE IF NOT EXISTS url(
        id INTEGER PRIMARY KEY,
        alias TEXT NOT NULL UNIQUE,
        url TEXT NOT NULL);
    CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
    `)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &storageSqlite{db: db}, nil
}

func (s *storageSqlite) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	// Подготавливаем запрос
	stmt, err := s.db.Prepare("INSERT INTO url(url,alias) values(?,?)")
	if err != nil {
		return 0, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	// Выполняем запрос
	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, ErrUrlExists)
		}

		return 0, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	// Получаем ID созданной записи
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	// Возвращаем ID
	return id, nil
}

func (s *storageSqlite) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resURL string

	err = stmt.QueryRow(alias).Scan(&resURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrUrlNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return resURL, nil
}

func (s *storageSqlite) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"
	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	_, err = stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return nil
}
