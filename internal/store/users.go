package store

import (
	"database/sql"
	"errors"

	"github.com/michaelschlottmann/darts-web/internal/models"
)

var ErrDuplicateUsername = errors.New("username already exists")

func (s *Store) CreateUser(name string) (*models.User, error) {
	// First check if user already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE name = ?)`
	err := s.db.QueryRow(checkQuery, name).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateUsername
	}

	// Insert new user
	query := `
		INSERT INTO users (name)
		VALUES (?)
		RETURNING id, name, created_at
	`
	var user models.User
	err = s.db.QueryRow(query, name).Scan(&user.ID, &user.Name, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) ListUsers() ([]models.User, error) {
	query := `SELECT id, name, created_at FROM users ORDER BY name`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *Store) GetUser(id int) (*models.User, error) {
	query := `SELECT id, name, created_at FROM users WHERE id = ?`
	var u models.User
	err := s.db.QueryRow(query, id).Scan(&u.ID, &u.Name, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
