package store

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

type migration struct {
	version int
	up      func(*sql.Tx) error
}

var migrations = []migration{
	{
		version: 1,
		up: func(tx *sql.Tx) error {
			_, err := tx.Exec(`ALTER TABLE throws ADD COLUMN valid INTEGER NOT NULL DEFAULT 1`)
			return err
		},
	},
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
	// Create schema_version table first
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Printf("Error creating schema_version table: %v", err)
		return err
	}

	// Create base tables (if not exists)
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
		double_out INTEGER DEFAULT 0,
		winner_id INTEGER,
		current_player_index INTEGER DEFAULT 0,
		current_throw_number INTEGER DEFAULT 0,
		current_turn_points INTEGER DEFAULT 0,
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

	-- Create indexes for better query performance
	CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
	CREATE INDEX IF NOT EXISTS idx_games_winner ON games(winner_id);
	CREATE INDEX IF NOT EXISTS idx_game_players_user ON game_players(user_id);
	CREATE INDEX IF NOT EXISTS idx_throws_game ON throws(game_id);
	CREATE INDEX IF NOT EXISTS idx_throws_user ON throws(user_id);
	CREATE INDEX IF NOT EXISTS idx_throws_created ON throws(created_at);
	`

	_, err = s.db.Exec(schema)
	if err != nil {
		log.Printf("Error initializing schema: %v", err)
		return err
	}

	// Run migrations
	for _, m := range migrations {
		if err := s.runMigration(m); err != nil {
			return fmt.Errorf("migration %d failed: %w", m.version, err)
		}
	}

	return nil
}

func (s *Store) runMigration(m migration) error {
	// Check if already applied
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM schema_version WHERE version = ?`, m.version).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Already applied
	}

	// Run migration in transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := m.up(tx); err != nil {
		return err
	}

	// Record migration
	_, err = tx.Exec(`INSERT INTO schema_version (version) VALUES (?)`, m.version)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) Close() error {
	return s.db.Close()
}
