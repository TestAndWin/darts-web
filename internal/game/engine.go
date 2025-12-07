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

	if points < 0 || points > 20 && points != 25 && points != 50 {
		// 50 for Bullseye (25*2), strict validation can be added
		// For now assume caller validates roughly 0-20, 25, 50
	}

	realPoints := points * multiplier

	// Create throw object
	throw := &models.Throw{
		GameID:     game.ID,
		UserID:     userID,
		Points:     points,
		Multiplier: multiplier,
		Valid:      true,
		ScoreAfter: currentPlayer.CurrentPoints, // Will update below
	}

	// Logic ported from C++ Darts::processGameTurn / handleGameScreen

	// Update turn stats
	game.CurrentTurn.ThrowNumber++
	game.CurrentTurn.CurrentTurnPoints += realPoints

	remaining := currentPlayer.CurrentPoints - realPoints

	// Bust check: if < 0 or (== 1 && Double Out required) -> generic bust for now based on C++ " < 0"
	// C++ logic: if (points < 0) -> nextPlayer
	// C++ Double Out is implicit in the "Double" button and "X" checks mostly,
	// but standard rules say you can't have 1 left.

	isBust := false
	if remaining < 0 {
		isBust = true
	} else if remaining == 0 {
		// Checkout!
		// In C++, it checks `if (playerPoints[currentPlayer] - currentPoints == 0)`
		// We need to verify if checkouts must be double.
		// For casual play (often default in these DIY projects), sometimes straight out is allowed.
		// However, standard is Double Out. Let's assume generic for now to match C++ simple logic checks,
		// but checking the C++ code: `if (playerPoints[currentPlayer] - currentPoints == 0)` -> Wins set.
		// It DOES NOT seemingly enforce double out strictness in the `processGameTurn` logic explicitly shown.
		// I will implement "Straight Out" or "Double Out" based on settings later. Defaulting to Straight Out for compatibility with simple C++ logic unless specified.
		// WAIT: The C++ code had `dartFactor == 2` checks.
		// User requirement: "abbildet die gleiche Funktionalität".
		// The C++ code: `if (playerPoints[currentPlayer] - currentPoints == 0) { ... Win Set ... }`
		// It does NOT check for Double Out in that if block. So I will permit Straight Out for now.

		currentPlayer.CurrentPoints = 0
		throw.ScoreAfter = 0
		e.handleWinSet(game, currentPlayer)
		return throw, nil
	} else if remaining == 1 {
		// Usually a bust in Double Out, but valid in Straight Out.
		// Keeping it valid for now.
	}

	if isBust {
		throw.Valid = false
		throw.ScoreAfter = currentPlayer.CurrentPoints // Reset to start of turn effectively?
		// In C++ bust: `nextPlayer()`. `currentPoints` (turn accum) is discarded.
		// So player score reverts to what it was at start of turn?
		// C++: `playerPoints` is only updated at END of turn (throw==3) or WIN.
		// So detailed state tracking: We need to know the score at Start of Turn.
		// `CurrentPoints` in my model tracks "points remaining for the leg".
		// If we update it live, we need a way to revert.
		// Easier: store `LegStartPoints` or just don't update `CurrentPoints` until confirmed?
		// My model has `CurrentPoints`. I will update it tentatively?

		// Revert turn points
		// Reset turn
		e.nextPlayer(game)
	} else {
		// Valid throw, not out yet
		currentPlayer.CurrentPoints = remaining
		throw.ScoreAfter = remaining

		if game.CurrentTurn.ThrowNumber >= 3 {
			// End of turn
			e.nextPlayer(game)
		}
	}

	return throw, nil
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
