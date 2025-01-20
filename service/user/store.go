package user

import (
	"database/sql"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetUserByEmail(email string) (*User, error) {
	rows := s.db.QueryRow("SELECT * FROM users WHERE email = $1", email)

	u := new(User)
	u, err := ScanRowToUser(rows)
	if err != nil {
		return nil, err
	}
	return u, nil

}

func ScanRowToUser(rows *sql.Row) (*User, error) {

	user := new(User)
	err := rows.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.GoogleAccessToken,
		&user.GoogleRefreshToken,
		&user.ProfilePicture,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) GetUserByID(id uuid.UUID) (*User, error) {
	rows := s.db.QueryRow("SELECT * FROM users WHERE id = ?", id)

	u := new(User)
	u, err := ScanRowToUser(rows)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil

}

func (s *Store) CreateUser(name, email,googleAccessToken, refreshToken, picture string) (*User, error) {
	row := s.db.QueryRow("INSERT INTO users (name, email,google_access_token, refresh_token, profile_picture) VALUES ($1, $2, $3, $4, $5) RETURNING id, name, email,google_access_token, profile_picture, created_at, updated_at", name, email,googleAccessToken,refreshToken, picture)
	
	createdUser, err := ScanRowToUser(row)
	if err != nil {
		return nil, err
	}
	return createdUser, nil
}

func (s *Store) UpdateUserRefreshToken(email, refreshToken string) error {
	_, err := s.db.Exec("UPDATE users SET refresh_token = $1 WHERE email = $2", refreshToken, email)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store)AddCalendarIDToUser(id uuid.UUID, planCalendarId, logCalendarId string) error {
	_ , err := s.db.Exec("INSERT INTO users_calendars (user_id, plans_id, logs_id) VALUES ($1, $2, $3)", id, planCalendarId, logCalendarId)

	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetCalendarIDByUserID(id uuid.UUID) (string, string, error) {
	rows := s.db.QueryRow("SELECT plans_id, logs_id FROM users_calendars WHERE user_id = $1", id)

	var plansID, logsID string
	err := rows.Scan(&plansID, &logsID)
	if err != nil {
		return "", "", err
	}
	return plansID, logsID, nil
}