package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

// TTTMatch represents a Tic-Tac-Toe match
type TTTMatch struct {
	ID        string
	Mode      string
	Size      int
	Board     [][]string
	Turn      string
	Winner    string
	State     string
	Players   map[string]string // userID -> symbol
	MoveCount int
	CreatedAt int64
}

// TTTMatchHandler implements the Match interface
type TTTMatchHandler struct{}

func (h *TTTMatchHandler) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	// Determine game mode and board size
	mode := GameModeClassic
	if modeParam, ok := params["mode"].(string); ok {
		mode = modeParam
	}

	size := 3
	if mode == GameModeAdvanced {
		size = 5
	}

	match := &TTTMatch{
		ID:        "",
		Mode:      mode,
		Size:      size,
		Board:     make([][]string, size),
		Turn:      PlayerX,
		Winner:    "",
		State:     GameStateWaiting,
		Players:   make(map[string]string),
		MoveCount: 0,
		CreatedAt: time.Now().Unix(),
	}

	// Initialize empty board
	for i := range match.Board {
		match.Board[i] = make([]string, size)
		for j := range match.Board[i] {
			match.Board[i][j] = Empty
		}
	}

	logger.Info("Initialized %s match with %dx%d board", mode, size, size)
	return match, 2, ""
}

func (h *TTTMatchHandler) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	match := state.(*TTTMatch)

	// Check if match is full
	if len(match.Players) >= 2 {
		return match, false, "Match is full"
	}

	// Check if match is already finished
	if match.State == GameStateFinished {
		return match, false, "Match is finished"
	}

	// Assign player symbol
	symbol := PlayerX
	if len(match.Players) == 1 {
		symbol = PlayerO
	}

	match.Players[presence.GetUserId()] = symbol

	// Start game if we have 2 players
	if len(match.Players) == 2 {
		match.State = GameStatePlaying
		logger.Info("Match started with 2 players")
	}

	return match, true, ""
}

func (h *TTTMatchHandler) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	match := state.(*TTTMatch)

	// Send match found notification
	for _, presence := range presences {
		matchFoundData := MatchFoundData{
			MatchID: match.ID,
			Mode:    match.Mode,
		}
		matchFoundBytes, _ := json.Marshal(matchFoundData)
		dispatcher.BroadcastMessage(OpcodeMatchFound, matchFoundBytes, []runtime.Presence{presence}, nil, false)
	}

	// Send current game state to all players
	stateData := StateData{
		Board:   match.Board,
		Turn:    match.Turn,
		Size:    match.Size,
		Mode:    match.Mode,
		Players: match.Players,
	}

	stateBytes, _ := json.Marshal(stateData)
	dispatcher.BroadcastMessage(OpcodeState, stateBytes, nil, nil, true)

	return match
}

func (h *TTTMatchHandler) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	match := state.(*TTTMatch)

	// Remove players
	for _, presence := range presences {
		delete(match.Players, presence.GetUserId())
	}

	// If game was in progress, mark as finished
	if match.State == GameStatePlaying {
		match.State = GameStateFinished
		logger.Info("Match ended due to player leaving")
	}

	return match
}

func (h *TTTMatchHandler) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	match := state.(*TTTMatch)

	// Process messages
	for _, message := range messages {
		if message.GetOpCode() == OpcodeMove {
			h.handleMove(ctx, logger, nk, dispatcher, match, message)
		}
	}

	return match
}

func (h *TTTMatchHandler) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	match := state.(*TTTMatch)

	// Update leaderboard if game was finished
	if match.State == GameStateFinished && match.Winner != "" {
		h.updateLeaderboard(ctx, logger, nk, match)
	}

	logger.Info("Match terminated")
	return match
}

func (h *TTTMatchHandler) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, ""
}

// handleMove processes a move from a player
func (h *TTTMatchHandler) handleMove(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, match *TTTMatch, message runtime.MatchData) {
	// Check if game is in playing state
	if match.State != GameStatePlaying {
		h.sendError(dispatcher, "Game is not in playing state")
		return
	}

	// Parse move data
	var moveData MoveData
	if err := json.Unmarshal(message.GetData(), &moveData); err != nil {
		h.sendError(dispatcher, "Invalid move data")
		return
	}

	// Validate move coordinates
	if moveData.Row < 0 || moveData.Row >= match.Size || moveData.Col < 0 || moveData.Col >= match.Size {
		h.sendError(dispatcher, "Invalid move coordinates")
		return
	}

	// Check if it's the player's turn
	playerSymbol, exists := match.Players[message.GetUserId()]
	if !exists {
		h.sendError(dispatcher, "Player not in match")
		return
	}

	if playerSymbol != match.Turn {
		h.sendError(dispatcher, "Not your turn")
		return
	}

	// Check if cell is empty
	if match.Board[moveData.Row][moveData.Col] != Empty {
		h.sendError(dispatcher, "Cell already occupied")
		return
	}

	// Make the move
	match.Board[moveData.Row][moveData.Col] = playerSymbol
	match.MoveCount++

	// Check for win or draw
	winner := h.checkWinner(match)
	if winner != "" {
		match.Winner = winner
		match.State = GameStateFinished
		logger.Info("Game finished! Winner: %s", winner)
		
		// Update leaderboard immediately when game ends
		h.updateLeaderboard(ctx, logger, nk, match)
	} else if match.MoveCount >= match.Size*match.Size {
		match.State = GameStateFinished
		logger.Info("Game finished! Draw")
		
		// Update leaderboard immediately when game ends (draw)
		h.updateLeaderboard(ctx, logger, nk, match)
	} else {
		// Switch turns
		if match.Turn == PlayerX {
			match.Turn = PlayerO
		} else {
			match.Turn = PlayerX
		}
	}

	// Broadcast updated state
	stateData := StateData{
		Board:   match.Board,
		Turn:    match.Turn,
		Winner:  match.Winner,
		Size:    match.Size,
		Mode:    match.Mode,
		Players: match.Players,
	}

	stateBytes, _ := json.Marshal(stateData)
	dispatcher.BroadcastMessage(OpcodeState, stateBytes, nil, nil, true)
}

// checkWinner checks if there's a winner
func (h *TTTMatchHandler) checkWinner(match *TTTMatch) string {
	size := match.Size

	// Check rows
	for i := 0; i < size; i++ {
		if match.Board[i][0] != Empty {
			won := true
			for j := 1; j < size; j++ {
				if match.Board[i][j] != match.Board[i][0] {
					won = false
					break
				}
			}
			if won {
				return match.Board[i][0]
			}
		}
	}

	// Check columns
	for j := 0; j < size; j++ {
		if match.Board[0][j] != Empty {
			won := true
			for i := 1; i < size; i++ {
				if match.Board[i][j] != match.Board[0][j] {
					won = false
					break
				}
			}
			if won {
				return match.Board[0][j]
			}
		}
	}

	// Check main diagonal
	if match.Board[0][0] != Empty {
		won := true
		for i := 1; i < size; i++ {
			if match.Board[i][i] != match.Board[0][0] {
				won = false
				break
			}
		}
		if won {
			return match.Board[0][0]
		}
	}

	// Check anti-diagonal
	if match.Board[0][size-1] != Empty {
		won := true
		for i := 1; i < size; i++ {
			if match.Board[i][size-1-i] != match.Board[0][size-1] {
				won = false
				break
			}
		}
		if won {
			return match.Board[0][size-1]
		}
	}

	return ""
}

// updateLeaderboard updates the leaderboard with game results
func (h *TTTMatchHandler) updateLeaderboard(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, match *TTTMatch) {
	for userID, symbol := range match.Players {
		// Determine score based on game result
		score := int64(0)
		won := false
		lost := false
		drawn := false

		if match.Winner == symbol {
			score = 10 // Win
			won = true
		} else if match.Winner == "" {
			score = 1 // Draw
			drawn = true
		} else {
			score = -5 // Loss
			lost = true
		}

		// Update leaderboard
		err := UpdateLeaderboard(ctx, logger, nk, userID, score)
		if err != nil {
			logger.Error("Failed to update leaderboard for user %s: %v", userID, err)
		}

		// Update user statistics
		err = UpdateUserStats(ctx, logger, nk, userID, won, lost, drawn)
		if err != nil {
			logger.Error("Failed to update user stats for user %s: %v", userID, err)
		}
	}

	logger.Info("Updated leaderboard and stats for match %s", match.ID)
}

// sendError sends an error message to all players
func (h *TTTMatchHandler) sendError(dispatcher runtime.MatchDispatcher, message string) {
	errorData := ErrorData{Msg: message}
	errorBytes, _ := json.Marshal(errorData)
	dispatcher.BroadcastMessage(OpcodeError, errorBytes, nil, nil, true)
}
