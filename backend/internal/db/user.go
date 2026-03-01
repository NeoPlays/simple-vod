package db

import (
	"context"
	"database/sql"
	"errors"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
	PasswordHash string `json:"password_hash"`
	Role Role `json:"role"`
}

func CreateUser(ctx context.Context, dbConn *sql.DB, u *User) error {
	res, err := dbConn.ExecContext(
		ctx,
		`INSERT INTO users (name, password_hash, role) VALUES (?, ?, ?)`,
		u.Name,
		u.PasswordHash,
		u.Role,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = int(id)
	return nil
}

func GetUserByID(ctx context.Context, dbConn *sql.DB, id int) (*User, error) {
	var u User
	err := dbConn.QueryRowContext(
		ctx,
		`SELECT id, name, password_hash, role FROM users WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Name, &u.PasswordHash, &u.Role)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func GetUserByName(ctx context.Context, dbConn *sql.DB, name string) (*User, error) {
	var u User
	err := dbConn.QueryRowContext(
		ctx,
		`SELECT id, name, password_hash, role FROM users WHERE name = ?`,
		name,
	).Scan(&u.ID, &u.Name, &u.PasswordHash, &u.Role)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func ListUsers(ctx context.Context, dbConn *sql.DB) ([]User, error) {
	rows, err := dbConn.QueryContext(ctx, `SELECT id, name, password_hash, role FROM users ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.PasswordHash, &u.Role); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func UpdateUser(ctx context.Context, dbConn *sql.DB, u User) error {
	res, err := dbConn.ExecContext(
		ctx,
		`UPDATE users SET name = ?, password_hash = ?, role = ? WHERE id = ?`,
		u.Name,
		u.PasswordHash,
		u.Role,
		u.ID,
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

func DeleteUser(ctx context.Context, dbConn *sql.DB, id int) error {
	res, err := dbConn.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
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

		