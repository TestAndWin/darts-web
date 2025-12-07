package store

import "github.com/michaelschlottmann/darts-web/internal/models"

func (s *Store) GetUserStats(userID int) (map[string]interface{}, error) {
	// Simple stats: Total Games, Games Won
	var totalGames int
	// Only count FINISHED games
	err := s.db.QueryRow(`
		SELECT COUNT(*) 
		FROM game_players gp
		JOIN games g ON gp.game_id = g.id
		WHERE gp.user_id = ? AND g.status = ?`, userID, models.GameStatusFinished).Scan(&totalGames)
	if err != nil {
		return nil, err
	}

	var gamesWon int
	err = s.db.QueryRow(`SELECT COUNT(*) FROM games WHERE winner_id = ? AND status = ?`, userID, models.GameStatusFinished).Scan(&gamesWon)
	if err != nil {
		return nil, err
	}

	// Average calculation (simplification: average of recorded throws)
	var totalPoints, totalThrows int
	// Only count throws from FINISHED games
	err = s.db.QueryRow(`
		SELECT COALESCE(SUM(t.points * t.multiplier), 0), COUNT(*) 
		FROM throws t
		JOIN games g ON t.game_id = g.id
		WHERE t.user_id = ? AND g.status = ?`, userID, models.GameStatusFinished).Scan(&totalPoints, &totalThrows)
	if err != nil {
		return nil, err
	}

	average := 0.0
	if totalThrows > 0 {
		average = float64(totalPoints*3) / float64(totalThrows) // 3-dart average default
	}

	return map[string]interface{}{
		"total_games":    totalGames,
		"wins":           gamesWon,
		"average_3_dart": average,
		"total_throws":   totalThrows,
	}, nil
}
