package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"subscriptions-api/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, s *model.Subscription) (uuid.UUID, error) {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	row := r.pool.QueryRow(ctx, query, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate)
	if err := row.Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return uuid.Nil, err
	}
	return s.ID, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	var s model.Subscription
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE id=$1`
	row := r.pool.QueryRow(ctx, query, id)
	if err := row.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, s *model.Subscription) (bool, error) {
	query := `
		UPDATE subscriptions
		SET service_name=$1, price=$2, user_id=$3, start_date=$4, end_date=$5, updated_at=now()
		WHERE id=$6
	`
	ct, err := r.pool.Exec(ctx, query, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate, id)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() == 1, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM subscriptions WHERE id=$1`, id)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() == 1, nil
}

type ListFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	Limit       int
	Offset      int
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]model.Subscription, error) {
	q := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions WHERE 1=1`
	args := []any{}
	idx := 1

	if f.UserID != nil {
		q += " AND user_id=$" + itoa(idx)
		args = append(args, *f.UserID)
		idx++
	}
	if f.ServiceName != nil {
		q += " AND service_name=$" + itoa(idx)
		args = append(args, *f.ServiceName)
		idx++
	}
	q += " ORDER BY created_at DESC"
	if f.Limit > 0 {
		q += " LIMIT $" + itoa(idx)
		args = append(args, f.Limit)
		idx++
	}
	if f.Offset > 0 {
		q += " OFFSET $" + itoa(idx)
		args = append(args, f.Offset)
		idx++
	}

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []model.Subscription
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, rows.Err()
}

// Summary: сумма price * кол-во активных месяцев в интервале [from,to]
func (r *Repository) Summary(ctx context.Context, from, to time.Time, userID *uuid.UUID, serviceName *string) (int64, error) {
	q := `
SELECT
 COALESCE(SUM(
  price * (
   (DATE_PART('year', ato) - DATE_PART('year', afrom)) * 12
   + (DATE_PART('month', ato) - DATE_PART('month', afrom)) + 1
  )
 ), 0) AS total
FROM (
 SELECT
  s.price,
  GREATEST(date_trunc('month', $1::date), date_trunc('month', s.start_date)) AS afrom,
  LEAST(date_trunc('month', COALESCE(s.end_date, $2::date)), date_trunc('month', $2::date)) AS ato
 FROM subscriptions s
 WHERE s.start_date <= $2::date
   AND COALESCE(s.end_date, '9999-12-31') >= $1::date
   %s
) t;
`
	// filters
	filter := ""
	args := []any{from, to}
	idx := 3

	if userID != nil {
		filter += " AND s.user_id=$" + itoa(idx)
		args = append(args, *userID)
		idx++
	}
	if serviceName != nil {
		filter += " AND s.service_name=$" + itoa(idx)
		args = append(args, *serviceName)
		idx++
	}

	query := sprintf(q, filter)
	var total int64
	err := r.pool.QueryRow(ctx, query, args...).Scan(&total)
	return total, err
}

// helpers
func itoa(i int) string                 { return fmt.Sprintf("%d", i) }
func sprintf(f string, a ...any) string { return fmt.Sprintf(f, a...) }
