package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/user/user-server/pkg/models"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func New(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) Initialize() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS invite_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token TEXT UNIQUE NOT NULL,
			used BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			used_at DATETIME
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating invite_tokens table: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

func (db *DB) CreateUser(user *models.User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	result, err := db.Exec(
		"INSERT INTO users (username, password, created_at, updated_at) VALUES (?, ?, ?, ?)",
		user.Username, user.Password, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	return nil
}

func (db *DB) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := db.QueryRow(
		"SELECT id, username, password, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *DB) CreateInviteToken(token string) (*models.InviteToken, error) {
	now := time.Now()
	inviteToken := &models.InviteToken{
		Token:     token,
		Used:      false,
		CreatedAt: now,
	}

	result, err := db.Exec(
		"INSERT INTO invite_tokens (token, used, created_at) VALUES (?, ?, ?)",
		inviteToken.Token, inviteToken.Used, inviteToken.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	inviteToken.ID = id
	return inviteToken, nil
}

func (db *DB) GetInviteToken(token string) (*models.InviteToken, error) {
	inviteToken := &models.InviteToken{}
	var usedAt sql.NullTime

	err := db.QueryRow(
		"SELECT id, token, used, created_at, used_at FROM invite_tokens WHERE token = ?",
		token,
	).Scan(&inviteToken.ID, &inviteToken.Token, &inviteToken.Used, &inviteToken.CreatedAt, &usedAt)
	if err != nil {
		return nil, err
	}

	if usedAt.Valid {
		inviteToken.UsedAt = &usedAt.Time
	}

	return inviteToken, nil
}

func (db *DB) MarkInviteTokenAsUsed(tokenID int64) error {
	now := time.Now()
	_, err := db.Exec(
		"UPDATE invite_tokens SET used = 1, used_at = ? WHERE id = ?",
		now, tokenID,
	)
	return err
} 