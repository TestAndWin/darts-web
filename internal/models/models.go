package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type GameStatus string

const (
	GameStatusPending  GameStatus = "PENDING"
	GameStatusActive   GameStatus = "ACTIVE"
	GameStatusFinished GameStatus = "FINISHED"
)

type Game struct {
	ID        int          `json:"id"`
	Status    GameStatus   `json:"status"`
	Settings  GameSettings `json:"settings"`
	WinnerID  *int         `json:"winner_id,omitempty"`
	CreatedAt time.Time    `json:"created_at"`

	Players     []GamePlayer `json:"players"`
	CurrentTurn *TurnStatus  `json:"current_turn"`
}

type GameSettings struct {
	TotalPoints int  `json:"total_points"` // 301, 501
	BestOfSets  int  `json:"best_of_sets"` // 1, 3, 5
	DoubleOut   bool `json:"double_out"`   // Require double to finish
}

type GamePlayer struct {
	UserID        int `json:"user_id"`
	Order         int `json:"order"`
	SetsWon       int `json:"sets_won"`
	LegsWon       int `json:"legs_won"` // Create this if we track legs properly
	CurrentPoints int `json:"current_points"`
}

type TurnStatus struct {
	PlayerIndex       int `json:"player_index"` // Index in Players array
	ThrowNumber       int `json:"throw_number"` // 0, 1, 2
	CurrentTurnPoints int `json:"current_turn_points"`
}

type Throw struct {
	ID         int       `json:"id"`
	GameID     int       `json:"game_id"`
	UserID     int       `json:"user_id"`
	Points     int       `json:"points"`
	Multiplier int       `json:"multiplier"` // 1, 2, 3
	Valid      bool      `json:"valid"`      // False if bust
	ScoreAfter int       `json:"score_after"`
	CreatedAt  time.Time `json:"created_at"`
}
