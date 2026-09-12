package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"forum/internal/models"
	"log"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func RegisterUser(ctx context.Context, conn *pgx.Conn, user models.RegisterRequest) (pgconn.CommandTag, error) {

	if !IsValidEmail(user.Email) {
		log.Println("Not valid Email!")
		return pgconn.CommandTag{}, errors.New("invalid email")
	}

	query := `
	INSERT INTO users(username, email, password_hash, created_at)
	VALUES ($1, $2, $3, $4)
	`
	hash, err := PasswordHash(user.Password)
	if err != nil {
		log.Println(err)
		return pgconn.CommandTag{}, errors.New("Cant hash password")
	}
	return conn.Exec(ctx, query, user.Username, user.Email, hash, time.Now())

}

func AuthenticateUser(ctx context.Context, conn *pgx.Conn, user models.LoginRequest) (int, bool) {
	query := `
	SELECT password_hash, id
	FROM users
	WHERE email = $1
	`

	var hash string
	var userID int
	err := conn.QueryRow(ctx, query, user.Email).Scan(&hash, &userID)
	if err != nil {
		log.Println(err)
		return 0, false
	}

	if !VerifyPassword(user.Password, hash) {
		log.Println("Unauthorized: Wrong password")
		return 0, false
	}

	return userID, true
}

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func PasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Fatal("cant hash password")
		return "", err
	}

	return string(hash), nil
}

func VerifyPassword(password string, hash string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Println(err)
	}
	return err == nil
}

func GenerateToken() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func InsertSession(ctx context.Context, conn *pgx.Conn, userID int, access string, refresh string) error {
	query := `
	INSERT INTO sessions(user_id, access_token_hash, access_expires_at, refresh_token_hash, refresh_expires_at)
	VALUES ($1, $2, $3, $4, $5)
	`
	_, err := conn.Exec(ctx, query, userID, access, time.Now().Add(15*time.Minute), refresh, time.Now().Add(14*24*time.Hour))
	return err
}

func UpdateTokens(ctx context.Context, conn *pgx.Conn, userID int, access_hash string, new_refresh_token_hash string, old_refresh_token_hash string) bool {
	query := `
	UPDATE sessions
	SET access_token_hash = $1,
	access_expires_at = $2,
	refresh_token_hash = $3,
	refresh_expires_at = $4
	WHERE user_id = $5 AND refresh_token_hash = $6
	`

	result, err := conn.Exec(ctx, query, access_hash, time.Now().Add(15*time.Minute), new_refresh_token_hash, time.Now().Add(14*24*time.Hour), userID, old_refresh_token_hash)
	if err != nil {
		log.Println(err)
		return false
	}

	if result.RowsAffected() == 0 {
		log.Println("Cant find mathcing rows")
		return false
	}

	return true
}

func HashToken(token string) string { //разобраться как работает
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func StartSessionCleanup(conn *pgx.Conn) {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			if _, err := conn.Exec(context.Background(), `DELETE FROM sessions WHERE refresh_expires_at < NOW()`); err != nil {
				log.Println("Failed to delete session:", err)
			}
		}
	}()
}

func DeleteSession(ctx context.Context, conn *pgx.Conn, refresh_hash string) error {
	query := `
	DELETE FROM sessions
	WHERE refresh_token_hash = $1 `

	result, err := conn.Exec(ctx, query, refresh_hash)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		log.Println("No session matches")
	}

	return nil
}

func ValidateToken(ctx context.Context, conn *pgx.Conn, token string, token_type string) (int, bool) {

	var query string
	switch token_type {
	case "access":
		query = `
	SELECT user_id
	FROM sessions
	WHERE access_token_hash = $1 AND access_expires_at > $2`
	case "refresh":
		query = `
	SELECT user_id
	FROM sessions
	WHERE refresh_token_hash = $1 AND refresh_expires_at > $2`
	default:
		return 0, false
	}

	token_hash := HashToken(token)
	var userID int
	if err := conn.QueryRow(ctx, query, token_hash, time.Now()).Scan(&userID); err != nil {
		log.Println(err)
		return 0, false
	}

	return userID, true
}
