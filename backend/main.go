package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	// Game modes
	GameModeClassic  = "classic"  // 3x3 board
	GameModeAdvanced = "advanced" // 5x5 board

	// Opcodes
	OpcodeMove        = 1
	OpcodeState       = 2
	OpcodeError       = 3
	OpcodeMatchFound  = 4
	OpcodeLeaderboard = 5

	// Game states
	GameStateWaiting  = "waiting"
	GameStatePlaying  = "playing"
	GameStateFinished = "finished"

	// Player symbols
	PlayerX = "X"
	PlayerO = "O"
	Empty   = ""
)

// MoveData represents a move from client
type MoveData struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// StateData represents game state broadcast
type StateData struct {
	Board   [][]string        `json:"board"`
	Turn    string            `json:"turn"`
	Winner  string            `json:"winner,omitempty"`
	Size    int               `json:"size"`
	Mode    string            `json:"mode"`
	Players map[string]string `json:"players"` // userID -> symbol
}

// ErrorData represents error message
type ErrorData struct {
	Msg string `json:"msg"`
}

// MatchFoundData represents match found notification
type MatchFoundData struct {
	MatchID string `json:"match_id"`
	Mode    string `json:"mode"`
}

// LeaderboardData represents leaderboard update
type LeaderboardData struct {
	Rankings []PlayerRanking `json:"rankings"`
}

// PlayerRanking represents a player's ranking
type PlayerRanking struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Score     int64  `json:"score"`
	Rank      int    `json:"rank"`
	GamesWon  int    `json:"games_won"`
	GamesLost int    `json:"games_lost"`
}

// InitModule initializes the Nakama module
func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("Initializing Tic-Tac-Toe module")

	// Register match handler
	if err := initializer.RegisterMatch("ttt_match", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		return &TTTMatchHandler{}, nil
	}); err != nil {
		return fmt.Errorf("failed to register match: %w", err)
	}

	// Initialize matchmaking system
	if err := InitMatchmaking(ctx, logger, db, nk, initializer); err != nil {
		return fmt.Errorf("failed to initialize matchmaking: %w", err)
	}

	// Initialize leaderboard system
	if err := InitLeaderboard(ctx, logger, db, nk, initializer); err != nil {
		return fmt.Errorf("failed to initialize leaderboard: %w", err)
	}

	logger.Info("Tic-Tac-Toe module initialized successfully")
	return nil
}
