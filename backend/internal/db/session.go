package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

const sessionDuration = 7 * 24 * time.Hour

type Session struct {
	Token     string    `json:"token"`
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func CreateSession(ctx context.Context, dbConn *sql.DB, userID int) (*Session, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	s := &Session{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(sessionDuration),
	}

	_, err = dbConn.ExecContext(ctx,
		`INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`,
		s.Token, s.UserID, s.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func GetSessionByToken(ctx context.Context, dbConn *sql.DB, token string) (*Session, error) {
	var s Session
	err := dbConn.QueryRowContext(ctx,
		`SELECT token, user_id, expires_at, created_at FROM sessions WHERE token = ? AND expires_at > CURRENT_TIMESTAMP`,
		token,
	).Scan(&s.Token, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func DeleteSession(ctx context.Context, dbConn *sql.DB, token string) error {
	_, err := dbConn.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
	return err
}

func DeleteExpiredSessions(ctx context.Context, dbConn *sql.DB) error {
	_, err := dbConn.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= CURRENT_TIMESTAMP`)
	return err
}
