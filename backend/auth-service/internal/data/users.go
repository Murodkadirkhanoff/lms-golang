package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"lms.chashma.uz/pkg/validator"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrDuplicateEmail = errors.New("duplicate email")
)

// User javobda frontend kutgan camelCase kalitlar bilan chiqadi
// (frontend/src/types/index.ts:3).
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Password  password  `json:"-"`
	Version   int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
}

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
	`

	args := []any{user.Name, user.Email, user.Password.hash, user.Role}

	err := m.DB.QueryRow(query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Constraint == "users_email_key" {
			return ErrDuplicateEmail
		}
		return err
	}

	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password_hash, role, version
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user User

	err := m.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Role,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) Get(id int64) (*User, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, name, email, password_hash, role, version
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user User

	err := m.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Role,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// GetByIDs internal endpoint uchun — boshqa servislar user ma'lumotini
// batch qilib oladi.
func (m UserModel) GetByIDs(ids []int64) ([]*User, error) {
	query := `
		SELECT id, created_at, name, email, role
		FROM users
		WHERE id = ANY($1) AND deleted_at IS NULL
	`

	rows, err := m.DB.Query(query, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}

	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.CreatedAt, &user.Name, &user.Email, &user.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, rows.Err()
}

// UpdatePassword parolni yangilaydi (password reset oqimi).
func (m UserModel) UpdatePassword(user *User) error {
	query := `
		UPDATE users
		SET password_hash = $1, version = version + 1
		WHERE id = $2
	`

	_, err := m.DB.Exec(query, user.Password.hash, user.ID)
	return err
}

// Count admin stats uchun jami (o'chirilmagan) foydalanuvchilar.
func (m UserModel) Count() (int, error) {
	var count int
	err := m.DB.QueryRow(`SELECT count(*) FROM users WHERE deleted_at IS NULL`).Scan(&count)
	return count, err
}

// List admin panel uchun sahifalangan foydalanuvchilar ro'yxati.
func (m UserModel) List(page, pageSize int) ([]*User, int, error) {
	query := `
		SELECT count(*) OVER(), id, created_at, name, email, role
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	rows, err := m.DB.Query(query, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	total := 0
	users := []*User{}

	for rows.Next() {
		var user User
		err := rows.Scan(&total, &user.ID, &user.CreatedAt, &user.Name, &user.Email, &user.Role)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	return users, total, rows.Err()
}
