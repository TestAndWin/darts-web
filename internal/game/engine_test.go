package game

import (
	"testing"

	"github.com/michaelschlottmann/darts-web/internal/models"
)

func TestValidateThrow(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name       string
		points     int
		multiplier int
		wantErr    bool
	}{
		{"Valid single 20", 20, 1, false},
		{"Valid double 20", 20, 2, false},
		{"Valid triple 20", 20, 3, false},
		{"Valid bull", 25, 1, false},
		{"Valid double bull", 25, 2, false},
		{"Invalid triple bull", 25, 3, true},
		{"Invalid points 26", 26, 1, true},
		{"Invalid points -1", -1, 1, true},
		{"Invalid points 21", 21, 1, true},
		{"Invalid multiplier 0", 20, 0, true},
		{"Invalid multiplier 4", 20, 4, true},
		{"Valid miss", 0, 1, false},
		{"Invalid double miss", 0, 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.validateThrow(tt.points, tt.multiplier)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateThrow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcessThrow_Bust(t *testing.T) {
	engine := NewEngine()

	game := &models.Game{
		ID:     1,
		Status: models.GameStatusActive,
		Settings: models.GameSettings{
			TotalPoints: 501,
			BestOfSets:  1,
		},
		Players: []models.GamePlayer{
			{UserID: 1, Order: 0, CurrentPoints: 30, SetsWon: 0},
		},
		CurrentTurn: &models.TurnStatus{
			PlayerIndex:       0,
			ThrowNumber:       0,
			CurrentTurnPoints: 0,
		},
	}

	// Throw that causes bust (30 - 60 = -30)
	throw, err := engine.ProcessThrow(game, 1, 20, 3)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}

	if throw.Valid {
		t.Error("Expected bust throw to be invalid")
	}

	if game.Players[0].CurrentPoints != 30 {
		t.Errorf("Expected points to remain 30 after bust, got %d", game.Players[0].CurrentPoints)
	}

	if game.CurrentTurn.PlayerIndex != 0 {
		t.Error("Expected player index to advance after bust")
	}
}

func TestProcessThrow_Checkout(t *testing.T) {
	engine := NewEngine()

	game := &models.Game{
		ID:     1,
		Status: models.GameStatusActive,
		Settings: models.GameSettings{
			TotalPoints: 501,
			BestOfSets:  1,
		},
		Players: []models.GamePlayer{
			{UserID: 1, Order: 0, CurrentPoints: 40, SetsWon: 0},
		},
		CurrentTurn: &models.TurnStatus{
			PlayerIndex:       0,
			ThrowNumber:       0,
			CurrentTurnPoints: 0,
		},
	}

	// Throw that checks out (40 - 40 = 0)
	throw, err := engine.ProcessThrow(game, 1, 20, 2)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}

	if !throw.Valid {
		t.Error("Expected checkout throw to be valid")
	}

	if game.Players[0].CurrentPoints != 0 {
		t.Errorf("Expected points to be 0 after checkout, got %d", game.Players[0].CurrentPoints)
	}

	if game.Status != models.GameStatusFinished {
		t.Errorf("Expected game status to be FINISHED, got %s", game.Status)
	}

	if game.WinnerID == nil || *game.WinnerID != 1 {
		t.Error("Expected winner to be set")
	}
}

func TestProcessThrow_RemainingOne(t *testing.T) {
	engine := NewEngine()

	game := &models.Game{
		ID:     1,
		Status: models.GameStatusActive,
		Settings: models.GameSettings{
			TotalPoints: 501,
			BestOfSets:  1,
		},
		Players: []models.GamePlayer{
			{UserID: 1, Order: 0, CurrentPoints: 2, SetsWon: 0},
		},
		CurrentTurn: &models.TurnStatus{
			PlayerIndex:       0,
			ThrowNumber:       0,
			CurrentTurnPoints: 0,
		},
	}

	// Throw that leaves 1 point (bust)
	throw, err := engine.ProcessThrow(game, 1, 1, 1)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}

	if throw.Valid {
		t.Error("Expected throw leaving 1 point to be invalid (bust)")
	}

	if game.Players[0].CurrentPoints != 2 {
		t.Errorf("Expected points to remain 2 after bust, got %d", game.Players[0].CurrentPoints)
	}
}
