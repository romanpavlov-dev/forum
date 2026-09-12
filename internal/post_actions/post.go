package post_actions

import (
	"context"
	"forum/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
)

func InsertPost(ctx context.Context, conn *pgx.Conn, userID int, post models.PostRequest) error {
	query := `
	INSERT INTO posts(user_id, title, content, created_at)
	VALUES($1, $2, $3, $4)
	`
	_, err := conn.Exec(ctx, query, userID, post.Title, post.Content, time.Now())

	return err
}

func EditPost(ctx context.Context, conn *pgx.Conn, userID int, postID int, post models.PostRequest) error {
	query := `
	UPDATE posts
	SET title = $1, content = $2, updated_at = $3
	WHERE user_id = $4 AND id = $5
	`
	_, err := conn.Exec(ctx, query, post.Title, post.Content, time.Now(), userID, postID)

	return err
}
