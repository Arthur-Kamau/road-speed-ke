package models

import "time"

type User struct {
	ID         int64     `json:"id"`
	GoogleID   string    `json:"google_id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	PictureURL string    `json:"picture_url"`
	CreatedAt  time.Time `json:"created_at"`
}

type GoogleAuthRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}
