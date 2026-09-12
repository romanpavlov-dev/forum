package main

import (
	"context"
	"forum/db/connection"
	"forum/internal/auth"
	"forum/internal/handlers"
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

	sql, err = os.ReadFile("db/migrations/003_create_posts.up.sql")
	if err != nil {
		log.Fatal(err) //cant read file

	}

	_, err = conn.Exec(ctx, string(sql))
	if err != nil {
		log.Fatal(err) //cant run query
	}

	handler := handlers.NewHandler(conn)

	auth.StartSessionCleanup(conn)

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./web"))
	mux.Handle("GET /", fileServer)

	mux.HandleFunc("POST /register", handler.HandleRegistration)
	mux.HandleFunc("POST /login", handler.HandleLogin)
	mux.HandleFunc("POST /refresh", handler.HandleRefresh)

	mux.HandleFunc("POST /logout", handler.Middleware(handler.HandleLogout))

	mux.HandleFunc("GET /posts", handler.HandleGetAllPosts)
	mux.HandleFunc("GET /posts/{id}", handler.HandleGetPost)

	mux.HandleFunc("POST /posts", handler.Middleware(handler.HandleCreatePost))

	mux.HandleFunc("PATCH /posts/{id}", handler.Middleware(handler.HandleEditPost))

	mux.HandleFunc("DELETE /posts/{id}", handler.Middleware(handler.HandleDeletePost))

	log.Fatal(http.ListenAndServe(":9090", mux))
}
