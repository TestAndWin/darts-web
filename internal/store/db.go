package store

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.initSchema(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS games (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		status TEXT NOT NULL,
		total_points INTEGER NOT NULL,
		best_of_sets INTEGER NOT NULL,
		winner_id INTEGER,
		current_player_index INTEGER DEFAULT 0,
		current_throw_number INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (winner_id) REFERENCES users(id)
	);
	
	CREATE TABLE IF NOT EXISTS game_players (
		game_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		player_order INTEGER NOT NULL,
		sets_won INTEGER DEFAULT 0,
		current_points INTEGER,
		PRIMARY KEY (game_id, user_id),
		FOREIGN KEY (game_id) REFERENCES games(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS throws (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		points INTEGER NOT NULL,
		multiplier INTEGER NOT NULL,
		score_after INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`

	_, err := s.db.Exec(schema)
	if err != nil {
		log.Printf("Error initializing schema: %v", err)
		return err
	}
	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
