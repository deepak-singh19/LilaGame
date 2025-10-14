package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
)

// DeviceAuthRequest represents device authentication request
type DeviceAuthRequest struct {
	DeviceID string `json:"device_id"`
	Username string `json:"username,omitempty"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Created  bool   `json:"created"`
}

// InitAuth initializes authentication hooks
func InitAuth(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	// Register device authentication RPC
	if err := initializer.RegisterRpc("device_auth", deviceAuthRPC); err != nil {
		return fmt.Errorf("failed to register device_auth RPC: %w", err)
	}

	// Register before hook for authentication
	if err := initializer.RegisterBeforeRt("MatchmakerAdd", beforeMatchmakerAdd); err != nil {
		return fmt.Errorf("failed to register beforeMatchmakerAdd hook: %w", err)
	}

	// Register after hook for successful authentication
	if err := initializer.RegisterAfterRt("AuthenticateDevice", afterDeviceAuth); err != nil {
		return fmt.Errorf("failed to register afterDeviceAuth hook: %w", err)
	}

	logger.Info("Authentication system initialized")
	return nil
}

// deviceAuthRPC handles device-based authentication
func deviceAuthRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var request DeviceAuthRequest
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		return "", fmt.Errorf("invalid request format: %w", err)
	}

	if request.DeviceID == "" {
		return "", fmt.Errorf("device_id is required")
	}

	// Generate username if not provided
	username := request.Username
	if username == "" {
		username = fmt.Sprintf("Player_%s", request.DeviceID[:8])
	}

	// Authenticate with device ID
	userID, username, created, err := nk.AuthenticateDevice(ctx, request.DeviceID, username, true)
	if err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// For now, return user info without JWT token
	response := AuthResponse{
		Token:    "", // JWT token generation can be added later
		UserID:   userID,
		Username: username,
		Created:  created,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	logger.Info("Device authenticated: userID=%s, username=%s, created=%v", userID, username, created)
	return string(responseBytes), nil
}

// beforeMatchmakerAdd validates authentication before matchmaking
func beforeMatchmakerAdd(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, envelope *rtapi.Envelope) (*rtapi.Envelope, error) {
	// Check if user is authenticated
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return nil, runtime.NewError("authentication required", 3)
	}

	// Validate matchmaker properties
	if envelope.GetMatchmakerAdd() != nil {
		query := envelope.GetMatchmakerAdd().Query
		if query == "" {
			// Set default query for classic mode
			envelope.GetMatchmakerAdd().Query = "*"
		}

		// Add mode property if not present
		if envelope.GetMatchmakerAdd().StringProperties == nil {
			envelope.GetMatchmakerAdd().StringProperties = make(map[string]string)
		}
		if _, exists := envelope.GetMatchmakerAdd().StringProperties["mode"]; !exists {
			envelope.GetMatchmakerAdd().StringProperties["mode"] = GameModeClassic
		}
	}

	return envelope, nil
}

// afterDeviceAuth handles post-authentication setup
func afterDeviceAuth(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out *rtapi.Envelope, in *rtapi.Envelope) error {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok {
		return nil
	}

	username, _ := ctx.Value(runtime.RUNTIME_CTX_USERNAME).(string)

	// Initialize user statistics
	err := initializeUserStats(ctx, logger, nk, userID, username)
	if err != nil {
		logger.Error("Failed to initialize user stats: %v", err)
	}

	return nil
}

// initializeUserStats sets up initial user statistics
func initializeUserStats(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, userID, username string) error {
	// Create user storage object for statistics
	objects := []*runtime.StorageWrite{
		{
			Collection: "user_stats",
			Key:        "stats",
			UserID:     userID,
			Value: `{
				"games_played": 0,
				"games_won": 0,
				"games_lost": 0,
				"games_drawn": 0,
				"total_score": 0,
				"created_at": ` + fmt.Sprintf("%d", time.Now().Unix()) + `,
				"username": "` + username + `"
			}`,
		},
	}

	_, err := nk.StorageWrite(ctx, objects)
	if err != nil {
		return fmt.Errorf("failed to create user stats: %w", err)
	}

	logger.Info("Initialized stats for user %s (%s)", username, userID)
	return nil
}

// UpdateUserStats updates user statistics after a game
func UpdateUserStats(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, userID string, won, lost, drawn bool) error {
	// Read current stats
	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{
		{
			Collection: "user_stats",
			Key:        "stats",
			UserID:     userID,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to read user stats: %w", err)
	}

	var stats map[string]interface{}
	if len(objects) > 0 {
		// Parse JSON value
		if err := json.Unmarshal([]byte(objects[0].Value), &stats); err != nil {
			stats = make(map[string]interface{})
		}
	} else {
		stats = make(map[string]interface{})
	}

	// Update stats
	if gamesPlayed, ok := stats["games_played"].(float64); ok {
		stats["games_played"] = gamesPlayed + 1
	} else {
		stats["games_played"] = 1
	}

	// Get current total score
	var totalScore float64
	if ts, ok := stats["total_score"].(float64); ok {
		totalScore = ts
	}

	if won {
		if gamesWon, ok := stats["games_won"].(float64); ok {
			stats["games_won"] = gamesWon + 1
		} else {
			stats["games_won"] = 1
		}
		stats["total_score"] = totalScore + 10
	} else if lost {
		if gamesLost, ok := stats["games_lost"].(float64); ok {
			stats["games_lost"] = gamesLost + 1
		} else {
			stats["games_lost"] = 1
		}
		stats["total_score"] = totalScore - 5
	} else if drawn {
		if gamesDrawn, ok := stats["games_drawn"].(float64); ok {
			stats["games_drawn"] = gamesDrawn + 1
		} else {
			stats["games_drawn"] = 1
		}
		stats["total_score"] = totalScore + 1
	}

	// Convert stats to JSON
	statsJSON, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	// Write updated stats
	_, err = nk.StorageWrite(ctx, []*runtime.StorageWrite{
		{
			Collection: "user_stats",
			Key:        "stats",
			UserID:     userID,
			Value:      string(statsJSON),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}

	return nil
}
