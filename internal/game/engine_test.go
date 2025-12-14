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

func TestProcessThrow_BustAfterMultipleThrows(t *testing.T) {
	engine := NewEngine()

	game := &models.Game{
		ID:     1,
		Status: models.GameStatusActive,
		Settings: models.GameSettings{
			TotalPoints: 501,
			BestOfSets:  1,
		},
		Players: []models.GamePlayer{
			{UserID: 1, Order: 0, CurrentPoints: 100, SetsWon: 0},
			{UserID: 2, Order: 1, CurrentPoints: 200, SetsWon: 0},
		},
		CurrentTurn: &models.TurnStatus{
			PlayerIndex:       0,
			ThrowNumber:       0,
			CurrentTurnPoints: 0,
		},
	}

	// First throw: Triple 20 (60 points) - should leave 40 points
	throw1, err := engine.ProcessThrow(game, 1, 20, 3)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}
	if !throw1.Valid {
		t.Error("Expected first throw to be valid")
	}
	if game.Players[0].CurrentPoints != 40 {
		t.Errorf("Expected 40 points after first throw, got %d", game.Players[0].CurrentPoints)
	}

	// Second throw: Triple 20 (60 points) - should bust (40 - 60 = -20)
	throw2, err := engine.ProcessThrow(game, 1, 20, 3)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}

	if throw2.Valid {
		t.Error("Expected bust throw to be invalid")
	}

	// Points should revert to 100 (start of turn), not 40
	if game.Players[0].CurrentPoints != 100 {
		t.Errorf("Expected points to revert to 100 after bust, got %d", game.Players[0].CurrentPoints)
	}

	// Turn should have moved to next player
	if game.CurrentTurn.PlayerIndex != 1 {
		t.Errorf("Expected player index to be 1 after bust, got %d", game.CurrentTurn.PlayerIndex)
	}

	// Turn points should be reset
	if game.CurrentTurn.CurrentTurnPoints != 0 {
		t.Errorf("Expected turn points to be reset to 0, got %d", game.CurrentTurn.CurrentTurnPoints)
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

func TestProcessThrow_RemainingOne_WithDoubleOut(t *testing.T) {
	engine := NewEngine()

	game := &models.Game{
		ID:     1,
		Status: models.GameStatusActive,
		Settings: models.GameSettings{
			TotalPoints: 501,
			BestOfSets:  1,
			DoubleOut:   true, // With double-out, landing on 1 is a bust
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

	// Throw that leaves 1 point (bust when double-out is enabled)
	throw, err := engine.ProcessThrow(game, 1, 1, 1)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}

	if throw.Valid {
		t.Error("Expected throw leaving 1 point to be invalid (bust) when double-out is enabled")
	}

	if game.Players[0].CurrentPoints != 2 {
		t.Errorf("Expected points to remain 2 after bust, got %d", game.Players[0].CurrentPoints)
	}
}

func TestProcessThrow_RemainingOne_NoDoubleOut(t *testing.T) {
	engine := NewEngine()

	game := &models.Game{
		ID:     1,
		Status: models.GameStatusActive,
		Settings: models.GameSettings{
			TotalPoints: 501,
			BestOfSets:  1,
			DoubleOut:   false, // Without double-out, landing on 1 is valid
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

	// Throw that leaves 1 point (valid when double-out is disabled)
	throw, err := engine.ProcessThrow(game, 1, 1, 1)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}

	if !throw.Valid {
		t.Error("Expected throw leaving 1 point to be valid when double-out is disabled")
	}

	if game.Players[0].CurrentPoints != 1 {
		t.Errorf("Expected points to be 1 after valid throw, got %d", game.Players[0].CurrentPoints)
	}

	// Game should not be finished yet (still needs to checkout with 1)
	if game.Status == models.GameStatusFinished {
		t.Error("Expected game to not be finished yet (player at 1 point)")
	}
}

func TestProcessThrow_DoubleOutCheckout_Valid(t *testing.T) {
	engine := NewEngine()

	game := &models.Game{
		ID:     1,
		Status: models.GameStatusActive,
		Settings: models.GameSettings{
			TotalPoints: 501,
			BestOfSets:  1,
			DoubleOut:   true,
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

	// Throw double 20 (40 points) - valid checkout with double
	throw, err := engine.ProcessThrow(game, 1, 20, 2)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}

	if !throw.Valid {
		t.Error("Expected double checkout to be valid when double-out is enabled")
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

func TestProcessThrow_DoubleOutCheckout_Invalid_SingleFinish(t *testing.T) {
	engine := NewEngine()

	game := &models.Game{
		ID:     1,
		Status: models.GameStatusActive,
		Settings: models.GameSettings{
			TotalPoints: 501,
			BestOfSets:  1,
			DoubleOut:   true,
		},
		Players: []models.GamePlayer{
			{UserID: 1, Order: 0, CurrentPoints: 20, SetsWon: 0},
		},
		CurrentTurn: &models.TurnStatus{
			PlayerIndex:       0,
			ThrowNumber:       0,
			CurrentTurnPoints: 0,
		},
	}

	// Throw single 20 (20 points) - should bust when double-out is required
	throw, err := engine.ProcessThrow(game, 1, 20, 1)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}

	if throw.Valid {
		t.Error("Expected single finish to be invalid when double-out is enabled")
	}

	if game.Players[0].CurrentPoints != 20 {
		t.Errorf("Expected points to remain 20 after bust, got %d", game.Players[0].CurrentPoints)
	}

	if game.Status == models.GameStatusFinished {
		t.Error("Expected game to not be finished after invalid checkout")
	}
}

func TestProcessThrow_DoubleOutCheckout_Invalid_TripleFinish(t *testing.T) {
	engine := NewEngine()

	game := &models.Game{
		ID:     1,
		Status: models.GameStatusActive,
		Settings: models.GameSettings{
			TotalPoints: 501,
			BestOfSets:  1,
			DoubleOut:   true,
		},
		Players: []models.GamePlayer{
			{UserID: 1, Order: 0, CurrentPoints: 60, SetsWon: 0},
		},
		CurrentTurn: &models.TurnStatus{
			PlayerIndex:       0,
			ThrowNumber:       0,
			CurrentTurnPoints: 0,
		},
	}

	// Throw triple 20 (60 points) - should bust when double-out is required
	throw, err := engine.ProcessThrow(game, 1, 20, 3)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}

	if throw.Valid {
		t.Error("Expected triple finish to be invalid when double-out is enabled")
	}

	if game.Players[0].CurrentPoints != 60 {
		t.Errorf("Expected points to remain 60 after bust, got %d", game.Players[0].CurrentPoints)
	}

	if game.Status == models.GameStatusFinished {
		t.Error("Expected game to not be finished after invalid checkout")
	}
}

func TestProcessThrow_DoubleOutCheckout_DoubleBull(t *testing.T) {
	engine := NewEngine()

	game := &models.Game{
		ID:     1,
		Status: models.GameStatusActive,
		Settings: models.GameSettings{
			TotalPoints: 501,
			BestOfSets:  1,
			DoubleOut:   true,
		},
		Players: []models.GamePlayer{
			{UserID: 1, Order: 0, CurrentPoints: 50, SetsWon: 0},
		},
		CurrentTurn: &models.TurnStatus{
			PlayerIndex:       0,
			ThrowNumber:       0,
			CurrentTurnPoints: 0,
		},
	}

	// Throw double bull (25 * 2 = 50 points) - valid checkout with double
	throw, err := engine.ProcessThrow(game, 1, 25, 2)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}

	if !throw.Valid {
		t.Error("Expected double bull checkout to be valid when double-out is enabled")
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

func TestProcessThrow_NoDoubleOut_AnyCheckout(t *testing.T) {
	engine := NewEngine()

	game := &models.Game{
		ID:     1,
		Status: models.GameStatusActive,
		Settings: models.GameSettings{
			TotalPoints: 501,
			BestOfSets:  1,
			DoubleOut:   false, // Double-out disabled
		},
		Players: []models.GamePlayer{
			{UserID: 1, Order: 0, CurrentPoints: 20, SetsWon: 0},
		},
		CurrentTurn: &models.TurnStatus{
			PlayerIndex:       0,
			ThrowNumber:       0,
			CurrentTurnPoints: 0,
		},
	}

	// Throw single 20 - should be valid when double-out is disabled
	throw, err := engine.ProcessThrow(game, 1, 20, 1)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}

	if !throw.Valid {
		t.Error("Expected single finish to be valid when double-out is disabled")
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

func TestProcessThrow_DoubleOut_BustAfterValidThrows(t *testing.T) {
	engine := NewEngine()

	game := &models.Game{
		ID:     1,
		Status: models.GameStatusActive,
		Settings: models.GameSettings{
			TotalPoints: 501,
			BestOfSets:  1,
			DoubleOut:   true,
		},
		Players: []models.GamePlayer{
			{UserID: 1, Order: 0, CurrentPoints: 100, SetsWon: 0},
			{UserID: 2, Order: 1, CurrentPoints: 200, SetsWon: 0},
		},
		CurrentTurn: &models.TurnStatus{
			PlayerIndex:       0,
			ThrowNumber:       0,
			CurrentTurnPoints: 0,
		},
	}

	// First throw: Triple 20 (60 points) - should leave 40 points
	throw1, err := engine.ProcessThrow(game, 1, 20, 3)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}
	if !throw1.Valid {
		t.Error("Expected first throw to be valid")
	}
	if game.Players[0].CurrentPoints != 40 {
		t.Errorf("Expected 40 points after first throw, got %d", game.Players[0].CurrentPoints)
	}

	// Second throw: Single 20 (20 points) - should leave 20 points
	throw2, err := engine.ProcessThrow(game, 1, 20, 1)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}
	if !throw2.Valid {
		t.Error("Expected second throw to be valid")
	}
	if game.Players[0].CurrentPoints != 20 {
		t.Errorf("Expected 20 points after second throw, got %d", game.Players[0].CurrentPoints)
	}

	// Third throw: Single 20 (20 points) - reaches 0 but not with double, should bust
	throw3, err := engine.ProcessThrow(game, 1, 20, 1)
	if err != nil {
		t.Fatalf("ProcessThrow() error = %v", err)
	}

	if throw3.Valid {
		t.Error("Expected third throw (single finish) to be invalid when double-out is enabled")
	}

	// Points should revert to 100 (start of turn), not 20
	if game.Players[0].CurrentPoints != 100 {
		t.Errorf("Expected points to revert to 100 after bust, got %d", game.Players[0].CurrentPoints)
	}

	// Turn should have moved to next player
	if game.CurrentTurn.PlayerIndex != 1 {
		t.Errorf("Expected player index to be 1 after bust, got %d", game.CurrentTurn.PlayerIndex)
	}

	// Turn points should be reset
	if game.CurrentTurn.CurrentTurnPoints != 0 {
		t.Errorf("Expected turn points to be reset to 0, got %d", game.CurrentTurn.CurrentTurnPoints)
	}
}
