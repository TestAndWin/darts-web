package store

import (
	"database/sql"

	"github.com/michaelschlottmann/darts-web/internal/models"
)

func (s *Store) CreateGame(totalPoints, bestOf int, playerIDs []int) (*models.Game, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create Game
	res, err := tx.Exec(`INSERT INTO games (status, total_points, best_of_sets, current_player_index, current_throw_number) VALUES (?, ?, ?, 0, 0)`,
		models.GameStatusPending, totalPoints, bestOf)
	if err != nil {
		return nil, err
	}
	gameID, _ := res.LastInsertId()

	// Add Players
	players := make([]models.GamePlayer, len(playerIDs))
	for i, uid := range playerIDs {
		_, err := tx.Exec(`INSERT INTO game_players (game_id, user_id, player_order, current_points) VALUES (?, ?, ?, ?)`,
			gameID, uid, i, totalPoints)
		if err != nil {
			return nil, err
		}
		players[i] = models.GamePlayer{
			UserID:        uid,
			Order:         i,
			CurrentPoints: totalPoints,
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Game{
		ID:          int(gameID),
		Status:      models.GameStatusPending,
		Settings:    models.GameSettings{TotalPoints: totalPoints, BestOfSets: bestOf},
		Players:     players,
		CurrentTurn: &models.TurnStatus{PlayerIndex: 0, ThrowNumber: 0},
	}, nil
}

func (s *Store) GetGame(id int) (*models.Game, error) {
	// Simple fetch, in production might join everything
	// For MVP, structured manual queries

	// Game
	var g models.Game
	var statusStr string

	// Create struct to hold turn status if it's nil
	g.CurrentTurn = &models.TurnStatus{}

	err := s.db.QueryRow(`SELECT id, status, total_points, best_of_sets, winner_id, current_player_index, current_throw_number, current_turn_points, created_at FROM games WHERE id = ?`, id).
		Scan(&g.ID, &statusStr, &g.Settings.TotalPoints, &g.Settings.BestOfSets, &g.WinnerID, &g.CurrentTurn.PlayerIndex, &g.CurrentTurn.ThrowNumber, &g.CurrentTurn.CurrentTurnPoints, &g.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, err
	}
	g.Status = models.GameStatus(statusStr)

	// Players
	rows, err := s.db.Query(`SELECT user_id, player_order, sets_won, current_points FROM game_players WHERE game_id = ? ORDER BY player_order`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p models.GamePlayer
		if err := rows.Scan(&p.UserID, &p.Order, &p.SetsWon, &p.CurrentPoints); err != nil {
			return nil, err
		}
		g.Players = append(g.Players, p)
	}

	return &g, nil
}

func (s *Store) SaveThrow(t *models.Throw) error {
	_, err := s.db.Exec(`INSERT INTO throws (game_id, user_id, points, multiplier, score_after) VALUES (?, ?, ?, ?, ?)`,
		t.GameID, t.UserID, t.Points, t.Multiplier, t.ScoreAfter)
	return err
}

func (s *Store) UpdateGame(g *models.Game) error {
	// Use transaction to ensure atomic updates
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update Game Status
	_, err = tx.Exec(`UPDATE games SET status = ?, winner_id = ?, current_player_index = ?, current_throw_number = ?, current_turn_points = ? WHERE id = ?`,
		g.Status, g.WinnerID, g.CurrentTurn.PlayerIndex, g.CurrentTurn.ThrowNumber, g.CurrentTurn.CurrentTurnPoints, g.ID)
	if err != nil {
		return err
	}

	// Update Players
	for _, p := range g.Players {
		_, err := tx.Exec(`UPDATE game_players SET sets_won = ?, current_points = ? WHERE game_id = ? AND user_id = ?`,
			p.SetsWon, p.CurrentPoints, g.ID, p.UserID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
