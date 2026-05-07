package db

import (
	"context"
	"database/sql"
	"time"
)

const registrationTokenDuration = 7 * 24 * time.Hour

type RegistrationToken struct {
	Token     string    `json:"token"`
	CreatedBy int       `json:"created_by"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func UserCount(ctx context.Context, dbConn *sql.DB) (int, error) {
	var count int
	err := dbConn.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func CreateRegistrationToken(ctx context.Context, dbConn *sql.DB, createdBy int) (*RegistrationToken, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	t := &RegistrationToken{
		Token:     token,
		CreatedBy: createdBy,
		ExpiresAt: time.Now().Add(registrationTokenDuration),
	}

	_, err = dbConn.ExecContext(ctx,
		`INSERT INTO registration_tokens (token, created_by, expires_at) VALUES (?, ?, ?)`,
		t.Token, t.CreatedBy, t.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func ConsumeRegistrationToken(ctx context.Context, dbConn *sql.DB, token string) error {
	res, err := dbConn.ExecContext(ctx,
		`DELETE FROM registration_tokens WHERE token = ? AND expires_at > CURRENT_TIMESTAMP`,
		token,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func DeleteExpiredRegistrationTokens(ctx context.Context, dbConn *sql.DB) error {
	_, err := dbConn.ExecContext(ctx, `DELETE FROM registration_tokens WHERE expires_at <= CURRENT_TIMESTAMP`)
	return err
}

func ListRegistrationTokens(ctx context.Context, dbConn *sql.DB) ([]RegistrationToken, error) {
	rows, err := dbConn.QueryContext(ctx,
		`SELECT token, created_by, expires_at, created_at FROM registration_tokens ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RegistrationToken
	for rows.Next() {
		var t RegistrationToken
		if err := rows.Scan(&t.Token, &t.CreatedBy, &t.ExpiresAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
