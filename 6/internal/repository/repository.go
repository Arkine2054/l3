package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"gitlab.com/arkine/l3/6/internal/models"
)

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{DB: db}
}

func (r *Repo) CreateSale(ctx context.Context, s *models.Sale) error {
	q := `INSERT INTO sales (kind, amount, category, note, created_at)
          VALUES ($1, $2, $3, $4, $5)
          RETURNING id, created_at`
	err := r.DB.QueryRowContext(ctx, q, s.Kind, s.Amount, s.Category, s.Note, s.CreatedAt).Scan(&s.ID, &s.CreatedAt)
	if err != nil {
		return fmt.Errorf("create sale: %w", err)
	}
	return nil
}

type ListFilter struct {
	From     *time.Time
	To       *time.Time
	Category *string
	Kind     *models.Kind
	Limit    int
	Offset   int
	SortBy   string
	Desc     bool
}

var allowedSort = map[string]bool{
	"created_at": true,
	"amount":     true,
	"category":   true,
}

func (r *Repo) ListSales(ctx context.Context, f ListFilter) ([]models.Sale, error) {
	sort := "created_at"
	if f.SortBy != "" && allowedSort[f.SortBy] {
		sort = f.SortBy
	}
	order := "ASC"
	if f.Desc {
		order = "DESC"
	}

	q := `SELECT id, kind, amount, category, note, created_at FROM sales WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if f.From != nil {
		q += fmt.Sprintf(" AND created_at >= $%d", idx)
		args = append(args, *f.From)
		idx++
	}
	if f.To != nil {
		q += fmt.Sprintf(" AND created_at <= $%d", idx)
		args = append(args, *f.To)
		idx++
	}
	if f.Category != nil {
		q += fmt.Sprintf(" AND category = $%d", idx)
		args = append(args, *f.Category)
		idx++
	}
	if f.Kind != nil {
		q += fmt.Sprintf(" AND kind = $%d", idx)
		args = append(args, *f.Kind)
		idx++
	}

	q += fmt.Sprintf(" ORDER BY %s %s", sort, order)

	if f.Limit > 0 {
		q += fmt.Sprintf(" LIMIT $%d", idx)
		args = append(args, f.Limit)
		idx++
	}
	if f.Offset > 0 {
		q += fmt.Sprintf(" OFFSET $%d", idx)
		args = append(args, f.Offset)
		idx++
	}

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list sales: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}(rows)

	res := []models.Sale{}
	for rows.Next() {
		var s models.Sale
		if err := rows.Scan(&s.ID, &s.Kind, &s.Amount, &s.Category, &s.Note, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan sale: %w", err)
		}
		res = append(res, s)
	}
	return res, nil
}

func (r *Repo) GetSale(ctx context.Context, id int64) (*models.Sale, error) {
	q := `SELECT id, kind, amount, category, note, created_at FROM sales WHERE id = $1`
	var s models.Sale
	err := r.DB.QueryRowContext(ctx, q, id).Scan(&s.ID, &s.Kind, &s.Amount, &s.Category, &s.Note, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get sale: %w", err)
	}
	return &s, nil
}

func (r *Repo) UpdateSale(ctx context.Context, s *models.Sale) error {
	q := `UPDATE sales SET kind=$1, amount=$2, category=$3, note=$4 WHERE id=$5`
	res, err := r.DB.ExecContext(ctx, q, s.Kind, s.Amount, s.Category, s.Note, s.ID)
	if err != nil {
		return fmt.Errorf("update sale: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repo) DeleteSale(ctx context.Context, id int64) error {
	q := `DELETE FROM sales WHERE id = $1`
	res, err := r.DB.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete sale: %w", err)
	}
	if ra, _ := res.RowsAffected(); ra == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repo) GetAnalytics(ctx context.Context, from, to *time.Time, kind *models.Kind, category *string) (models.Analytics, error) {
	var a models.Analytics
	a.From = from
	a.To = to

	where := "WHERE 1=1"
	args := []interface{}{}
	idx := 1
	if from != nil {
		where += fmt.Sprintf(" AND created_at >= $%d", idx)
		args = append(args, *from)
		idx++
	}
	if to != nil {
		where += fmt.Sprintf(" AND created_at <= $%d", idx)
		args = append(args, *to)
		idx++
	}
	if kind != nil {
		where += fmt.Sprintf(" AND kind = $%d", idx)
		args = append(args, *kind)
		idx++
	}
	if category != nil {
		where += fmt.Sprintf(" AND category = $%d", idx)
		args = append(args, *category)
		idx++
	}

	q1 := fmt.Sprintf(`SELECT COUNT(*)::bigint, COALESCE(SUM(amount),0)::numeric::float8, COALESCE(AVG(amount),0)::numeric::float8
        FROM sales %s`, where)
	row := r.DB.QueryRowContext(ctx, q1, args...)
	var cnt sql.NullInt64
	var sum, avg sql.NullFloat64
	if err := row.Scan(&cnt, &sum, &avg); err != nil {
		return a, fmt.Errorf("analytics basic: %w", err)
	}
	if cnt.Valid {
		a.Count = cnt.Int64
	}
	if sum.Valid {
		a.Sum = sum.Float64
	}
	if avg.Valid {
		a.Avg = avg.Float64
	}

	q2 := fmt.Sprintf(`SELECT
           COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY amount)::numeric::float8, 0),
           COALESCE(percentile_cont(0.9) WITHIN GROUP (ORDER BY amount)::numeric::float8, 0)
        FROM sales %s`, where)
	row = r.DB.QueryRowContext(ctx, q2, args...)
	var median, p90 sql.NullFloat64
	if err := row.Scan(&median, &p90); err != nil {
		return a, fmt.Errorf("analytics percentiles: %w", err)
	}
	if median.Valid {
		a.Median = median.Float64
	}
	if p90.Valid {
		a.Percentile90 = p90.Float64
	}
	return a, nil
}
