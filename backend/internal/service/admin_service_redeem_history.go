package service

import (
	"context"
	"database/sql"
	"fmt"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"sort"
	"strings"
	"time"
)

func isMissingLuckyWheelTableError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return (strings.Contains(message, "lucky_wheel_sessions") && strings.Contains(message, "does not exist")) ||
		strings.Contains(message, "no such table")
}

func (s *adminServiceImpl) listRedeemCodesForAdminMerge(ctx context.Context, needed int, codeType, status, search, sortBy, sortOrder string) ([]RedeemCode, int64, error) {
	batchSize := redeemCodeAdminMergeBatchSize(needed)
	codes := make([]RedeemCode, 0, needed)
	var total int64
	for page := 1; len(codes) < needed; page++ {
		params := pagination.PaginationParams{Page: page, PageSize: batchSize, SortBy: sortBy, SortOrder: sortOrder}
		batch, result, err := s.redeemCodeRepo.ListWithFilters(ctx, params, codeType, status, search)
		if err != nil {
			return nil, 0, err
		}
		if result != nil {
			total = result.Total
		}
		codes = append(codes, batch...)
		if len(batch) < params.Limit() || (total > 0 && int64(len(codes)) >= total) {
			break
		}
	}
	if len(codes) > needed {
		codes = codes[:needed]
	}
	return codes, total, nil
}

func redeemCodeAdminMergeBatchSize(needed int) int {
	if needed < 1 {
		return pagination.DefaultPagination().Limit()
	}
	if needed > 1000 {
		return 1000
	}
	return needed
}

func (s *adminServiceImpl) listAffiliateRedeemCodes(ctx context.Context, params pagination.PaginationParams, status, search string) ([]RedeemCode, int64, error) {
	if s == nil || s.entClient == nil {
		return nil, 0, nil
	}
	if status != "" && status != StatusUsed {
		return nil, 0, nil
	}

	search = strings.ToLower(strings.TrimSpace(search))
	where := "WHERE l.action = ?"
	args := []any{"transfer"}
	if search != "" {
		where += " AND (LOWER('aff-' || CAST(l.id AS TEXT)) LIKE ? OR LOWER(COALESCE(u.email, '')) LIKE ? OR LOWER(COALESCE(u.username, '')) LIKE ?)"
		like := "%" + search + "%"
		args = append(args, like, like, like)
	}

	countArgs := append([]any{}, args...)
	countRows, err := s.entClient.QueryContext(ctx, luckyWheelBindVars(s.entClient.Driver().Dialect(), `
SELECT COUNT(*)
FROM user_affiliate_ledger l
LEFT JOIN users u ON u.id = l.user_id
`+where), countArgs...)
	if err != nil {
		return nil, 0, err
	}
	var total sql.NullInt64
	if countRows.Next() {
		if err := countRows.Scan(&total); err != nil {
			_ = countRows.Close()
			return nil, 0, err
		}
	}
	if err := countRows.Err(); err != nil {
		_ = countRows.Close()
		return nil, 0, err
	}
	_ = countRows.Close()

	orderBy := affiliateRedeemCodeOrderBy(params.SortBy, params.SortOrder)
	args = append(args, params.Limit(), params.Offset())
	rows, err := s.entClient.QueryContext(ctx, luckyWheelBindVars(s.entClient.Driver().Dialect(), `
SELECT l.id,
       CAST(COALESCE(l.platform_amount, l.amount) AS DOUBLE PRECISION) AS display_amount,
       CAST(l.platform_amount AS DOUBLE PRECISION),
       l.created_at,
       l.user_id,
       u.email,
       u.username
FROM user_affiliate_ledger l
LEFT JOIN users u ON u.id = l.user_id
`+where+`
`+orderBy+`
LIMIT ?
OFFSET ?`), args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	codes := make([]RedeemCode, 0, params.Limit())
	for rows.Next() {
		var (
			id             int64
			value          float64
			platformAmount sql.NullFloat64
			createdAt      time.Time
			userID         int64
			email          sql.NullString
			username       sql.NullString
		)
		if err := rows.Scan(&id, &value, &platformAmount, &createdAt, &userID, &email, &username); err != nil {
			return nil, 0, err
		}
		usedAt := createdAt
		userSummary := &User{ID: userID}
		if email.Valid {
			userSummary.Email = email.String
		}
		if username.Valid {
			userSummary.Username = username.String
		}
		code := RedeemCode{
			ID:        -id,
			Code:      fmt.Sprintf("AFF-%d", id),
			Type:      RedeemTypeAffiliateBalance,
			Value:     value,
			Status:    StatusUsed,
			UsedBy:    &userID,
			UsedAt:    &usedAt,
			CreatedAt: createdAt,
			User:      userSummary,
			Source:    "affiliate_transfer",
		}
		if platformAmount.Valid {
			code.PlatformAmount = &platformAmount.Float64
		}
		codes = append(codes, code)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if !total.Valid {
		return codes, 0, nil
	}
	return codes, total.Int64, nil
}

func (s *adminServiceImpl) listLuckyWheelRedeemCodes(ctx context.Context, params pagination.PaginationParams, status, search string) ([]RedeemCode, int64, error) {
	if s == nil || s.entClient == nil {
		return nil, 0, nil
	}
	if status != "" && status != StatusUsed {
		return nil, 0, nil
	}

	search = strings.ToLower(strings.TrimSpace(search))
	where := "WHERE s.settled = ? AND s.settled_bonus_amount IS NOT NULL AND s.settled_bonus_amount > 0"
	args := []any{true}
	if search != "" {
		where += " AND (LOWER('lucky-' || CAST(s.source_order_id AS TEXT)) LIKE ? OR LOWER(COALESCE(s.matched_tier_name, '')) LIKE ? OR LOWER(COALESCE(u.email, '')) LIKE ?)"
		like := "%" + search + "%"
		args = append(args, like, like, like)
	}

	countArgs := append([]any{}, args...)
	countQuery := luckyWheelBindVars(s.entClient.Driver().Dialect(), `
SELECT COUNT(*)
FROM lucky_wheel_sessions s
LEFT JOIN users u ON u.id = s.user_id
`+where)
	countRows, err := s.entClient.QueryContext(ctx, countQuery, countArgs...)
	if err != nil {
		if isMissingLuckyWheelTableError(err) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	var total sql.NullInt64
	if countRows.Next() {
		if err := countRows.Scan(&total); err != nil {
			_ = countRows.Close()
			return nil, 0, err
		}
	}
	if err := countRows.Err(); err != nil {
		_ = countRows.Close()
		return nil, 0, err
	}
	_ = countRows.Close()

	orderBy := luckyWheelRedeemCodeOrderBy(params.SortBy, params.SortOrder)
	args = append(args, params.Limit(), params.Offset())
	query := luckyWheelBindVars(s.entClient.Driver().Dialect(), `
SELECT s.id, s.source_order_id, s.settled_bonus_amount, s.settled_at, s.created_at, s.matched_tier_name,
       s.user_id, u.email, u.username
FROM lucky_wheel_sessions s
LEFT JOIN users u ON u.id = s.user_id
`+where+`
`+orderBy+`
LIMIT ?
OFFSET ?`)
	rows, err := s.entClient.QueryContext(ctx, query, args...)
	if err != nil {
		if isMissingLuckyWheelTableError(err) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	codes := make([]RedeemCode, 0, params.Limit())
	for rows.Next() {
		var (
			id            int64
			sourceOrderID int64
			value         float64
			usedAt        time.Time
			createdAt     time.Time
			tierName      sql.NullString
			userID        int64
			email         sql.NullString
			username      sql.NullString
		)
		if err := rows.Scan(&id, &sourceOrderID, &value, &usedAt, &createdAt, &tierName, &userID, &email, &username); err != nil {
			return nil, 0, err
		}
		usedAtCopy := usedAt
		userSummary := &User{ID: userID}
		if email.Valid {
			userSummary.Email = email.String
		}
		if username.Valid {
			userSummary.Username = username.String
		}
		notes := ""
		if tierName.Valid {
			notes = tierName.String
		}
		codes = append(codes, RedeemCode{
			ID:        -id,
			Code:      fmt.Sprintf("LUCKY-%d", sourceOrderID),
			Type:      RedeemTypeLuckyWheelBonus,
			Value:     value,
			Status:    StatusUsed,
			UsedBy:    &userID,
			UsedAt:    &usedAtCopy,
			Notes:     notes,
			CreatedAt: createdAt,
			User:      userSummary,
			Source:    "lucky_wheel",
		})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if !total.Valid {
		return codes, 0, nil
	}
	return codes, total.Int64, nil
}

func normalizeRedeemCodeMergeSortBy(sortBy string) string {
	sortBy = strings.ToLower(strings.TrimSpace(sortBy))
	if sortBy == "" || sortBy == "id" {
		return "used_at"
	}
	return sortBy
}

func (s *adminServiceImpl) listLuckyWheelRedeemCodesForAdminMerge(ctx context.Context, needed int, status, search, sortBy, sortOrder string) ([]RedeemCode, int64, error) {
	batchSize := redeemCodeAdminMergeBatchSize(needed)
	codes := make([]RedeemCode, 0, needed)
	var total int64
	for page := 1; len(codes) < needed; page++ {
		params := pagination.PaginationParams{Page: page, PageSize: batchSize, SortBy: sortBy, SortOrder: sortOrder}
		batch, batchTotal, err := s.listLuckyWheelRedeemCodes(ctx, params, status, search)
		if err != nil {
			return nil, 0, err
		}
		total = batchTotal
		codes = append(codes, batch...)
		if len(batch) < params.Limit() || (total > 0 && int64(len(codes)) >= total) {
			break
		}
	}
	if len(codes) > needed {
		codes = codes[:needed]
	}
	return codes, total, nil
}

func mergeRedeemCodeAdminList(redeemCodes, affiliateCodes, luckyWheelCodes []RedeemCode, params pagination.PaginationParams, sortBy, sortOrder string) []RedeemCode {
	combined := append(append(append([]RedeemCode{}, redeemCodes...), affiliateCodes...), luckyWheelCodes...)
	sort.SliceStable(combined, func(i, j int) bool {
		cmp := compareRedeemCodeForAdminList(combined[i], combined[j], sortBy)
		if cmp == 0 {
			cmp = compareInt64(combined[i].ID, combined[j].ID)
		}
		if strings.ToLower(strings.TrimSpace(sortOrder)) == pagination.SortOrderAsc {
			return cmp < 0
		}
		return cmp > 0
	})
	offset := params.Offset()
	if offset >= len(combined) {
		return []RedeemCode{}
	}
	end := offset + params.Limit()
	if end > len(combined) {
		end = len(combined)
	}
	return combined[offset:end]
}

func (s *adminServiceImpl) listLuckyWheelBalanceHistory(ctx context.Context, userID int64, params pagination.PaginationParams) ([]RedeemCode, int64, error) {
	if s == nil || s.entClient == nil || userID <= 0 {
		return nil, 0, nil
	}

	rows, err := s.entClient.QueryContext(ctx, luckyWheelBindVars(s.entClient.Driver().Dialect(), `
SELECT id, source_order_id, settled_bonus_amount, settled_at, created_at, matched_tier_name
FROM lucky_wheel_sessions
WHERE user_id = ? AND settled = ? AND settled_bonus_amount IS NOT NULL AND settled_bonus_amount > 0
ORDER BY COALESCE(settled_at, updated_at, created_at) DESC, id DESC
LIMIT ?
OFFSET ?`), userID, true, params.Limit(), params.Offset())
	if err != nil {
		if isMissingLuckyWheelTableError(err) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	codes := make([]RedeemCode, 0, params.Limit())
	for rows.Next() {
		var (
			id            int64
			sourceOrderID int64
			value         float64
			usedAt        time.Time
			createdAt     time.Time
			tierName      sql.NullString
		)
		if err := rows.Scan(&id, &sourceOrderID, &value, &usedAt, &createdAt, &tierName); err != nil {
			return nil, 0, err
		}
		usedAtCopy := usedAt
		notes := ""
		if tierName.Valid {
			notes = tierName.String
		}
		codes = append(codes, RedeemCode{
			ID:        -id,
			Code:      fmt.Sprintf("LUCKY-%d", sourceOrderID),
			Type:      RedeemTypeLuckyWheelBonus,
			Value:     value,
			Status:    StatusUsed,
			UsedBy:    &userID,
			UsedAt:    &usedAtCopy,
			Notes:     notes,
			CreatedAt: createdAt,
			Source:    "lucky_wheel",
		})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	total, err := countLuckyWheelBalanceHistory(ctx, s.entClient, userID)
	if err != nil {
		return nil, 0, err
	}
	return codes, total, nil
}

func compareInt64(left, right int64) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func (s *adminServiceImpl) listAffiliateRedeemCodesForAdminMerge(ctx context.Context, needed int, status, search, sortBy, sortOrder string) ([]RedeemCode, int64, error) {
	batchSize := redeemCodeAdminMergeBatchSize(needed)
	codes := make([]RedeemCode, 0, needed)
	var total int64
	for page := 1; len(codes) < needed; page++ {
		params := pagination.PaginationParams{Page: page, PageSize: batchSize, SortBy: sortBy, SortOrder: sortOrder}
		batch, batchTotal, err := s.listAffiliateRedeemCodes(ctx, params, status, search)
		if err != nil {
			return nil, 0, err
		}
		total = batchTotal
		codes = append(codes, batch...)
		if len(batch) < params.Limit() || (total > 0 && int64(len(codes)) >= total) {
			break
		}
	}
	if len(codes) > needed {
		codes = codes[:needed]
	}
	return codes, total, nil
}

func luckyWheelRedeemCodeOrderBy(sortBy, sortOrder string) string {
	desc := strings.ToLower(strings.TrimSpace(sortOrder)) != pagination.SortOrderAsc
	direction := "DESC"
	if !desc {
		direction = "ASC"
	}

	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "type":
		return "ORDER BY s.id " + direction
	case "value":
		return "ORDER BY s.settled_bonus_amount " + direction + ", s.id " + direction
	case "status":
		return "ORDER BY s.id " + direction
	case "used_at":
		return "ORDER BY s.settled_at " + direction + ", s.id " + direction
	case "created_at":
		return "ORDER BY s.created_at " + direction + ", s.id " + direction
	case "code":
		return "ORDER BY s.source_order_id " + direction + ", s.id " + direction
	default:
		return "ORDER BY s.id " + direction
	}
}

func compareRedeemCodeForAdminList(left, right RedeemCode, sortBy string) int {
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "type":
		return strings.Compare(left.Type, right.Type)
	case "value":
		return compareFloat64(left.Value, right.Value)
	case "status":
		return strings.Compare(left.Status, right.Status)
	case "used_at":
		return compareTime(redeemCodeHistoryTime(left), redeemCodeHistoryTime(right))
	case "created_at":
		return compareTime(left.CreatedAt, right.CreatedAt)
	case "code":
		return strings.Compare(left.Code, right.Code)
	default:
		return compareInt64(left.ID, right.ID)
	}
}

func countLuckyWheelBalanceHistory(ctx context.Context, client *dbent.Client, userID int64) (int64, error) {
	rows, err := client.QueryContext(ctx, luckyWheelBindVars(client.Driver().Dialect(), `
SELECT COUNT(*)
FROM lucky_wheel_sessions
WHERE user_id = ? AND settled = ? AND settled_bonus_amount IS NOT NULL AND settled_bonus_amount > 0`), userID, true)
	if err != nil {
		if isMissingLuckyWheelTableError(err) {
			return 0, nil
		}
		return 0, err
	}
	defer func() { _ = rows.Close() }()

	var total sql.NullInt64
	if rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, err
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if !total.Valid {
		return 0, nil
	}
	return total.Int64, nil
}

func (s *adminServiceImpl) getUserTotalRecharged(ctx context.Context, userID int64) (float64, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, ErrUserNotFound
	}
	return user.TotalRecharged, nil
}

func (s *adminServiceImpl) listLuckyWheelBalanceHistoryForMerge(ctx context.Context, userID int64, needed int) ([]RedeemCode, int64, error) {
	if needed <= 0 {
		return nil, 0, nil
	}
	return s.listLuckyWheelBalanceHistory(ctx, userID, pagination.PaginationParams{Page: 1, PageSize: needed})
}

func affiliateRedeemCodeOrderBy(sortBy, sortOrder string) string {
	desc := strings.ToLower(strings.TrimSpace(sortOrder)) != pagination.SortOrderAsc
	direction := "DESC"
	if !desc {
		direction = "ASC"
	}

	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "type":
		return "ORDER BY l.id " + direction
	case "value":
		return "ORDER BY display_amount " + direction + ", l.id " + direction
	case "status":
		return "ORDER BY l.id " + direction
	case "used_at", "created_at":
		return "ORDER BY l.created_at " + direction + ", l.id " + direction
	case "code":
		return "ORDER BY l.id " + direction
	default:
		return "ORDER BY l.id " + direction
	}
}

func compareTime(left, right time.Time) int {
	if left.Before(right) {
		return -1
	}
	if left.After(right) {
		return 1
	}
	return 0
}

func compareFloat64(left, right float64) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
