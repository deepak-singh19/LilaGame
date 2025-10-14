package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

// MatchmakingRequest represents a matchmaking request
type MatchmakingRequest struct {
	Mode string `json:"mode"`
}

// MatchmakingResponse represents matchmaking response
type MatchmakingResponse struct {
	Ticket string `json:"ticket"`
	Mode   string `json:"mode"`
}

// MatchmakingQueue represents a player waiting for a match
type MatchmakingQueue struct {
	UserID    string
	Mode      string
	Timestamp time.Time
}

// Global matchmaking queue
var (
	matchmakingQueue = make(map[string]*MatchmakingQueue)
	queueMutex       sync.RWMutex
)

// InitMatchmaking initializes matchmaking system
func InitMatchmaking(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	// Register matchmaking RPC
	if err := initializer.RegisterRpc("start_matchmaking", startMatchmakingRPC); err != nil {
		return fmt.Errorf("failed to register start_matchmaking RPC: %w", err)
	}

	if err := initializer.RegisterRpc("stop_matchmaking", stopMatchmakingRPC); err != nil {
		return fmt.Errorf("failed to register stop_matchmaking RPC: %w", err)
	}

	// Register matchmaker matched handler
	if err := initializer.RegisterMatchmakerMatched(func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error) {
		return handleMatchmakerMatched(ctx, logger, nk, entries)
	}); err != nil {
		return fmt.Errorf("failed to register matchmaker matched handler: %w", err)
	}

	logger.Info("Matchmaking system initialized")
	return nil
}

// startMatchmakingRPC starts the matchmaking process
func startMatchmakingRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var request MatchmakingRequest
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		return "", fmt.Errorf("invalid request format: %w", err)
	}

	// Validate game mode
	if request.Mode != GameModeClassic && request.Mode != GameModeAdvanced {
		request.Mode = GameModeClassic // Default to classic
	}

	// Get user ID from context
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		return "", fmt.Errorf("user not authenticated")
	}

	// Add player to matchmaking queue
	queueMutex.Lock()
	defer queueMutex.Unlock()

	// Check if there's already a player waiting for the same mode
	var opponent *MatchmakingQueue
	for _, queuedPlayer := range matchmakingQueue {
		if queuedPlayer.Mode == request.Mode && queuedPlayer.UserID != userID {
			opponent = queuedPlayer
			break
		}
	}

	if opponent != nil {
		// Found an opponent! Create a match
		logger.Info("Found opponent for user %s: %s, mode: %s", userID, opponent.UserID, request.Mode)

		// Remove both players from queue
		delete(matchmakingQueue, opponent.UserID)

		// Create a match
		matchID, err := nk.MatchCreate(ctx, "ttt_match", map[string]interface{}{
			"mode": request.Mode,
		})
		if err != nil {
			logger.Error("Failed to create match: %v", err)
			// Add current player to queue as fallback
			matchmakingQueue[userID] = &MatchmakingQueue{
				UserID:    userID,
				Mode:      request.Mode,
				Timestamp: time.Now(),
			}
			ticket := fmt.Sprintf("ticket_%s_%d", userID, time.Now().Unix())
			response := MatchmakingResponse{
				Ticket: ticket,
				Mode:   request.Mode,
			}
			responseBytes, _ := json.Marshal(response)
			return string(responseBytes), nil
		}

		logger.Info("Created match %s for users %s and %s", matchID, userID, opponent.UserID)

		// Send notification to the opponent player about the match creation
		notification := map[string]interface{}{
			"type":     "match_created",
			"match_id": matchID,
			"mode":     request.Mode,
		}

		// Send notification to opponent
		notificationSend := &runtime.NotificationSend{
			UserID:     opponent.UserID,
			Subject:    "Match Created",
			Content:    notification,
			Code:       1,
			Persistent: true,
		}

		if err := nk.NotificationsSend(ctx, []*runtime.NotificationSend{notificationSend}); err != nil {
			logger.Error("Failed to send notification to opponent: %v", err)
		} else {
			logger.Info("Sent match creation notification to opponent %s", opponent.UserID)
		}

		// Return match info to current player
		response := MatchmakingResponse{
			Ticket: matchID,
			Mode:   request.Mode,
		}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			return "", fmt.Errorf("failed to marshal response: %w", err)
		}
		return string(responseBytes), nil
	} else {
		// No opponent found, add to queue
		matchmakingQueue[userID] = &MatchmakingQueue{
			UserID:    userID,
			Mode:      request.Mode,
			Timestamp: time.Now(),
		}

		ticket := fmt.Sprintf("ticket_%s_%d", userID, time.Now().Unix())
		logger.Info("Added user %s to matchmaking queue for mode %s, ticket: %s", userID, request.Mode, ticket)

		response := MatchmakingResponse{
			Ticket: ticket,
			Mode:   request.Mode,
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			return "", fmt.Errorf("failed to marshal response: %w", err)
		}

		logger.Info("User %s started matchmaking for mode %s, ticket: %s", userID, request.Mode, ticket)
		return string(responseBytes), nil
	}
}

// stopMatchmakingRPC stops the matchmaking process
func stopMatchmakingRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var request struct {
		Ticket string `json:"ticket"`
	}
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		return "", fmt.Errorf("invalid request format: %w", err)
	}

	// Get user ID from context
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		return "", fmt.Errorf("user not authenticated")
	}

	// Remove player from matchmaking queue
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if _, exists := matchmakingQueue[userID]; exists {
		delete(matchmakingQueue, userID)
		logger.Info("Removed user %s from matchmaking queue, ticket: %s", userID, request.Ticket)
	} else {
		logger.Info("User %s was not in matchmaking queue, ticket: %s", userID, request.Ticket)
	}

	logger.Info("User %s stopped matchmaking for ticket: %s", userID, request.Ticket)
	return `{"success": true}`, nil
}

// handleMatchmakerMatched handles when matchmaking finds a match
func handleMatchmakerMatched(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error) {
	if len(entries) != 2 {
		return "", fmt.Errorf("expected exactly 2 players, got %d", len(entries))
	}

	// Determine game mode from matchmaker properties
	mode := GameModeClassic
	props := entries[0].GetProperties()
	if modeProp, ok := props["mode"]; ok {
		if modeStr, ok := modeProp.(string); ok {
			mode = modeStr
		}
	}

	// Create match
	matchID, err := nk.MatchCreate(ctx, "ttt_match", map[string]interface{}{
		"mode": mode,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create match: %w", err)
	}

	logger.Info("Created match %s for mode %s with players: %s, %s",
		matchID, mode, entries[0].GetPresence().GetUserId(), entries[1].GetPresence().GetUserId())

	return matchID, nil
}

// GetMatchmakingStatus returns current matchmaking status
func GetMatchmakingStatus(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, userID string) ([]runtime.MatchmakerEntry, error) {
	// Return empty since matchmaker API not available
	logger.Info("Getting matchmaking status for user: %s", userID)
	return []runtime.MatchmakerEntry{}, nil
}

// CreateCustomMatch creates a custom match for testing
func CreateCustomMatch(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, mode string) (string, error) {
	matchID, err := nk.MatchCreate(ctx, "ttt_match", map[string]interface{}{
		"mode": mode,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create custom match: %w", err)
	}

	logger.Info("Created custom match %s for mode %s", matchID, mode)
	return matchID, nil
}

// GetActiveMatches returns currently active matches
func GetActiveMatches(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) ([]*api.Match, error) {
	matches, err := nk.MatchList(ctx, 10, true, "", nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list matches: %w", err)
	}

	return matches, nil
}

// JoinMatch allows a player to join a specific match
func JoinMatch(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, matchID, userID string) error {
	// This would typically be handled through the client SDK
	// For now, we'll just log the attempt
	logger.Info("User %s attempting to join match %s", userID, matchID)
	return nil
}

// GetMatchInfo returns information about a specific match
func GetMatchInfo(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, matchID string) (*api.Match, error) {
	matches, err := nk.MatchList(ctx, 1, true, matchID, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get match info: %w", err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("match not found")
	}

	return matches[0], nil
}
