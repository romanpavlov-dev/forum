package handlers

import "github.com/jackc/pgx/v5"

type Handler struct {
	conn *pgx.Conn
}

func NewHandler(conn *pgx.Conn) *Handler {
	return &Handler{
		conn: conn,
	}
}
