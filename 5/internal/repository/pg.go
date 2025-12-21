package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"gitlab.com/arkine/l3/5/internal/models"
)

var (
	ErrNoSeatsAvailable = errors.New("no seats available")
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func ConnectPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	return db, nil
}

func (r *Repository) ListEvents(ctx context.Context) ([]models.EventInfo, error) {
	query := `
		SELECT 
			e.id, e.title, e.date, e.total_seats,
			COUNT(b.id) as booked,
			(e.total_seats - COUNT(b.id)) as available
		FROM events e
		LEFT JOIN bookings b ON e.id = b.event_id
		GROUP BY e.id
		ORDER BY e.date;
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}(rows)

	var events []models.EventInfo
	for rows.Next() {
		var e models.EventInfo
		if err := rows.Scan(&e.ID, &e.Title, &e.Date, &e.TotalSeats, &e.Booked, &e.Available); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *Repository) CreateEvent(ctx context.Context, e *models.Event) error {
	query := `
		INSERT INTO events (title, date, total_seats, created_at)
		VALUES ($1, $2, $3, now())
		RETURNING id, created_at;
	`
	return r.DB.QueryRowContext(ctx, query, e.Title, e.Date, e.TotalSeats).
		Scan(&e.ID, &e.CreatedAt)
}

func (r *Repository) GetEventByID(ctx context.Context, id int) (*models.Event, error) {
	query := `
		SELECT id, title, date, total_seats, created_at
		FROM events
		WHERE id = $1;
	`
	var e models.Event
	if err := r.DB.QueryRowContext(ctx, query, id).
		Scan(&e.ID, &e.Title, &e.Date, &e.TotalSeats, &e.CreatedAt); err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) CountBookedSeats(ctx context.Context, eventID int) (int, error) {
	query := `SELECT COUNT(*) FROM bookings WHERE event_id = $1;`
	var count int
	if err := r.DB.QueryRowContext(ctx, query, eventID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) CreateBooking(ctx context.Context, eventID int, userName string) (*models.Booking, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}(tx)

	var total, booked int
	if err := tx.QueryRowContext(ctx, `SELECT total_seats FROM events WHERE id = $1;`, eventID).Scan(&total); err != nil {
		return nil, err
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM bookings WHERE event_id = $1;`, eventID).Scan(&booked); err != nil {
		return nil, err
	}

	if booked >= total {
		return nil, ErrNoSeatsAvailable
	}

	booking := &models.Booking{
		EventID:   eventID,
		UserName:  userName,
		Paid:      false,
		CreatedAt: time.Now(),
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO bookings (event_id, user_name, paid, created_at)
		VALUES ($1, $2, FALSE, now())
		RETURNING id, created_at;
	`, eventID, userName).Scan(&booking.ID, &booking.CreatedAt)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return booking, nil
}

func (r *Repository) MarkBookingPaid(ctx context.Context, bookingID int) error {
	query := `
		UPDATE bookings
		SET paid = TRUE, updated_at = now()
		WHERE id = $1;
	`
	_, err := r.DB.ExecContext(ctx, query, bookingID)
	return err
}

func (r *Repository) DeleteExpiredBookings(ctx context.Context, expirationTime time.Time) error {
	query := `
		DELETE FROM bookings
		WHERE paid = FALSE
		  AND created_at < $1;
	`
	res, err := r.DB.ExecContext(ctx, query, expirationTime)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows > 0 {
		fmt.Printf("[repository] удалено просроченных броней: %d\n", rows)
	}
	return nil
}
