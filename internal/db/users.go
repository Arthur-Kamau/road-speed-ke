package db

import (
	"context"
	"fmt"

	"github.com/kamau/speed/internal/models"
)

func (q *Queries) UpsertUser(ctx context.Context, googleID, email, name, pictureURL string) (*models.User, error) {
	var u models.User
	err := q.pool.QueryRow(ctx, `
		INSERT INTO users (google_id, email, name, picture_url)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (google_id) DO UPDATE SET
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			picture_url = EXCLUDED.picture_url
		RETURNING id, google_id, email, name, picture_url, created_at
	`, googleID, email, name, pictureURL).Scan(
		&u.ID, &u.GoogleID, &u.Email, &u.Name, &u.PictureURL, &u.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upserting user: %w", err)
	}
	return &u, nil
}

func (q *Queries) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var u models.User
	err := q.pool.QueryRow(ctx, `
		SELECT id, google_id, email, name, picture_url, created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.GoogleID, &u.Email, &u.Name, &u.PictureURL, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting user by id: %w", err)
	}
	return &u, nil
}

func (q *Queries) GetUserByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	var u models.User
	err := q.pool.QueryRow(ctx, `
		SELECT id, google_id, email, name, picture_url, created_at
		FROM users WHERE google_id = $1
	`, googleID).Scan(&u.ID, &u.GoogleID, &u.Email, &u.Name, &u.PictureURL, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting user by google_id: %w", err)
	}
	return &u, nil
}
