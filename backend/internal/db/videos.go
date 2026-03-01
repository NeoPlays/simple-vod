package db

import (
	"context"
	"database/sql"
	"errors"
)

type Video struct {
	ID int `json:"id"`
	Name string `json:"name"`
}

func CreateVideo(ctx context.Context, dbConn *sql.DB, v *Video) error {
	res, err := dbConn.ExecContext(
		ctx,
		`INSERT INTO videos (name) VALUES (?)`,
		v.Name,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	v.ID = int(id)
	return nil
}


func CreateVideoIfMissing(ctx context.Context, dbConn *sql.DB, name string) (created bool, err error) {
	res, err := dbConn.ExecContext(ctx, `INSERT OR IGNORE INTO videos (name) VALUES (?)`, name)
	if err != nil {
		return false, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func GetVideoByID(ctx context.Context, dbConn *sql.DB, id int) (*Video, error) {
	var v Video
	err := dbConn.QueryRowContext(
		ctx,
		`SELECT id, name FROM videos WHERE id = ?`,
		id,
	).Scan(&v.ID, &v.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

func ListVideos(ctx context.Context, dbConn *sql.DB) ([]Video, error) {
	rows, err := dbConn.QueryContext(ctx, `SELECT id, name FROM videos ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Video
	for rows.Next() {
		var v Video
		if err := rows.Scan(&v.ID, &v.Name); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func UpdateVideo(ctx context.Context, dbConn *sql.DB, v Video) error {
	res, err := dbConn.ExecContext(
		ctx,
		`UPDATE videos SET name = ? WHERE id = ?`,
		v.Name,
		v.ID,
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

func DeleteVideo(ctx context.Context, dbConn *sql.DB, id int) error {
	res, err := dbConn.ExecContext(ctx, `DELETE FROM videos WHERE id = ?`, id)
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
