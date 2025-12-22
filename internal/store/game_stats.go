package store

import (
	"github.com/michaelschlottmann/darts-web/internal/models"
)

// GameStatistics represents comprehensive statistics for a finished game
type GameStatistics struct {
	GameID          int                `json:"game_id"`
	TotalSetsPlayed int                `json:"total_sets_played"`
	Players         []PlayerGameStats  `json:"players"`
}

// PlayerGameStats contains all statistics for a single player in a game
type PlayerGameStats struct {
	UserID       int          `json:"user_id"`
	UserName     string       `json:"user_name"`
	OverallStats OverallStats `json:"overall_stats"`
	SetStats     []SetStats   `json:"set_stats"`
}

// OverallStats contains aggregate statistics across all sets
type OverallStats struct {
	TotalThrows  int     `json:"total_throws"`
	TotalPoints  int     `json:"total_points"`
	Average3Dart float64 `json:"average_3_dart"`
}

// SetStats contains statistics for a single set
type SetStats struct {
	SetNumber    int     `json:"set_number"`
	TotalThrows  int     `json:"total_throws"`
	TotalPoints  int     `json:"total_points"`
	Average3Dart float64 `json:"average_3_dart"`
	WonSet       bool    `json:"won_set"`
}

// GetGameStatistics calculates comprehensive statistics for a finished game
func (s *Store) GetGameStatistics(gameID int) (*GameStatistics, error) {
	// Fetch game details
	game, err := s.GetGame(gameID)
	if err != nil {
		return nil, err
	}
	if game == nil {
		return nil, nil
	}

	// Fetch all throws for this game, ordered by creation time
	rows, err := s.db.Query(`
		SELECT id, game_id, user_id, points, multiplier, score_after, valid, created_at
		FROM throws
		WHERE game_id = ?
		ORDER BY created_at ASC
	`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var throws []models.Throw
	for rows.Next() {
		var t models.Throw
		var validInt int
		if err := rows.Scan(&t.ID, &t.GameID, &t.UserID, &t.Points, &t.Multiplier, &t.ScoreAfter, &validInt, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.Valid = validInt == 1
		throws = append(throws, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Detect set boundaries and group throws by set
	sets := detectSetBoundaries(throws, game.Settings.TotalPoints, game.Players)

	// Calculate statistics for each player
	playerStats := make(map[int]*PlayerGameStats)
	for _, player := range game.Players {
		// Fetch user name
		var userName string
		err := s.db.QueryRow("SELECT name FROM users WHERE id = ?", player.UserID).Scan(&userName)
		if err != nil {
			userName = "Unknown"
		}

		playerStats[player.UserID] = &PlayerGameStats{
			UserID:       player.UserID,
			UserName:     userName,
			OverallStats: OverallStats{},
			SetStats:     make([]SetStats, 0),
		}
	}

	// Only calculate set stats if there are sets
	if len(sets) > 0 {
		// Calculate statistics for each set
		for setNum, setThrows := range sets {
			// Calculate per-player stats for this set
			setPlayerStats := calculateSetPlayerStats(setThrows, game.Settings.TotalPoints)

			// Determine who won this set (player who reached 0)
			setWinner := findSetWinner(setThrows)

			// Add set stats to each player
			for userID, stats := range playerStats {
				// Get player stats for this set (will be zero values if player didn't throw)
				playerSetStat := setPlayerStats[userID]

				setStats := SetStats{
					SetNumber:    setNum + 1,
					TotalThrows:  playerSetStat.totalThrows,
					TotalPoints:  playerSetStat.totalPoints,
					Average3Dart: playerSetStat.average3Dart,
					WonSet:       userID == setWinner,
				}
				stats.SetStats = append(stats.SetStats, setStats)

				// Accumulate overall stats
				stats.OverallStats.TotalThrows += playerSetStat.totalThrows
				stats.OverallStats.TotalPoints += playerSetStat.totalPoints
			}
		}
	}

	// Calculate overall 3-dart averages
	for _, stats := range playerStats {
		if stats.OverallStats.TotalThrows > 0 {
			stats.OverallStats.Average3Dart = (float64(stats.OverallStats.TotalPoints) / float64(stats.OverallStats.TotalThrows)) * 3
		}
	}

	// Convert map to slice
	var playersSlice []PlayerGameStats
	for _, player := range game.Players {
		if stats, ok := playerStats[player.UserID]; ok {
			playersSlice = append(playersSlice, *stats)
		}
	}

	return &GameStatistics{
		GameID:          gameID,
		TotalSetsPlayed: len(sets),
		Players:         playersSlice,
	}, nil
}

// detectSetBoundaries analyzes throw history to detect set boundaries
// Returns a slice of sets, where each set is a slice of throws
func detectSetBoundaries(throws []models.Throw, totalPoints int, players []models.GamePlayer) [][]models.Throw {
	if len(throws) == 0 {
		return [][]models.Throw{}
	}

	var sets [][]models.Throw
	currentSet := []models.Throw{}

	// Track each player's last known score to detect resets
	playerScores := make(map[int]int)
	for _, player := range players {
		playerScores[player.UserID] = totalPoints
	}

	for _, throw := range throws {
		currentSet = append(currentSet, throw)

		// Check if this throw caused a checkout (player reached 0)
		if throw.ScoreAfter == 0 {
			// Set complete - player won
			sets = append(sets, currentSet)
			currentSet = []models.Throw{}
			// Reset all player scores for next set
			for userID := range playerScores {
				playerScores[userID] = totalPoints
			}
			continue
		}

		// Check for score reset (new set after previous set ended)
		// This happens when a player's score jumps back to totalPoints after being lower
		if prevScore, exists := playerScores[throw.UserID]; exists {
			// If score is back at totalPoints and it was previously lower, new set started
			if throw.ScoreAfter == totalPoints && prevScore < totalPoints && prevScore > 0 {
				// Previous set ended, start new set
				// Note: current throw belongs to new set, so don't include it in previous set
				if len(currentSet) > 1 {
					sets = append(sets, currentSet[:len(currentSet)-1])
					currentSet = []models.Throw{throw}
					// Reset tracking
					for userID := range playerScores {
						playerScores[userID] = totalPoints
					}
				}
			}
		}

		// Update player's score
		playerScores[throw.UserID] = throw.ScoreAfter
	}

	// Add final set if it has throws
	if len(currentSet) > 0 {
		sets = append(sets, currentSet)
	}

	return sets
}

// playerSetStats is a helper struct for calculating statistics
type playerSetStats struct {
	totalThrows  int
	totalPoints  int
	average3Dart float64
}

// calculateSetPlayerStats calculates statistics for each player in a set
func calculateSetPlayerStats(setThrows []models.Throw, totalPoints int) map[int]playerSetStats {
	stats := make(map[int]playerSetStats)

	for _, throw := range setThrows {
		s := stats[throw.UserID]
		s.totalThrows++
		// Only include valid throws in point calculation
		// Bust throws count toward throw total but not points
		if throw.Valid {
			s.totalPoints += throw.Points * throw.Multiplier
		}
		stats[throw.UserID] = s
	}

	// Calculate 3-dart averages
	for userID, s := range stats {
		if s.totalThrows > 0 {
			s.average3Dart = (float64(s.totalPoints) / float64(s.totalThrows)) * 3
			stats[userID] = s
		}
	}

	return stats
}

// findSetWinner returns the user ID of the player who won the set
// Returns 0 if no winner found (shouldn't happen in finished games)
func findSetWinner(setThrows []models.Throw) int {
	// Look for the throw that brought a player to exactly 0
	for _, throw := range setThrows {
		if throw.ScoreAfter == 0 {
			return throw.UserID
		}
	}
	return 0
}
