package post_actions

import (
	"context"
	"forum/internal/models"
	"log"
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
	WHERE user_id = $4 AND post_id = $5
	`
	_, err := conn.Exec(ctx, query, post.Title, post.Content, time.Now(), userID, postID)

	return err
}

func DeletePost(ctx context.Context, conn *pgx.Conn, userID int, postID int) error {
	query := `
	DELETE FROM posts
	WHERE userID = $1 AND postID = $2`

	_, err := conn.Exec(ctx, query, userID, postID)

	return err

}

func GetPost(ctx context.Context, conn *pgx.Conn, postID int) (models.Post, error) {
	query := `
	SELECT post_id, user_id, title, content, created_at
	FROM posts
	WHERE post_id = $1`

	var post models.Post
	if err := conn.QueryRow(ctx, query, postID).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
	); err != nil {
		log.Println(err)
		return models.Post{}, err
	}

	return post, nil
}

func GetAllPosts(ctx context.Context, conn *pgx.Conn) ([]models.Post, error) {
	query := `
	SELECT post_id, user_id, title, content, created_at
	FROM posts
	ORDER BY created_at DESC`

	posts := make([]models.Post, 0)

	rows, err := conn.Query(ctx, query)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
		)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err

	}

	return posts, nil
}
