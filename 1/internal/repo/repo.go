package repo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"gitlab.com/arkine/l3/1/internal/model"
)

type Repo struct {
	DB *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{DB: db}
}

func (r *Repo) CreateNotification(ctx context.Context, n *model.Notification) (int64, error) {
	var id int64
	err := r.DB.QueryRowContext(ctx,
		`INSERT INTO notifications (recipient, channel, message, send_at, status, attempts, last_error, created_at)
		 VALUES ($1,$2,$3,$4,$5,0,null,NOW()) RETURNING id`,
		n.Recipient, n.Channel, n.Message, n.SendAt, n.Status,
	).Scan(&id)
	if err != nil {
		log.Printf("CreateNotification failed: %v", err)
		return 0, fmt.Errorf("create notification: %w", err)
	}
	return id, nil
}

func (r *Repo) GetNotification(ctx context.Context, id int64) (*model.Notification, error) {
	var n model.Notification
	var lastErr sql.NullString

	err := r.DB.QueryRowContext(ctx,
		`SELECT id, recipient, channel, message, send_at, status, attempts, last_error, created_at
		 FROM notifications WHERE id=$1`, id).
		Scan(&n.ID, &n.Recipient, &n.Channel, &n.Message, &n.SendAt, &n.Status, &n.Attempts, &lastErr, &n.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("GetNotification failed for id=%d: %v", id, err)
		return nil, fmt.Errorf("get notification: %w", err)
	}

	if lastErr.Valid {
		n.LastError = &lastErr.String
	}

	return &n, nil
}

func (r *Repo) UpdateStatus(ctx context.Context, id int64, status model.Status) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE notifications SET status=$1 WHERE id=$2`, status, id)
	if err != nil {
		log.Printf("UpdateStatus failed for id=%d: %v", id, err)
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

func (r *Repo) UpdateAttemptsAndError(ctx context.Context, id int64, attempts int, lastErr string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE notifications SET attempts=$1, last_error=$2 WHERE id=$3`, attempts, lastErr, id)
	if err != nil {
		log.Printf("UpdateAttemptsAndError failed for id=%d: %v", id, err)
		return fmt.Errorf("update attempts/error: %w", err)
	}
	return nil
}

func (r *Repo) ListRecent(ctx context.Context, limit int) ([]model.Notification, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, recipient, channel, message, send_at, status, attempts, last_error, created_at
		 FROM notifications ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		log.Printf("ListRecent query failed: %v", err)
		return nil, fmt.Errorf("list recent: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			log.Printf("ListRecent rows close error: %v", cerr)
		}
	}()

	var out []model.Notification
	for rows.Next() {
		var n model.Notification
		var lastErr sql.NullString
		if err := rows.Scan(&n.ID, &n.Recipient, &n.Channel, &n.Message, &n.SendAt, &n.Status, &n.Attempts, &lastErr, &n.CreatedAt); err != nil {
			log.Printf("ListRecent scan error: %v", err)
			return nil, fmt.Errorf("list recent scan: %w", err)
		}
		if lastErr.Valid {
			n.LastError = &lastErr.String
		}
		out = append(out, n)
	}

	if err := rows.Err(); err != nil {
		log.Printf("ListRecent rows error: %v", err)
		return nil, fmt.Errorf("list recent rows: %w", err)
	}

	return out, nil
}

func (r *Repo) UpdateNotification(ctx context.Context, id int64, msg string, sendAt time.Time) error {
	_, err := r.DB.ExecContext(ctx, `
        UPDATE notifications
        SET message=$1, send_at=$2, status='scheduled'
        WHERE id=$3
    `, msg, sendAt, id)
	return err
}

func (r *Repo) CancelNotification(ctx context.Context, id int64) error {
	res, err := r.DB.ExecContext(ctx, `UPDATE notifications SET status='cancelled' WHERE id=$1`, id)
	if err != nil {
		log.Printf("CancelNotification failed for id=%d: %v", id, err)
		return fmt.Errorf("cancel notification: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		log.Printf("CancelNotification RowsAffected failed for id=%d: %v", id, err)
		return fmt.Errorf("rows affected: %w", err)
	}

	if affected == 0 {
		log.Printf("CancelNotification: no rows affected for id=%d", id)
		return fmt.Errorf("notification not found")
	}

	return nil
}
