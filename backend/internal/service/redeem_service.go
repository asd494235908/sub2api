package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/redeemcode"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrRedeemCodeNotFound  = infraerrors.NotFound("REDEEM_CODE_NOT_FOUND", "redeem code not found")
	ErrRedeemCodeUsed      = infraerrors.Conflict("REDEEM_CODE_USED", "redeem code already used")
	ErrInsufficientBalance = infraerrors.BadRequest("INSUFFICIENT_BALANCE", "insufficient balance")
	ErrRedeemRateLimited   = infraerrors.TooManyRequests("REDEEM_RATE_LIMITED", "too many failed attempts, please try again later")
	ErrRedeemCodeLocked    = infraerrors.Conflict("REDEEM_CODE_LOCKED", "redeem code is being processed, please try again")
	ErrWeeklyQuotaDisabled = infraerrors.Forbidden("WEEKLY_QUOTA_DISABLED", "weekly quota is disabled")
	ErrWeeklyQuotaClaimed  = infraerrors.Conflict("WEEKLY_QUOTA_ALREADY_CLAIMED", "weekly quota already claimed for current window")
)

const (
	redeemMaxErrorsPerHour  = 20
	redeemRateLimitDuration = time.Hour
	redeemLockDuration      = 10 * time.Second // 锁超时时间，防止死锁
	weeklyQuotaWindow       = 7 * 24 * time.Hour
)

const (
	WeeklyQuotaStatusDisabled  = "disabled"
	WeeklyQuotaStatusClaimable = "claimable"
	WeeklyQuotaStatusClaimed   = "claimed"
)

type WeeklyQuotaInfo struct {
	Enabled          bool
	Amount           float64
	Status           string
	WindowStartedAt  time.Time
	WindowEndsAt     time.Time
	ClaimedAt        *time.Time
	NextClaimAt      *time.Time
	TotalClaimCount  int64
	TotalClaimAmount float64
}

type WeeklyQuotaClaimResult struct {
	Message         string
	Type            string
	Value           float64
	NewBalance      float64
	ClaimedAt       time.Time
	WindowStartedAt time.Time
	WindowEndsAt    time.Time
	NextClaimAt     time.Time
}

// RedeemCache defines cache operations for redeem service
type RedeemCache interface {
	GetRedeemAttemptCount(ctx context.Context, userID int64) (int, error)
	IncrementRedeemAttemptCount(ctx context.Context, userID int64) error

	AcquireRedeemLock(ctx context.Context, code string, ttl time.Duration) (bool, error)
	ReleaseRedeemLock(ctx context.Context, code string) error
}

type WeeklyQuotaSettingsProvider interface {
	GetAllSettings(ctx context.Context) (*SystemSettings, error)
}

type RedeemCodeRepository interface {
	Create(ctx context.Context, code *RedeemCode) error
	CreateBatch(ctx context.Context, codes []RedeemCode) error
	GetByID(ctx context.Context, id int64) (*RedeemCode, error)
	GetByCode(ctx context.Context, code string) (*RedeemCode, error)
	Update(ctx context.Context, code *RedeemCode) error
	Delete(ctx context.Context, id int64) error
	Use(ctx context.Context, id, userID int64) error

	List(ctx context.Context, params pagination.PaginationParams) ([]RedeemCode, *pagination.PaginationResult, error)
	ListWithFilters(ctx context.Context, params pagination.PaginationParams, codeType, status, search string) ([]RedeemCode, *pagination.PaginationResult, error)
	ListByUser(ctx context.Context, userID int64, limit int) ([]RedeemCode, error)
	// ListByUserPaginated returns paginated balance/concurrency history for a specific user.
	// codeType filter is optional - pass empty string to return all types.
	ListByUserPaginated(ctx context.Context, userID int64, params pagination.PaginationParams, codeType string) ([]RedeemCode, *pagination.PaginationResult, error)
	// SumPositiveBalanceByUser returns the total recharged amount (sum of positive balance values) for a user.
	SumPositiveBalanceByUser(ctx context.Context, userID int64) (float64, error)
}

type RedeemActivityRepository interface {
	ListUserActivity(ctx context.Context, userID int64, limit int) ([]RedeemHistoryItem, error)
}

// GenerateCodesRequest 生成兑换码请求
type GenerateCodesRequest struct {
	Count int     `json:"count"`
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

// RedeemCodeResponse 兑换码响应
type RedeemCodeResponse struct {
	Code      string    `json:"code"`
	Value     float64   `json:"value"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// RedeemService 兑换码服务
type RedeemService struct {
	redeemRepo           RedeemCodeRepository
	userRepo             UserRepository
	subscriptionService  *SubscriptionService
	cache                RedeemCache
	billingCacheService  *BillingCacheService
	entClient            *dbent.Client
	authCacheInvalidator APIKeyAuthCacheInvalidator
	settingsProvider     WeeklyQuotaSettingsProvider
}

// NewRedeemService 创建兑换码服务实例
func NewRedeemService(
	redeemRepo RedeemCodeRepository,
	userRepo UserRepository,
	subscriptionService *SubscriptionService,
	cache RedeemCache,
	billingCacheService *BillingCacheService,
	entClient *dbent.Client,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
	settingsProvider WeeklyQuotaSettingsProvider,
) *RedeemService {
	return &RedeemService{
		redeemRepo:           redeemRepo,
		userRepo:             userRepo,
		subscriptionService:  subscriptionService,
		cache:                cache,
		billingCacheService:  billingCacheService,
		entClient:            entClient,
		authCacheInvalidator: authCacheInvalidator,
		settingsProvider:     settingsProvider,
	}
}

// GenerateRandomCode 生成随机兑换码
func (s *RedeemService) GenerateRandomCode() (string, error) {
	// 生成16字节随机数据
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	// 转换为十六进制字符串
	code := hex.EncodeToString(bytes)

	// 格式化为 XXXX-XXXX-XXXX-XXXX 格式
	parts := []string{
		strings.ToUpper(code[0:8]),
		strings.ToUpper(code[8:16]),
		strings.ToUpper(code[16:24]),
		strings.ToUpper(code[24:32]),
	}

	return strings.Join(parts, "-"), nil
}

// GenerateCodes 批量生成兑换码
func (s *RedeemService) GenerateCodes(ctx context.Context, req GenerateCodesRequest) ([]RedeemCode, error) {
	if req.Count <= 0 {
		return nil, errors.New("count must be greater than 0")
	}

	// 邀请码类型不需要数值，其他类型需要非零值（支持负数用于退款）
	if req.Type != RedeemTypeInvitation && req.Value == 0 {
		return nil, errors.New("value must not be zero")
	}

	if req.Count > 1000 {
		return nil, errors.New("cannot generate more than 1000 codes at once")
	}

	codeType := req.Type
	if codeType == "" {
		codeType = RedeemTypeBalance
	}

	// 邀请码类型的 value 设为 0
	value := req.Value
	if codeType == RedeemTypeInvitation {
		value = 0
	}

	codes := make([]RedeemCode, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		code, err := s.GenerateRandomCode()
		if err != nil {
			return nil, fmt.Errorf("generate code: %w", err)
		}

		codes = append(codes, RedeemCode{
			Code:   code,
			Type:   codeType,
			Value:  value,
			Status: StatusUnused,
		})
	}

	// 批量插入
	if err := s.redeemRepo.CreateBatch(ctx, codes); err != nil {
		return nil, fmt.Errorf("create batch codes: %w", err)
	}

	return codes, nil
}

// CreateCode creates a redeem code with caller-provided code value.
// It is primarily used by admin integrations that require an external order ID
// to be mapped to a deterministic redeem code.
func (s *RedeemService) CreateCode(ctx context.Context, code *RedeemCode) error {
	if code == nil {
		return errors.New("redeem code is required")
	}
	code.Code = strings.TrimSpace(code.Code)
	if code.Code == "" {
		return errors.New("code is required")
	}
	if code.Type == "" {
		code.Type = RedeemTypeBalance
	}
	if code.Type != RedeemTypeInvitation && code.Value == 0 {
		return errors.New("value must not be zero")
	}
	if code.Status == "" {
		code.Status = StatusUnused
	}

	if err := s.redeemRepo.Create(ctx, code); err != nil {
		return fmt.Errorf("create redeem code: %w", err)
	}
	return nil
}

// checkRedeemRateLimit 检查用户兑换错误次数是否超限
func (s *RedeemService) checkRedeemRateLimit(ctx context.Context, userID int64) error {
	if s.cache == nil {
		return nil
	}

	count, err := s.cache.GetRedeemAttemptCount(ctx, userID)
	if err != nil {
		// Redis 出错时不阻止用户操作
		return nil
	}

	if count >= redeemMaxErrorsPerHour {
		return ErrRedeemRateLimited
	}

	return nil
}

// incrementRedeemErrorCount 增加用户兑换错误计数
func (s *RedeemService) incrementRedeemErrorCount(ctx context.Context, userID int64) {
	if s.cache == nil {
		return
	}

	_ = s.cache.IncrementRedeemAttemptCount(ctx, userID)
}

// acquireRedeemLock 尝试获取兑换码的分布式锁
// 返回 true 表示获取成功，false 表示锁已被占用
func (s *RedeemService) acquireRedeemLock(ctx context.Context, code string) bool {
	if s.cache == nil {
		return true // 无 Redis 时降级为不加锁
	}

	ok, err := s.cache.AcquireRedeemLock(ctx, code, redeemLockDuration)
	if err != nil {
		// Redis 出错时不阻止操作，依赖数据库层面的状态检查
		return true
	}
	return ok
}

// releaseRedeemLock 释放兑换码的分布式锁
func (s *RedeemService) releaseRedeemLock(ctx context.Context, code string) {
	if s.cache == nil {
		return
	}

	_ = s.cache.ReleaseRedeemLock(ctx, code)
}

// Redeem 使用兑换码
func (s *RedeemService) Redeem(ctx context.Context, userID int64, code string) (*RedeemCode, error) {
	// 检查限流
	if err := s.checkRedeemRateLimit(ctx, userID); err != nil {
		return nil, err
	}

	// 获取分布式锁，防止同一兑换码并发使用
	if !s.acquireRedeemLock(ctx, code) {
		return nil, ErrRedeemCodeLocked
	}
	defer s.releaseRedeemLock(ctx, code)

	// 查找兑换码
	redeemCode, err := s.redeemRepo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrRedeemCodeNotFound) {
			s.incrementRedeemErrorCount(ctx, userID)
			return nil, ErrRedeemCodeNotFound
		}
		return nil, fmt.Errorf("get redeem code: %w", err)
	}

	// 检查兑换码状态
	if !redeemCode.CanUse() {
		s.incrementRedeemErrorCount(ctx, userID)
		return nil, ErrRedeemCodeUsed
	}

	// 验证兑换码类型的前置条件
	if redeemCode.Type == RedeemTypeSubscription && redeemCode.GroupID == nil {
		return nil, infraerrors.BadRequest("REDEEM_CODE_INVALID", "invalid subscription redeem code: missing group_id")
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	// 使用数据库事务保证兑换码标记与权益发放的原子性
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 将事务放入 context，使 repository 方法能够使用同一事务
	txCtx := dbent.NewTxContext(ctx, tx)

	// 【关键】先标记兑换码为已使用，确保并发安全
	// 利用数据库乐观锁（WHERE status = 'unused'）保证原子性
	if err := s.redeemRepo.Use(txCtx, redeemCode.ID, userID); err != nil {
		if errors.Is(err, ErrRedeemCodeNotFound) || errors.Is(err, ErrRedeemCodeUsed) {
			return nil, ErrRedeemCodeUsed
		}
		return nil, fmt.Errorf("mark code as used: %w", err)
	}

	// 执行兑换逻辑（兑换码已被锁定，此时可安全操作）
	switch redeemCode.Type {
	case RedeemTypeBalance:
		amount := redeemCode.Value
		// 负数为退款扣减，余额最低为 0
		if amount < 0 && user.Balance+amount < 0 {
			amount = -user.Balance
		}
		if err := s.userRepo.UpdateBalance(txCtx, userID, amount); err != nil {
			return nil, fmt.Errorf("update user balance: %w", err)
		}

	case RedeemTypeConcurrency:
		delta := int(redeemCode.Value)
		// 负数为退款扣减，并发数最低为 0
		if delta < 0 && user.Concurrency+delta < 0 {
			delta = -user.Concurrency
		}
		if err := s.userRepo.UpdateConcurrency(txCtx, userID, delta); err != nil {
			return nil, fmt.Errorf("update user concurrency: %w", err)
		}

	case RedeemTypeSubscription:
		validityDays := redeemCode.ValidityDays
		if validityDays < 0 {
			// 负数天数：缩短订阅，减到 0 则取消订阅
			if err := s.reduceOrCancelSubscription(txCtx, userID, *redeemCode.GroupID, -validityDays, redeemCode.Code); err != nil {
				return nil, fmt.Errorf("reduce or cancel subscription: %w", err)
			}
		} else {
			if validityDays == 0 {
				validityDays = 30
			}
			_, _, err := s.subscriptionService.AssignOrExtendSubscription(txCtx, &AssignSubscriptionInput{
				UserID:       userID,
				GroupID:      *redeemCode.GroupID,
				ValidityDays: validityDays,
				AssignedBy:   0, // 系统分配
				Notes:        fmt.Sprintf("通过兑换码 %s 兑换", redeemCode.Code),
			})
			if err != nil {
				return nil, fmt.Errorf("assign or extend subscription: %w", err)
			}
		}

	default:
		return nil, fmt.Errorf("unsupported redeem type: %s", redeemCode.Type)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	// 事务提交成功后失效缓存
	s.invalidateRedeemCaches(ctx, userID, redeemCode)

	// 重新获取更新后的兑换码
	redeemCode, err = s.redeemRepo.GetByID(ctx, redeemCode.ID)
	if err != nil {
		return nil, fmt.Errorf("get updated redeem code: %w", err)
	}

	return redeemCode, nil
}

// invalidateRedeemCaches 失效兑换相关的缓存
func (s *RedeemService) invalidateRedeemCaches(ctx context.Context, userID int64, redeemCode *RedeemCode) {
	switch redeemCode.Type {
	case RedeemTypeBalance:
		if s.authCacheInvalidator != nil {
			s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
		}
		if s.billingCacheService == nil {
			return
		}
		go func() {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.billingCacheService.InvalidateUserBalance(cacheCtx, userID)
		}()
	case RedeemTypeConcurrency:
		if s.authCacheInvalidator != nil {
			s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
		}
		if s.billingCacheService == nil {
			return
		}
	case RedeemTypeSubscription:
		if s.authCacheInvalidator != nil {
			s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
		}
		if s.billingCacheService == nil {
			return
		}
		if redeemCode.GroupID != nil {
			groupID := *redeemCode.GroupID
			go func() {
				cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = s.billingCacheService.InvalidateSubscription(cacheCtx, userID, groupID)
			}()
		}
	}
}

// GetByID 根据ID获取兑换码
func (s *RedeemService) GetByID(ctx context.Context, id int64) (*RedeemCode, error) {
	code, err := s.redeemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get redeem code: %w", err)
	}
	return code, nil
}

// GetByCode 根据Code获取兑换码
func (s *RedeemService) GetByCode(ctx context.Context, code string) (*RedeemCode, error) {
	redeemCode, err := s.redeemRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("get redeem code: %w", err)
	}
	return redeemCode, nil
}

// List 获取兑换码列表（管理员功能）
func (s *RedeemService) List(ctx context.Context, params pagination.PaginationParams) ([]RedeemCode, *pagination.PaginationResult, error) {
	codes, pagination, err := s.redeemRepo.List(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("list redeem codes: %w", err)
	}
	return codes, pagination, nil
}

// Delete 删除兑换码（管理员功能）
func (s *RedeemService) Delete(ctx context.Context, id int64) error {
	// 检查兑换码是否存在
	code, err := s.redeemRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get redeem code: %w", err)
	}

	// 不允许删除已使用的兑换码
	if code.IsUsed() {
		return infraerrors.Conflict("REDEEM_CODE_DELETE_USED", "cannot delete used redeem code")
	}

	if err := s.redeemRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete redeem code: %w", err)
	}

	return nil
}

// GetStats 获取兑换码统计信息
func (s *RedeemService) GetStats(ctx context.Context) (map[string]any, error) {
	// TODO: 实现统计逻辑
	// 统计未使用、已使用的兑换码数量
	// 统计总面值等

	stats := map[string]any{
		"total_codes":  0,
		"unused_codes": 0,
		"used_codes":   0,
		"total_value":  0.0,
	}

	return stats, nil
}

// GetUserHistory 获取用户的兑换历史
func (s *RedeemService) GetUserHistory(ctx context.Context, userID int64, limit int) ([]RedeemCode, error) {
	codes, err := s.redeemRepo.ListByUser(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("get user redeem history: %w", err)
	}
	return codes, nil
}

func (s *RedeemService) GetUserActivityHistory(ctx context.Context, userID int64, limit int) ([]RedeemHistoryItem, error) {
	if activityRepo, ok := s.redeemRepo.(RedeemActivityRepository); ok {
		items, err := activityRepo.ListUserActivity(ctx, userID, limit)
		if err != nil {
			return nil, fmt.Errorf("get user activity history: %w", err)
		}
		return items, nil
	}
	codes, err := s.GetUserHistory(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	items := make([]RedeemHistoryItem, 0, len(codes))
	for i := range codes {
		code := codes[i]
		items = append(items, RedeemHistoryItem{
			ID:           code.ID,
			Code:         code.Code,
			Type:         code.Type,
			Value:        code.Value,
			Status:       code.Status,
			UsedBy:       code.UsedBy,
			UsedAt:       code.UsedAt,
			Notes:        code.Notes,
			CreatedAt:    code.CreatedAt,
			GroupID:      code.GroupID,
			ValidityDays: code.ValidityDays,
			User:         code.User,
			Group:        code.Group,
			Source:       "redeem_code",
		})
	}
	return items, nil
}

func (s *RedeemService) getWeeklyQuotaSettings(ctx context.Context) (*SystemSettings, error) {
	if s.settingsProvider == nil {
		return &SystemSettings{}, nil
	}
	settings, err := s.settingsProvider.GetAllSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get weekly quota settings: %w", err)
	}
	if settings == nil {
		return &SystemSettings{}, nil
	}
	return settings, nil
}

func weeklyQuotaWindowForUser(createdAt, now time.Time) (time.Time, time.Time) {
	createdAt = createdAt.UTC()
	now = now.UTC()
	if now.Before(createdAt) {
		return createdAt, createdAt.Add(weeklyQuotaWindow)
	}
	elapsed := now.Sub(createdAt)
	windowIndex := int64(elapsed / weeklyQuotaWindow)
	windowStart := createdAt.Add(time.Duration(windowIndex) * weeklyQuotaWindow)
	return windowStart, windowStart.Add(weeklyQuotaWindow)
}

func weeklyQuotaRedeemCode(userID int64, windowStart time.Time) string {
	return fmt.Sprintf("WQ-%d-%d", userID, windowStart.UTC().Unix())
}

func (s *RedeemService) getWeeklyQuotaClaimedAt(ctx context.Context, userID int64, windowStart, windowEnd time.Time) (*time.Time, error) {
	tx := dbent.TxFromContext(ctx)
	if tx == nil {
		return s.getWeeklyQuotaClaimedAtWithClient(ctx, s.entClient, userID, windowStart, windowEnd)
	}
	return s.getWeeklyQuotaClaimedAtWithTx(ctx, tx, userID, windowStart, windowEnd)
}

func (s *RedeemService) getWeeklyQuotaClaimedAtWithClient(ctx context.Context, client *dbent.Client, userID int64, windowStart, windowEnd time.Time) (*time.Time, error) {
	if client == nil {
		return nil, nil
	}
	record, err := client.RedeemCode.Query().
		Where(
			redeemcode.TypeEQ(RedeemTypeWeeklyBalance),
			redeemcode.UsedByEQ(userID),
			redeemcode.UsedAtGTE(windowStart),
			redeemcode.UsedAtLT(windowEnd),
		).
		Order(dbent.Desc(redeemcode.FieldUsedAt)).
		First(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("query weekly quota claim record: %w", err)
	}
	return record.UsedAt, nil
}

func (s *RedeemService) getWeeklyQuotaClaimedAtWithTx(ctx context.Context, tx *dbent.Tx, userID int64, windowStart, windowEnd time.Time) (*time.Time, error) {
	if tx == nil {
		return nil, nil
	}
	record, err := tx.RedeemCode.Query().
		Where(
			redeemcode.TypeEQ(RedeemTypeWeeklyBalance),
			redeemcode.UsedByEQ(userID),
			redeemcode.UsedAtGTE(windowStart),
			redeemcode.UsedAtLT(windowEnd),
		).
		Order(dbent.Desc(redeemcode.FieldUsedAt)).
		First(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("query weekly quota claim record: %w", err)
	}
	return record.UsedAt, nil
}

func (s *RedeemService) getWeeklyQuotaTotals(ctx context.Context, userID int64) (int64, float64, error) {
	if s.entClient == nil {
		return 0, 0, nil
	}
	records, err := s.entClient.RedeemCode.Query().
		Where(
			redeemcode.TypeEQ(RedeemTypeWeeklyBalance),
			redeemcode.UsedByEQ(userID),
		).
		All(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("query weekly quota totals: %w", err)
	}
	var total float64
	for _, record := range records {
		total += record.Value
	}
	return int64(len(records)), total, nil
}

func (s *RedeemService) GetWeeklyQuotaInfo(ctx context.Context, userID int64) (*WeeklyQuotaInfo, error) {
	settings, err := s.getWeeklyQuotaSettings(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	windowStart, windowEnd := weeklyQuotaWindowForUser(user.CreatedAt, time.Now().UTC())
	info := &WeeklyQuotaInfo{
		Enabled:          settings.WeeklyQuotaEnabled,
		Amount:           settings.WeeklyQuotaAmount,
		Status:           WeeklyQuotaStatusDisabled,
		WindowStartedAt:  windowStart,
		WindowEndsAt:     windowEnd,
		TotalClaimCount:  0,
		TotalClaimAmount: 0,
	}
	if !settings.WeeklyQuotaEnabled {
		return info, nil
	}

	totalCount, totalAmount, err := s.getWeeklyQuotaTotals(ctx, userID)
	if err != nil {
		return nil, err
	}
	info.TotalClaimCount = totalCount
	info.TotalClaimAmount = totalAmount
	info.Status = WeeklyQuotaStatusClaimable

	claimedAt, err := s.getWeeklyQuotaClaimedAt(ctx, userID, windowStart, windowEnd)
	if err != nil {
		return nil, err
	}
	if claimedAt != nil {
		info.ClaimedAt = claimedAt
		nextClaimAt := windowEnd
		info.NextClaimAt = &nextClaimAt
		info.Status = WeeklyQuotaStatusClaimed
	}

	return info, nil
}

func (s *RedeemService) ClaimWeeklyQuota(ctx context.Context, userID int64) (*WeeklyQuotaClaimResult, error) {
	settings, err := s.getWeeklyQuotaSettings(ctx)
	if err != nil {
		return nil, err
	}
	if !settings.WeeklyQuotaEnabled {
		return nil, ErrWeeklyQuotaDisabled
	}
	if settings.WeeklyQuotaAmount <= 0 {
		return nil, ErrWeeklyQuotaDisabled
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	now := time.Now().UTC()
	windowStart, windowEnd := weeklyQuotaWindowForUser(user.CreatedAt, now)
	claimedAt, err := s.getWeeklyQuotaClaimedAt(ctx, userID, windowStart, windowEnd)
	if err != nil {
		return nil, err
	}
	if claimedAt != nil {
		return nil, ErrWeeklyQuotaClaimed
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)

	claimedAt, err = s.getWeeklyQuotaClaimedAt(txCtx, userID, windowStart, windowEnd)
	if err != nil {
		return nil, err
	}
	if claimedAt != nil {
		return nil, ErrWeeklyQuotaClaimed
	}

	if err := s.userRepo.UpdateBalance(txCtx, userID, settings.WeeklyQuotaAmount); err != nil {
		return nil, fmt.Errorf("update user balance: %w", err)
	}

	redeemRecord := &RedeemCode{
		Code:         weeklyQuotaRedeemCode(userID, windowStart),
		Type:         RedeemTypeWeeklyBalance,
		Value:        settings.WeeklyQuotaAmount,
		Status:       StatusUsed,
		UsedBy:       &userID,
		UsedAt:       &now,
		CreatedAt:    now,
		ValidityDays: 7,
		Notes:        fmt.Sprintf("weekly quota claim for window starting %s", windowStart.Format(time.RFC3339)),
	}
	if err := s.redeemRepo.Create(txCtx, redeemRecord); err != nil {
		return nil, fmt.Errorf("create weekly quota redeem record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	s.invalidateRedeemCaches(ctx, userID, &RedeemCode{Type: RedeemTypeBalance})

	refreshedUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get updated user: %w", err)
	}

	return &WeeklyQuotaClaimResult{
		Message:         "Weekly quota claimed successfully",
		Type:            RedeemTypeWeeklyBalance,
		Value:           settings.WeeklyQuotaAmount,
		NewBalance:      refreshedUser.Balance,
		ClaimedAt:       now,
		WindowStartedAt: windowStart,
		WindowEndsAt:    windowEnd,
		NextClaimAt:     windowEnd,
	}, nil
}

// reduceOrCancelSubscription 缩短订阅天数，剩余天数 <= 0 时取消订阅
func (s *RedeemService) reduceOrCancelSubscription(ctx context.Context, userID, groupID int64, reduceDays int, code string) error {
	sub, err := s.subscriptionService.userSubRepo.GetByUserIDAndGroupID(ctx, userID, groupID)
	if err != nil {
		return ErrSubscriptionNotFound
	}

	now := time.Now()
	remaining := int(sub.ExpiresAt.Sub(now).Hours() / 24)
	if remaining < 0 {
		remaining = 0
	}

	notes := fmt.Sprintf("通过兑换码 %s 退款扣减 %d 天", code, reduceDays)

	if remaining <= reduceDays {
		// 剩余天数不足，直接取消订阅
		if err := s.subscriptionService.userSubRepo.UpdateStatus(ctx, sub.ID, SubscriptionStatusExpired); err != nil {
			return fmt.Errorf("cancel subscription: %w", err)
		}
		// 设置过期时间为当前时间
		if err := s.subscriptionService.userSubRepo.ExtendExpiry(ctx, sub.ID, now); err != nil {
			return fmt.Errorf("set subscription expiry: %w", err)
		}
	} else {
		// 缩短天数
		newExpiresAt := sub.ExpiresAt.AddDate(0, 0, -reduceDays)
		if err := s.subscriptionService.userSubRepo.ExtendExpiry(ctx, sub.ID, newExpiresAt); err != nil {
			return fmt.Errorf("reduce subscription: %w", err)
		}
	}

	// 追加备注
	newNotes := sub.Notes
	if newNotes != "" {
		newNotes += "\n"
	}
	newNotes += notes
	if err := s.subscriptionService.userSubRepo.UpdateNotes(ctx, sub.ID, newNotes); err != nil {
		return fmt.Errorf("update subscription notes: %w", err)
	}

	// 失效缓存
	s.subscriptionService.InvalidateSubCache(userID, groupID)

	return nil
}
