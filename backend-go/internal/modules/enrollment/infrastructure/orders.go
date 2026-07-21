package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/chashma/lms/internal/modules/enrollment/application"
	"github.com/chashma/lms/internal/modules/enrollment/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrderRepository is the pgx-backed enrollment.OrderRepository.
type OrderRepository struct {
	pool *pgxpool.Pool
}

// NewOrderRepository builds an OrderRepository.
func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

var _ application.OrderRepository = (*OrderRepository)(nil)

// Insert writes an order and its items in one transaction (total is kept by a
// DB trigger; the caller already holds the computed total).
func (r *OrderRepository) Insert(ctx context.Context, o *domain.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := tx.QueryRow(ctx, `
		INSERT INTO enrollment.orders (user_id, status, payment_method)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`, o.UserID, o.Status, o.PaymentMethod).Scan(&o.DBID, &o.CreatedAt); err != nil {
		return err
	}
	for _, it := range o.Items {
		if _, err := tx.Exec(ctx, `
			INSERT INTO enrollment.order_items
			    (order_id, course_id, lesson_id, title, instructor_name, thumbnail_color, price)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			o.DBID, it.CourseID, it.LessonID, it.CourseTitle, it.Instructor, it.ThumbnailColor, it.Price); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func scanOrder(row pgx.Row) (domain.Order, error) {
	var o domain.Order
	err := row.Scan(&o.DBID, &o.CreatedAt, &o.UserID, &o.Total, &o.PaymentMethod, &o.Status)
	return o, err
}

// FindForUser returns an order (with items) owned by the user.
func (r *OrderRepository) FindForUser(ctx context.Context, id, userID int64) (*domain.Order, error) {
	if id < 1 {
		return nil, domain.ErrNotFound
	}
	o, err := scanOrder(r.pool.QueryRow(ctx, `
		SELECT id, created_at, user_id, total_amount::float8, payment_method, status
		FROM enrollment.orders
		WHERE id = $1 AND user_id = $2`, id, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	items, err := r.itemsForOrders(ctx, []int64{o.DBID})
	if err != nil {
		return nil, err
	}
	o.Items = items[o.DBID]
	return &o, nil
}

// ListByUser returns a page of a user's orders (with items) plus total count.
func (r *OrderRepository) ListByUser(ctx context.Context, userID int64, page, pageSize int) ([]domain.Order, int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT count(*) OVER() AS total, id, created_at, user_id, total_amount::float8, payment_method, status
		FROM enrollment.orders
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3`, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := []domain.Order{}
	total := 0
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&total, &o.DBID, &o.CreatedAt, &o.UserID, &o.Total, &o.PaymentMethod, &o.Status); err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if len(orders) == 0 {
		return orders, total, nil
	}

	ids := make([]int64, len(orders))
	for i, o := range orders {
		ids[i] = o.DBID
	}
	items, err := r.itemsForOrders(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	for i := range orders {
		orders[i].Items = items[orders[i].DBID]
	}
	return orders, total, nil
}

// Revenue returns the sum of all paid orders.
func (r *OrderRepository) Revenue(ctx context.Context) (float64, error) {
	var rev float64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(sum(total_amount), 0)::float8 FROM enrollment.orders WHERE status = 'paid'`).Scan(&rev)
	return rev, err
}

// RevenueForItems returns total revenue for the given courses/lessons and the
// last six months' breakdown (missing months filled with zero).
func (r *OrderRepository) RevenueForItems(ctx context.Context, courseIDs, lessonIDs []int64) (application.RevenueResult, error) {
	safeCourseIDs := courseIDs
	if len(safeCourseIDs) == 0 {
		safeCourseIDs = []int64{-1}
	}
	safeLessonIDs := lessonIDs
	if len(safeLessonIDs) == 0 {
		safeLessonIDs = []int64{-1}
	}

	rows, err := r.pool.Query(ctx, `
		SELECT to_char(date_trunc('month', o.created_at), 'YYYY-MM') AS month,
		       COALESCE(sum(oi.price), 0)::float8 AS revenue
		FROM enrollment.order_items oi
		JOIN enrollment.orders o ON o.id = oi.order_id
		WHERE o.status = 'paid'
		  AND (oi.course_id = ANY($1) OR oi.lesson_id = ANY($2))
		GROUP BY month`, safeCourseIDs, safeLessonIDs)
	if err != nil {
		return application.RevenueResult{}, err
	}
	defer rows.Close()

	byMonth := map[string]float64{}
	total := 0.0
	for rows.Next() {
		var month string
		var revenue float64
		if err := rows.Scan(&month, &revenue); err != nil {
			return application.RevenueResult{}, err
		}
		byMonth[month] = revenue
		total += revenue
	}
	if err := rows.Err(); err != nil {
		return application.RevenueResult{}, err
	}

	now := time.Now().UTC()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthly := make([]domain.MonthRevenue, 0, 6)
	for i := 5; i >= 0; i-- {
		month := firstOfMonth.AddDate(0, -i, 0).Format("2006-01")
		monthly = append(monthly, domain.MonthRevenue{Month: month, Revenue: byMonth[month]})
	}
	return application.RevenueResult{Total: total, Monthly: monthly}, nil
}

func (r *OrderRepository) itemsForOrders(ctx context.Context, orderIDs []int64) (map[int64][]domain.OrderItem, error) {
	byOrder := map[int64][]domain.OrderItem{}
	rows, err := r.pool.Query(ctx, `
		SELECT order_id, course_id, lesson_id, title, instructor_name, thumbnail_color, price::float8
		FROM enrollment.order_items
		WHERE order_id = ANY($1)
		ORDER BY id`, orderIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var orderID int64
		var it domain.OrderItem
		if err := rows.Scan(&orderID, &it.CourseID, &it.LessonID, &it.CourseTitle, &it.Instructor, &it.ThumbnailColor, &it.Price); err != nil {
			return nil, err
		}
		byOrder[orderID] = append(byOrder[orderID], it)
	}
	return byOrder, rows.Err()
}
