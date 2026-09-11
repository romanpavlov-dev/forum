package connection

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func CreateConn(ctx context.Context) *pgx.Conn {
	conn, err := pgx.Connect(ctx, "postgres://postgres:7658@localhost:5432/postgres")
	if err != nil {
		log.Fatal(err, "Could not connect database!")
		return nil
	}
	return conn
}
