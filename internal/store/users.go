package store

import (
	"database/sql"

	"github.com/michaelschlottmann/darts-web/internal/models"
)

func (s *Store) CreateUser(name string) (*models.User, error) {
	// Use INSERT OR IGNORE with RETURNING to handle duplicates atomically
	// This is more efficient than separate SELECT + INSERT
	query := `
		INSERT INTO users (name)
		VALUES (?)
		ON CONFLICT(name) DO UPDATE SET name=name
		RETURNING id, name, created_at
	`
	var user models.User
	err := s.db.QueryRow(query, name).Scan(&user.ID, &user.Name, &user.CreatedAt)
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
