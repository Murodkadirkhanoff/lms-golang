package data

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"errors"
	"time"
)

var ErrInvalidToken = errors.New("invalid or expired token")

// Token — parolni tiklash tokeni (Greenlight uslubi): DB'da faqat sha256
// hash saqlanadi, plaintext foydalanuvchiga yuboriladi.
type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
}

type TokenModel struct {
	DB *sql.DB
}

func (m TokenModel) New(userID int64, ttl time.Duration) (*Token, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token := &Token{
		Plaintext: base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes),
		UserID:    userID,
		Expiry:    time.Now().Add(ttl),
	}

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	query := `
		INSERT INTO password_reset_tokens (hash, user_id, expiry)
		VALUES ($1, $2, $3)
	`

	_, err = m.DB.Exec(query, token.Hash, token.UserID, token.Expiry)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// UserIDForToken tokenni tekshirib, egasining id'sini qaytaradi.
func (m TokenModel) UserIDForToken(plaintext string) (int64, error) {
	hash := sha256.Sum256([]byte(plaintext))

	query := `
		SELECT user_id
		FROM password_reset_tokens
		WHERE hash = $1 AND expiry > NOW()
	`

	var userID int64

	err := m.DB.QueryRow(query, hash[:]).Scan(&userID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, ErrInvalidToken
		default:
			return 0, err
		}
	}

	return userID, nil
}

func (m TokenModel) DeleteAllForUser(userID int64) error {
	_, err := m.DB.Exec(`DELETE FROM password_reset_tokens WHERE user_id = $1`, userID)
	return err
}
