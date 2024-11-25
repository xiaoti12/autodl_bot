package storage

import (
	"autodl_bot/models"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type UserStorage struct {
	db *sql.DB
}

var DBPath = "users.db"

func NewUserStorage() (*UserStorage, error) {
	db, err := sql.Open("sqlite3", DBPath)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		telegram_id INTEGER PRIMARY KEY,
		username TEXT NOT NULL,
		password TEXT NOT NULL
	)`)

	if err != nil {
		db.Close()
		return nil, err
	}
	return &UserStorage{db: db}, nil
}

func (s *UserStorage) SaveUser(tgID int, username, password string) error {
	_, err := s.db.Exec(
		"INSERT OR REPLACE INTO users (telegram_id, username, password) VALUES (?, ?, ?)",
		tgID, username, password,
	)
	return err
}

func (s *UserStorage) LoadUser() (map[int]*models.AutoDLConfig, error) {
	rows, err := s.db.Query("SELECT telegram_id, username, password FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make(map[int]*models.AutoDLConfig)
	for rows.Next() {
		var tgID int
		var username, password string
		if err := rows.Scan(&tgID, &username, &password); err != nil {
			return nil, err
		}
		users[tgID] = &models.AutoDLConfig{
			Username: username,
			Password: password,
		}
	}
	return users, nil
}
