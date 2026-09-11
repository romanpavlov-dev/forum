package auth

import (
	"context"
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
	}
	return conn.Exec(ctx, query, user.Username, user.Email, hash, time.Now())

}

func AuthenticateUser(ctx context.Context, conn *pgx.Conn, user models.LoginRequest) bool {
	query := `
	SELECT password_hash
	FROM users
	WHERE email = $1
	`
	hash, err := conn.Exec(ctx, query, user.Email)
	if err != nil {
		log.Println(err)
		return false
	}

	return VerifyPassword(user.Password, hash.String())
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
