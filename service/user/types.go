package user

import (
	"time"

	"github.com/google/uuid"
)

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id uuid.UUID) (*User, error)
	CreateUser(name, email, picture string) (*User, error)
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Name 	string    `json:"name"`
	Email	 string    `json:"email"`
	ProfilePicture string `json:"profile_picture"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type LoginPayload struct {
	AccessToken string `json:"access_token"`
}

type GoogleUser struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Picture string `json:"picture"`
}

type LoginResponse struct {
	Token string `json:"token"`
}