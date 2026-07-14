package data

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
)

// OrderItem JSON shakli frontend OrderItem tipiga mos (types/index.ts:136).
type OrderItem struct {
	CourseID       *int64  `json:"-"`
	LessonID       *int64  `json:"-"`
	CourseTitle    string  `json:"courseTitle"`
	Instructor     string  `json:"instructor"`
	ThumbnailColor string  `json:"thumbnailColor"`
	Price          float64 `json:"price"`
}

// Order JSON shakli frontend Order tipiga mos: id string, date, status
// (completed|refunded|pending), paymentMethod, items, total.
type Order struct {
	ID            int64        `json:"-"`
	CreatedAt     time.Time    `json:"-"`
	UserID        int64        `json:"-"`
	Status        string       `json:"-"`
	PaymentMethod string       `json:"paymentMethod"`
	Items         []*OrderItem `json:"items"`
	Total         float64      `json:"total"`

	// Frontend uchun hisoblangan maydonlar (MarshalJSON pastda emas —
	// oddiy ko'rinish maydonlari handler to'ldiradi).
	PublicID     string `json:"id"`
	Date         string `json:"date"`
	PublicStatus string `json:"status"`
}

// Finalize DB maydonlaridan frontend kutgan ko'rinish maydonlarini to'ldiradi.
func (o *Order) Finalize() {
	o.PublicID = strconv.FormatInt(o.ID, 10)
	o.Date = o.CreatedAt.Format("2006-01-02")

	switch o.Status {
	case "paid":
		o.PublicStatus = "completed"
	case "failed", "cancelled":
		o.PublicStatus = "refunded"
	default:
		o.PublicStatus = "pending"
	}
}

type OrderModel struct {
	DB *sql.DB
}

// Insert buyurtma va itemlarni bitta tranzaksiyada yozadi.
// total_amount'ni DB trigger hisoblaydi.
func (m OrderModel) Insert(order *Order) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		`INSERT INTO orders (user_id, status, payment_method) VALUES ($1, $2, $3) RETURNING id, created_at`,
		order.UserID, order.Status, order.PaymentMethod,
	).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return err
	}

	for _, item := range order.Items {
		_, err = tx.Exec(
			`INSERT INTO order_items (order_id, course_id, lesson_id, title, instructor_name, thumbnail_color, price)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			order.ID, item.CourseID, item.LessonID, item.CourseTitle, item.Instructor, item.ThumbnailColor, item.Price,
		)
		if err != nil {
			return err
		}
		order.Total += item.Price
	}

	return tx.Commit()
}

func (m OrderModel) GetForUser(id, userID int64) (*Order, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, user_id, total_amount, payment_method, status
		FROM orders
		WHERE id = $1 AND user_id = $2
	`

	var order Order

	err := m.DB.QueryRow(query, id, userID).Scan(
		&order.ID,
		&order.CreatedAt,
		&order.UserID,
		&order.Total,
		&order.PaymentMethod,
		&order.Status,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	order.Items, err = m.itemsForOrders([]int64{order.ID})
	if err != nil {
		return nil, err
	}

	order.Finalize()

	return &order, nil
}

func (m OrderModel) ListByUser(userID int64) ([]*Order, error) {
	query := `
		SELECT id, created_at, user_id, total_amount, payment_method, status
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
	`

	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []*Order{}
	byID := map[int64]*Order{}
	ids := []int64{}

	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.ID,
			&order.CreatedAt,
			&order.UserID,
			&order.Total,
			&order.PaymentMethod,
			&order.Status,
		)
		if err != nil {
			return nil, err
		}
		order.Items = []*OrderItem{}
		order.Finalize()
		orders = append(orders, &order)
		byID[order.ID] = &order
		ids = append(ids, order.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return orders, nil
	}

	items, orderIDs, err := m.itemsWithOrderIDs(ids)
	if err != nil {
		return nil, err
	}

	for i, item := range items {
		if order, ok := byID[orderIDs[i]]; ok {
			order.Items = append(order.Items, item)
		}
	}

	return orders, nil
}

// Revenue barcha to'langan buyurtmalar summasi (admin stats).
func (m OrderModel) Revenue() (float64, error) {
	var revenue float64
	err := m.DB.QueryRow(
		`SELECT COALESCE(sum(total_amount), 0) FROM orders WHERE status = 'paid'`,
	).Scan(&revenue)
	return revenue, err
}

func (m OrderModel) itemsForOrders(orderIDs []int64) ([]*OrderItem, error) {
	items, _, err := m.itemsWithOrderIDs(orderIDs)
	return items, err
}

func (m OrderModel) itemsWithOrderIDs(orderIDs []int64) ([]*OrderItem, []int64, error) {
	query := `
		SELECT order_id, course_id, lesson_id, title, instructor_name, thumbnail_color, price
		FROM order_items
		WHERE order_id = ANY($1)
		ORDER BY id
	`

	rows, err := m.DB.Query(query, int64Array(orderIDs))
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	items := []*OrderItem{}
	ids := []int64{}

	for rows.Next() {
		var item OrderItem
		var orderID int64
		err := rows.Scan(&orderID, &item.CourseID, &item.LessonID, &item.CourseTitle, &item.Instructor, &item.ThumbnailColor, &item.Price)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, &item)
		ids = append(ids, orderID)
	}

	return items, ids, rows.Err()
}
