package user

import (
	"time"

	"github.com/google/uuid"
)

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id uuid.UUID) (*User, error)
	CreateUser(name, email,googleAccessToken, picture string) (*User, error)
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Name 	string    `json:"name"`
	Email	 string    `json:"email"`
	GoogleAccessToken string `json:"google_access_token"`
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
	GoogleAccessToken string `json:"google_access_token"`
	Picture string `json:"picture"`
}

type LoginResponse struct {
	Token string `json:"token"`
	IsNewUser bool `json:"isNewUser"`
}