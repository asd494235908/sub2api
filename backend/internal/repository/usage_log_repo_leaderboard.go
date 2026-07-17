package repository

import (
	"context"
	"fmt"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/lib/pq"
	"time"
)

func (r *usageLogRepository) GetUserTokenLeaderboard(ctx context.Context, startTime, endTime time.Time, limit int, ignoredUserIDs []int64) (result *usagestats.UserTokenLeaderboardResponse, err error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		WITH user_token_usage AS (
			SELECT
				u.user_id,
				COALESCE(us.email, '') as email,
				COALESCE(us.username, '') as username,
				COALESCE(SUM(u.input_tokens + u.output_tokens + u.cache_creation_tokens + u.cache_read_tokens), 0) as tokens,
				COUNT(*) as requests,
				COALESCE(SUM(u.actual_cost), 0) as actual_cost
			FROM usage_logs u
			LEFT JOIN users us ON u.user_id = us.id
			WHERE u.created_at >= $1 AND u.created_at < $2
	`
	args := []any{startTime, endTime}
	if len(ignoredUserIDs) > 0 {
		query += fmt.Sprintf(" AND u.user_id <> ALL($%d)", len(args)+1)
		args = append(args, pq.Array(ignoredUserIDs))
	}
	query += fmt.Sprintf(`
			GROUP BY u.user_id, us.email, us.username
		),
		ranked AS (
			SELECT
				ROW_NUMBER() OVER (ORDER BY tokens DESC, requests DESC, user_id ASC) AS rank,
				user_id,
				email,
				username,
				tokens,
				requests,
				actual_cost,
				COALESCE(SUM(tokens) OVER (), 0) as total_tokens,
				COALESCE(SUM(requests) OVER (), 0) as total_requests
			FROM user_token_usage
			ORDER BY tokens DESC, requests DESC, user_id ASC
			LIMIT $%d
		)
		SELECT
			rank,
			user_id,
			email,
			username,
			tokens,
			requests,
			actual_cost,
			total_tokens,
			total_requests
		FROM ranked
		ORDER BY rank ASC
	`, len(args)+1)
	args = append(args, limit)

	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			result = nil
		}
	}()

	ranking := make([]usagestats.UserTokenLeaderboardItem, 0)
	totalTokens := int64(0)
	totalRequests := int64(0)
	for rows.Next() {
		var row usagestats.UserTokenLeaderboardItem
		if err = rows.Scan(&row.Rank, &row.UserID, &row.Email, &row.Username, &row.Tokens, &row.Requests, &row.ActualCost, &totalTokens, &totalRequests); err != nil {
			return nil, err
		}
		ranking = append(ranking, row)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &usagestats.UserTokenLeaderboardResponse{
		Ranking:       ranking,
		TotalTokens:   totalTokens,
		TotalRequests: totalRequests,
	}, nil
}
