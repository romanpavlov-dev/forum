package main

import (
	"context"
	"forum/db/connection"
	"forum/internal"
	"forum/internal/auth"
	"log"
	"net/http"
	"os"
)

func main() {

	ctx := context.Background()

	conn := connection.CreateConn(ctx)
	defer conn.Close(ctx)

	sql, err := os.ReadFile("db/migrations/001_create_users.up.sql")
	if err != nil {
		log.Fatal(err) //cant read file

	}

	_, err = conn.Exec(ctx, string(sql))
	if err != nil {
		log.Fatal(err) //cant run query
	}

	sql, err = os.ReadFile("db/migrations/002_create_auth_sessions.up.sql")
	if err != nil {
		log.Fatal(err) //cant read file

	}

	_, err = conn.Exec(ctx, string(sql))
	if err != nil {
		log.Fatal(err) //cant run query
	}

	handler := internal.NewHandler(conn)

	auth.StartSessionCleanup(conn)

	http.HandleFunc("/register", handler.HandleRegistration)
	http.HandleFunc("/login", handler.HandleLogin)
	http.HandleFunc("/logout", handler.Middleware(handler.HandleLogout))
	http.HandleFunc("/", internal.MainHandler)
	http.HandleFunc("/refresh", handler.HandleRefresh)

	log.Fatal(http.ListenAndServe(":9090", nil))
}
