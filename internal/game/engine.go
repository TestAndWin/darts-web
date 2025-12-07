package game

import (
	"errors"

	"github.com/michaelschlottmann/darts-web/internal/models"
)

var (
	ErrGameFinished  = errors.New("game is already finished")
	ErrNotPlayerTurn = errors.New("not this player's turn")
	ErrInvalidThrow  = errors.New("invalid throw value")
)

type Engine struct {
}

func NewEngine() *Engine {
	return &Engine{}
}

// ProcessThrow handles a single dart throw logic
func (e *Engine) ProcessThrow(game *models.Game, userID int, points int, multiplier int) (*models.Throw, error) {
	if game.Status == models.GameStatusFinished {
		return nil, ErrGameFinished
	}

	currentPlayer := &game.Players[game.CurrentTurn.PlayerIndex]
	if currentPlayer.UserID != userID {
		return nil, ErrNotPlayerTurn
	}

	// Validate throw values
	if err := e.validateThrow(points, multiplier); err != nil {
		return nil, err
	}

	realPoints := points * multiplier

	// Store the score at the start of this throw for bust handling
	scoreBeforeThrow := currentPlayer.CurrentPoints

	// Create throw object
	throw := &models.Throw{
		GameID:     game.ID,
		UserID:     userID,
		Points:     points,
		Multiplier: multiplier,
		Valid:      true,
		ScoreAfter: scoreBeforeThrow,
	}

	// Update turn stats
	game.CurrentTurn.ThrowNumber++
	game.CurrentTurn.CurrentTurnPoints += realPoints

	remaining := currentPlayer.CurrentPoints - realPoints

	// Bust check: remaining < 0 or remaining == 1 (impossible to checkout)
	isBust := false
	if remaining < 0 || remaining == 1 {
		isBust = true
	} else if remaining == 0 {
		// Checkout - player wins the leg/set
		currentPlayer.CurrentPoints = 0
		throw.ScoreAfter = 0
		e.handleWinSet(game, currentPlayer)
		return throw, nil
	}

	if isBust {
		// On bust: score reverts to beginning of turn, turn ends
		// Calculate the score at the start of this turn
		// Current points + all turn points - this throw = turn start score
		turnStartScore := currentPlayer.CurrentPoints + (game.CurrentTurn.CurrentTurnPoints - realPoints)

		throw.Valid = false
		throw.ScoreAfter = turnStartScore
		currentPlayer.CurrentPoints = turnStartScore

		// Reset turn points and move to next player
		game.CurrentTurn.CurrentTurnPoints = 0
		e.nextPlayer(game)
	} else {
		// Valid throw, update score
		currentPlayer.CurrentPoints = remaining
		throw.ScoreAfter = remaining

		if game.CurrentTurn.ThrowNumber >= 3 {
			// End of turn after 3 throws
			e.nextPlayer(game)
		}
	}

	return throw, nil
}

// validateThrow checks if the throw values are valid
func (e *Engine) validateThrow(points int, multiplier int) error {
	// Validate points
	if points < 0 || points > 25 {
		return ErrInvalidThrow
	}

	// Points must be 0-20 or 25 (bull)
	if points > 20 && points != 25 {
		return ErrInvalidThrow
	}

	// Validate multiplier
	if multiplier < 1 || multiplier > 3 {
		return ErrInvalidThrow
	}

	// Bull can only be single (25) or double (50), not triple
	if points == 25 && multiplier == 3 {
		return ErrInvalidThrow
	}

	// Miss (0) can only have multiplier 1
	if points == 0 && multiplier != 1 {
		return ErrInvalidThrow
	}

	return nil
}

func (e *Engine) handleWinSet(game *models.Game, player *models.GamePlayer) {
	player.SetsWon++

	setsNeeded := (game.Settings.BestOfSets + 1) / 2

	if player.SetsWon >= setsNeeded {
		game.Status = models.GameStatusFinished
		wID := player.UserID
		game.WinnerID = &wID
	} else {
		// Next set
		// Reset points for all players
		for i := range game.Players {
			game.Players[i].CurrentPoints = game.Settings.TotalPoints
		}
		// In C++, nextPlayer() is called after a set win unless match done.
		e.nextPlayer(game)
	}
}

func (e *Engine) nextPlayer(game *models.Game) {
	game.CurrentTurn.ThrowNumber = 0
	game.CurrentTurn.CurrentTurnPoints = 0

	game.CurrentTurn.PlayerIndex++
	if game.CurrentTurn.PlayerIndex >= len(game.Players) {
		game.CurrentTurn.PlayerIndex = 0
	}
}
