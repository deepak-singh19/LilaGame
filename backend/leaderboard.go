package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/heroiclabs/nakama-common/runtime"
)

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	UserID     string  `json:"user_id"`
	Username   string  `json:"username"`
	Score      int64   `json:"score"`
	Rank       int     `json:"rank"`
	GamesWon   int     `json:"games_won"`
	GamesLost  int     `json:"games_lost"`
	GamesDrawn int     `json:"games_drawn"`
	WinRate    float64 `json:"win_rate"`
}

// LeaderboardResponse represents leaderboard response
type LeaderboardResponse struct {
	Entries []LeaderboardEntry `json:"entries"`
	Total   int                `json:"total"`
}

// PlayerStats represents detailed player statistics
type PlayerStats struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	Score       int64   `json:"score"`
	Rank        int     `json:"rank"`
	GamesWon    int     `json:"games_won"`
	GamesLost   int     `json:"games_lost"`
	GamesDrawn  int     `json:"games_drawn"`
	GamesPlayed int     `json:"games_played"`
	WinRate     float64 `json:"win_rate"`
	CreatedAt   int64   `json:"created_at"`
}

// InitLeaderboard initializes the leaderboard system
func InitLeaderboard(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	// Register leaderboard RPCs
	if err := initializer.RegisterRpc("get_leaderboard", getLeaderboardRPC); err != nil {
		return fmt.Errorf("failed to register get_leaderboard RPC: %w", err)
	}

	if err := initializer.RegisterRpc("get_player_stats", getPlayerStatsRPC); err != nil {
		return fmt.Errorf("failed to register get_player_stats RPC: %w", err)
	}

	if err := initializer.RegisterRpc("get_weekly_leaderboard", getWeeklyLeaderboardRPC); err != nil {
		return fmt.Errorf("failed to register get_weekly_leaderboard RPC: %w", err)
	}

	// Register clear leaderboard RPC for testing
	if err := initializer.RegisterRpc("clear_leaderboards", clearLeaderboardsRPC); err != nil {
		return fmt.Errorf("failed to register clear_leaderboards RPC: %w", err)
	}

	// Create leaderboards
	if err := createLeaderboards(ctx, logger, nk); err != nil {
		return fmt.Errorf("failed to create leaderboards: %w", err)
	}

	logger.Info("Leaderboard system initialized")
	return nil
}

// createLeaderboards creates all necessary leaderboards
func createLeaderboards(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) error {
	// Create main leaderboard
	leaderboardID := "ttt_leaderboard"
	leaderboard, err := nk.LeaderboardsGetId(ctx, []string{leaderboardID})
	if err != nil {
		return fmt.Errorf("failed to check leaderboard: %w", err)
	}

	if len(leaderboard) == 0 {
		metadata := map[string]interface{}{
			"description": "Player Performance",
		}
		err = nk.LeaderboardCreate(ctx, leaderboardID, true, "desc", "incr", "0 0 * * 0", metadata, true)
		if err != nil {
			return fmt.Errorf("failed to create leaderboard: %w", err)
		}
		logger.Info("Created leaderboard: %s", leaderboardID)
	}

	// Create weekly leaderboard
	weeklyLeaderboardID := "ttt_weekly_leaderboard"
	weeklyLeaderboard, err := nk.LeaderboardsGetId(ctx, []string{weeklyLeaderboardID})
	if err != nil {
		return fmt.Errorf("failed to check weekly leaderboard: %w", err)
	}

	if len(weeklyLeaderboard) == 0 {
		metadata := map[string]interface{}{
			"description": "Weekly Player Performance",
		}
		// Weekly reset every Sunday at midnight
		err = nk.LeaderboardCreate(ctx, weeklyLeaderboardID, true, "desc", "incr", "0 0 * * 0", metadata, true)
		if err != nil {
			return fmt.Errorf("failed to create weekly leaderboard: %w", err)
		}
		logger.Info("Created weekly leaderboard: %s", weeklyLeaderboardID)
	}

	return nil
}

// getLeaderboardRPC returns the current leaderboard
func getLeaderboardRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var request struct {
		Limit int `json:"limit"`
	}
	if payload != "" {
		if err := json.Unmarshal([]byte(payload), &request); err != nil {
			request.Limit = 10 // Default limit
		}
	}
	if request.Limit <= 0 || request.Limit > 100 {
		request.Limit = 10
	}

	leaderboardID := "ttt_leaderboard"

	// Get leaderboard records
	records, _, _, _, err := nk.LeaderboardRecordsList(ctx, leaderboardID, nil, request.Limit, "", 0)
	if err != nil {
		return "", fmt.Errorf("failed to get leaderboard records: %w", err)
	}

	// Convert to our format
	entries := make([]LeaderboardEntry, len(records))
	for i, record := range records {
		// Get additional stats from user storage
		stats, err := getUserStats(ctx, nk, record.OwnerId)
		if err != nil {
			logger.Error("Failed to get stats for user %s: %v", record.OwnerId, err)
			stats = &PlayerStats{} // Use empty stats
		}

		winRate := 0.0
		if stats.GamesPlayed > 0 {
			winRate = float64(stats.GamesWon) / float64(stats.GamesPlayed) * 100
		}

		entries[i] = LeaderboardEntry{
			UserID:     record.OwnerId,
			Username:   record.Username.GetValue(),
			Score:      record.Score,
			Rank:       i + 1,
			GamesWon:   stats.GamesWon,
			GamesLost:  stats.GamesLost,
			GamesDrawn: stats.GamesDrawn,
			WinRate:    winRate,
		}
	}

	response := LeaderboardResponse{
		Entries: entries,
		Total:   len(entries),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal leaderboard response: %w", err)
	}

	return string(responseBytes), nil
}

// getPlayerStatsRPC returns detailed player statistics
func getPlayerStatsRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var request struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		return "", fmt.Errorf("invalid request format: %w", err)
	}

	// Get user's leaderboard record
	leaderboardID := "ttt_leaderboard"
	records, _, _, _, err := nk.LeaderboardRecordsList(ctx, leaderboardID, []string{request.UserID}, 1, "", 0)
	if err != nil {
		return "", fmt.Errorf("failed to get player record: %w", err)
	}

	var stats PlayerStats
	if len(records) > 0 {
		record := records[0]
		userStats, err := getUserStats(ctx, nk, record.OwnerId)
		if err != nil {
			return "", fmt.Errorf("failed to get user stats: %w", err)
		}

		winRate := 0.0
		if userStats.GamesPlayed > 0 {
			winRate = float64(userStats.GamesWon) / float64(userStats.GamesPlayed) * 100
		}

		stats = PlayerStats{
			UserID:      record.OwnerId,
			Username:    record.Username.GetValue(),
			Score:       record.Score,
			Rank:        1, // This would need to be calculated properly
			GamesWon:    userStats.GamesWon,
			GamesLost:   userStats.GamesLost,
			GamesDrawn:  userStats.GamesDrawn,
			GamesPlayed: userStats.GamesPlayed,
			WinRate:     winRate,
			CreatedAt:   userStats.CreatedAt,
		}
	}

	responseBytes, err := json.Marshal(stats)
	if err != nil {
		return "", fmt.Errorf("failed to marshal player stats: %w", err)
	}

	return string(responseBytes), nil
}

// getWeeklyLeaderboardRPC returns the weekly leaderboard
func getWeeklyLeaderboardRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var request struct {
		Limit int `json:"limit"`
	}
	if payload != "" {
		if err := json.Unmarshal([]byte(payload), &request); err != nil {
			request.Limit = 10
		}
	}
	if request.Limit <= 0 || request.Limit > 100 {
		request.Limit = 10
	}

	leaderboardID := "ttt_weekly_leaderboard"

	// Get weekly leaderboard records
	records, _, _, _, err := nk.LeaderboardRecordsList(ctx, leaderboardID, nil, request.Limit, "", 0)
	if err != nil {
		return "", fmt.Errorf("failed to get weekly leaderboard records: %w", err)
	}

	// Convert to our format
	entries := make([]LeaderboardEntry, len(records))
	for i, record := range records {
		entries[i] = LeaderboardEntry{
			UserID:   record.OwnerId,
			Username: record.Username.GetValue(),
			Score:    record.Score,
			Rank:     i + 1,
		}
	}

	response := LeaderboardResponse{
		Entries: entries,
		Total:   len(entries),
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal weekly leaderboard response: %w", err)
	}

	return string(responseBytes), nil
}

// clearLeaderboardsRPC clears all leaderboard data (for testing)
func clearLeaderboardsRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	// Delete all records from main leaderboard
	err := nk.LeaderboardDelete(ctx, "ttt_leaderboard")
	if err != nil {
		logger.Error("Failed to clear main leaderboard: %v", err)
	}

	// Delete all records from weekly leaderboard
	err = nk.LeaderboardDelete(ctx, "ttt_weekly_leaderboard")
	if err != nil {
		logger.Error("Failed to clear weekly leaderboard: %v", err)
	}

	// Recreate leaderboards
	if err := createLeaderboards(ctx, logger, nk); err != nil {
		return "", fmt.Errorf("failed to recreate leaderboards: %w", err)
	}

	logger.Info("Cleared and recreated all leaderboards")
	
	response := map[string]interface{}{
		"message": "Leaderboards cleared and recreated successfully",
		"success": true,
	}
	
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(responseBytes), nil
}

// getUserStats retrieves user statistics from storage
func getUserStats(ctx context.Context, nk runtime.NakamaModule, userID string) (*PlayerStats, error) {
	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{
		{
			Collection: "user_stats",
			Key:        "stats",
			UserID:     userID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read user stats: %w", err)
	}

	if len(objects) == 0 {
		return &PlayerStats{}, nil
	}

	var stats map[string]interface{}
	if err := json.Unmarshal([]byte(objects[0].Value), &stats); err != nil {
		return &PlayerStats{}, nil
	}

	playerStats := &PlayerStats{
		UserID: userID,
	}

	if username, ok := stats["username"].(string); ok {
		playerStats.Username = username
	}
	if gamesWon, ok := stats["games_won"].(float64); ok {
		playerStats.GamesWon = int(gamesWon)
	}
	if gamesLost, ok := stats["games_lost"].(float64); ok {
		playerStats.GamesLost = int(gamesLost)
	}
	if gamesDrawn, ok := stats["games_drawn"].(float64); ok {
		playerStats.GamesDrawn = int(gamesDrawn)
	}
	if gamesPlayed, ok := stats["games_played"].(float64); ok {
		playerStats.GamesPlayed = int(gamesPlayed)
	}
	if totalScore, ok := stats["total_score"].(float64); ok {
		playerStats.Score = int64(totalScore)
	}
	if createdAt, ok := stats["created_at"].(float64); ok {
		playerStats.CreatedAt = int64(createdAt)
	}

	return playerStats, nil
}

// UpdateLeaderboard updates leaderboard with game results
func UpdateLeaderboard(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, userID string, score int64) error {
	// Get user information including username
	users, err := nk.UsersGetId(ctx, []string{userID}, []string{})
	if err != nil {
		logger.Error("Failed to get user info for user %s: %v", userID, err)
		// Continue without username if we can't get it
	}

	username := ""
	if len(users) > 0 {
		username = users[0].Username
		logger.Info("Retrieved username for user %s: %s", userID, username)
	} else {
		logger.Warn("No user found for ID %s", userID)
	}

	// Update main leaderboard
	_, err = nk.LeaderboardRecordWrite(ctx, "ttt_leaderboard", userID, username, score, 0, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to update main leaderboard: %w", err)
	}

	// Update weekly leaderboard
	_, err = nk.LeaderboardRecordWrite(ctx, "ttt_weekly_leaderboard", userID, username, score, 0, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to update weekly leaderboard: %w", err)
	}

	logger.Info("Updated leaderboards for user %s (%s) with score %d", userID, username, score)
	return nil
}

// GetPlayerRank returns a player's current rank
func GetPlayerRank(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, userID string) (int, error) {
	leaderboardID := "ttt_leaderboard"
	records, _, _, _, err := nk.LeaderboardRecordsList(ctx, leaderboardID, []string{userID}, 1, "", 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get player rank: %w", err)
	}

	if len(records) == 0 {
		return 0, fmt.Errorf("player not found in leaderboard")
	}

	// Get all records to calculate rank
	allRecords, _, _, _, err := nk.LeaderboardRecordsList(ctx, leaderboardID, nil, 1000, "", 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get all records: %w", err)
	}

	// Sort by score to find rank
	sort.Slice(allRecords, func(i, j int) bool {
		return allRecords[i].Score > allRecords[j].Score
	})

	for i, record := range allRecords {
		if record.OwnerId == userID {
			return i + 1, nil
		}
	}

	return 0, fmt.Errorf("player rank not found")
}
