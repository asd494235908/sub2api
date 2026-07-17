package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/tidwall/gjson"
)

const (
	defaultPromptArchiveQueueCapacity = 1024
	defaultPromptArchiveWorkerCount   = 8
	defaultPromptArchiveBatchSize     = 32
	defaultPromptArchiveFlushSeconds  = 2
	defaultPromptArchiveInlineMax     = 1024 * 1024
	promptArchivePresignTTL           = 24 * time.Hour
)

type PromptArchiveService struct {
	repo        PromptArchiveRepository
	store       PromptArchiveObjectStore
	cfg         config.ArchiveConfig
	queue       chan *PromptArchiveEnvelope
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	dropped     atomic.Uint64
	failed      atomic.Uint64
	stored      atomic.Uint64
	lastSuccess atomic.Int64
	lastFailure atomic.Int64
	lastErr     atomic.Value
	settings    atomic.Value
}

func DefaultPromptArchiveSettings() *PromptArchiveSettings {
	return &PromptArchiveSettings{
		Enabled:   false,
		AllGroups: false,
		GroupIDs:  []int64{},
	}
}

func NewPromptArchiveService(repo PromptArchiveRepository, store PromptArchiveObjectStore, cfg *config.Config) *PromptArchiveService {
	archiveCfg := config.ArchiveConfig{
		QueueCapacity:        defaultPromptArchiveQueueCapacity,
		WorkerCount:          defaultPromptArchiveWorkerCount,
		BatchSize:            defaultPromptArchiveBatchSize,
		FlushIntervalSeconds: defaultPromptArchiveFlushSeconds,
		OverflowPolicy:       config.PromptArchiveOverflowPolicyDropAndLog,
		InlineDataMaxBytes:   defaultPromptArchiveInlineMax,
	}
	if cfg != nil {
		archiveCfg = cfg.Archive
		if archiveCfg.QueueCapacity <= 0 {
			archiveCfg.QueueCapacity = defaultPromptArchiveQueueCapacity
		}
		if archiveCfg.WorkerCount <= 0 {
			archiveCfg.WorkerCount = defaultPromptArchiveWorkerCount
		}
		if archiveCfg.BatchSize <= 0 {
			archiveCfg.BatchSize = defaultPromptArchiveBatchSize
		}
		if archiveCfg.FlushIntervalSeconds <= 0 {
			archiveCfg.FlushIntervalSeconds = defaultPromptArchiveFlushSeconds
		}
		if archiveCfg.OverflowPolicy == "" {
			archiveCfg.OverflowPolicy = config.PromptArchiveOverflowPolicyDropAndLog
		}
		if archiveCfg.InlineDataMaxBytes <= 0 {
			archiveCfg.InlineDataMaxBytes = defaultPromptArchiveInlineMax
		}
	}
	svc := &PromptArchiveService{
		repo:  repo,
		store: store,
		cfg:   archiveCfg,
		queue: make(chan *PromptArchiveEnvelope, archiveCfg.QueueCapacity),
	}
	svc.ctx, svc.cancel = context.WithCancel(context.Background())
	svc.lastErr.Store("")
	svc.settings.Store(DefaultPromptArchiveSettings())
	return svc
}

func (s *PromptArchiveService) Start() {
	if s == nil {
		return
	}
	for i := 0; i < s.cfg.WorkerCount; i++ {
		s.wg.Add(1)
		go s.runWorker()
	}
}

func (s *PromptArchiveService) Stop() {
	if s == nil {
		return
	}
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

func (s *PromptArchiveService) runWorker() {
	defer s.wg.Done()
	for {
		select {
		case <-s.ctx.Done():
			return
		case env := <-s.queue:
			if env == nil {
				continue
			}
			_ = s.processEnvelope(s.ctx, env)
		}
	}
}

func (s *PromptArchiveService) Capture(ctx context.Context, env *PromptArchiveEnvelope) PromptArchiveSubmitMode {
	if s == nil || env == nil || !s.cfg.Enabled {
		return PromptArchiveSubmitModeDropped
	}
	enabled, err := s.isEnabledForGroup(ctx, env.GroupID)
	if err != nil || !enabled {
		if err != nil {
			logger.WriteSinkEvent("warn", "service.prompt_archive", "prompt_archive_skipped", map[string]any{
				"request_id": env.RequestID,
				"group_id":   env.GroupID,
				"reason":     "settings_lookup_failed",
				"error":      err.Error(),
			})
		}
		return PromptArchiveSubmitModeDropped
	}
	return s.Enqueue(env)
}

func (s *PromptArchiveService) isEnabledForGroup(ctx context.Context, groupID int64) (bool, error) {
	if s == nil || s.repo == nil {
		return false, nil
	}
	if cached, ok := s.settings.Load().(*PromptArchiveSettings); ok && cached != nil {
		if cached.Enabled {
			if cached.AllGroups {
				return true, nil
			}
			for _, id := range cached.GroupIDs {
				if id == groupID {
					return true, nil
				}
			}
		}
	}
	settings, err := s.repo.GetSettings(ctx)
	if err != nil {
		return false, err
	}
	if settings == nil {
		settings = DefaultPromptArchiveSettings()
	}
	s.settings.Store(settings)
	if !settings.Enabled {
		return false, nil
	}
	if settings.AllGroups {
		return true, nil
	}
	for _, id := range settings.GroupIDs {
		if id == groupID {
			return true, nil
		}
	}
	return false, nil
}

func (s *PromptArchiveService) GetSettings(ctx context.Context) (*PromptArchiveSettings, error) {
	if s == nil || s.repo == nil {
		return DefaultPromptArchiveSettings(), nil
	}
	settings, err := s.repo.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		settings = DefaultPromptArchiveSettings()
	}
	s.settings.Store(settings)
	return settings, nil
}

func (s *PromptArchiveService) UpdateSettings(ctx context.Context, settings *PromptArchiveSettings, operatorID int64) (*PromptArchiveSettings, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("prompt archive repository not configured")
	}
	updated, err := s.repo.UpsertSettings(ctx, settings, operatorID)
	if err != nil {
		return nil, err
	}
	s.settings.Store(updated)
	return updated, nil
}

func (s *PromptArchiveService) ListRecords(ctx context.Context, filter *PromptArchiveRecordFilter) (*PromptArchiveRecordList, error) {
	if s == nil || s.repo == nil {
		return &PromptArchiveRecordList{}, nil
	}
	return s.repo.ListRecords(ctx, filter)
}

func (s *PromptArchiveService) GetRecordByID(ctx context.Context, id int64) (*PromptArchiveRecordDetail, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	record, err := s.repo.GetRecordByID(ctx, id)
	if err != nil || record == nil {
		return record, err
	}
	if s.store != nil && strings.TrimSpace(record.ObjectKey) != "" {
		if url, signErr := s.store.PresignURL(ctx, record.ObjectKey, promptArchivePresignTTL); signErr == nil {
			record.PresignedURL = url
		}
	}
	return record, nil
}

func (s *PromptArchiveService) GetSessionByID(ctx context.Context, sessionID string, groupID int64) (*PromptArchiveSessionDetail, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	session, err := s.repo.GetSessionByID(ctx, sessionID, groupID)
	if err != nil || session == nil {
		return session, err
	}
	if s.store != nil {
		for _, record := range session.Records {
			if record != nil && strings.TrimSpace(record.ObjectKey) != "" {
				if url, signErr := s.store.PresignURL(ctx, record.ObjectKey, promptArchivePresignTTL); signErr == nil {
					record.PresignedURL = url
				}
			}
		}
	}
	return session, nil
}

func (s *PromptArchiveService) Enqueue(env *PromptArchiveEnvelope) PromptArchiveSubmitMode {
	if s == nil || env == nil {
		return PromptArchiveSubmitModeDropped
	}
	select {
	case s.queue <- env:
		logger.WriteSinkEvent("info", "service.prompt_archive", "prompt_archive_enqueued", map[string]any{
			"request_id":        env.RequestID,
			"client_request_id": env.ClientRequestID,
			"user_id":           env.UserID,
			"group_id":          env.GroupID,
			"model":             env.Model,
			"protocol":          env.Protocol,
		})
		return PromptArchiveSubmitModeEnqueued
	default:
		s.dropped.Add(1)
		logger.WriteSinkEvent("warn", "service.prompt_archive", "prompt_archive_dropped", map[string]any{
			"request_id":        env.RequestID,
			"client_request_id": env.ClientRequestID,
			"user_id":           env.UserID,
			"group_id":          env.GroupID,
			"model":             env.Model,
			"protocol":          env.Protocol,
			"queue_len":         len(s.queue),
		})
		return PromptArchiveSubmitModeDropped
	}
}

func (s *PromptArchiveService) Health() PromptArchiveHealth {
	health := PromptArchiveHealth{
		QueueDepth:    int64(len(s.queue)),
		QueueCapacity: int64(cap(s.queue)),
		DroppedCount:  s.dropped.Load(),
		FailedCount:   s.failed.Load(),
		StoredCount:   s.stored.Load(),
	}
	if v := s.lastSuccess.Load(); v > 0 {
		health.LastSuccessAt = time.Unix(0, v).UTC()
	}
	if v := s.lastFailure.Load(); v > 0 {
		health.LastFailureAt = time.Unix(0, v).UTC()
	}
	if v, ok := s.lastErr.Load().(string); ok {
		health.LastFailureError = v
	}
	return health
}

func (s *PromptArchiveService) processEnvelope(ctx context.Context, env *PromptArchiveEnvelope) error {
	if env == nil {
		return nil
	}
	if s.repo == nil {
		return fmt.Errorf("prompt archive repository not configured")
	}
	record := &PromptArchivePersistedRecord{
		RequestID:        env.RequestID,
		ClientRequestID:  env.ClientRequestID,
		SessionID:        env.SessionID,
		UserID:           env.UserID,
		UsernameSnapshot: env.UsernameSnapshot,
		EmailSnapshot:    env.EmailSnapshot,
		APIKeyID:         env.APIKeyID,
		GroupID:          env.GroupID,
		Protocol:         env.Protocol,
		Endpoint:         env.Endpoint,
		Model:            env.Model,
		SystemPrompt:     env.SystemPrompt,
		UserPromptText:   env.UserPromptText,
		PromptSummary:    buildPromptArchiveSummary(env.UserPromptText),
		Status:           PromptArchiveRecordStatusPending,
		CreatedAt:        env.CreatedAt,
	}
	if env.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}

	mdKey := buildPromptArchiveMarkdownObjectKey(record)
	markdown := buildPromptArchiveMarkdown(record)

	attachments := make([]PromptArchiveAttachment, 0, len(env.Attachments))
	for _, attachment := range env.Attachments {
		normalized, err := s.normalizeAttachment(ctx, record, attachment)
		if err != nil {
			record.Status = PromptArchiveRecordStatusFailed
			record.ErrorMessage = err.Error()
			record.Attachments = attachments
			_, _ = s.repo.InsertRecord(ctx, record)
			s.failed.Add(1)
			s.lastFailure.Store(time.Now().UnixNano())
			s.lastErr.Store(err.Error())
			logger.WriteSinkEvent("error", "service.prompt_archive", "prompt_archive_failed", map[string]any{
				"request_id": env.RequestID,
				"user_id":    env.UserID,
				"group_id":   env.GroupID,
				"error":      err.Error(),
			})
			return nil
		}
		attachments = append(attachments, normalized)
	}
	record.Attachments = attachments

	if s.store == nil {
		record.Status = PromptArchiveRecordStatusFailed
		record.ErrorMessage = "prompt archive object store not configured"
		record.ObjectKey = mdKey
		_, _ = s.repo.InsertRecord(ctx, record)
		s.failed.Add(1)
		s.lastFailure.Store(time.Now().UnixNano())
		s.lastErr.Store(record.ErrorMessage)
		return nil
	}

	if _, err := s.store.UploadBytes(ctx, mdKey, []byte(markdown), "text/markdown; charset=utf-8"); err != nil {
		record.Status = PromptArchiveRecordStatusFailed
		record.ErrorMessage = err.Error()
		record.ObjectKey = mdKey
		_, _ = s.repo.InsertRecord(ctx, record)
		s.failed.Add(1)
		s.lastFailure.Store(time.Now().UnixNano())
		s.lastErr.Store(err.Error())
		logger.WriteSinkEvent("error", "service.prompt_archive", "prompt_archive_failed", map[string]any{
			"request_id": env.RequestID,
			"user_id":    env.UserID,
			"group_id":   env.GroupID,
			"object_key": mdKey,
			"error":      err.Error(),
		})
		return nil
	}

	record.Status = PromptArchiveRecordStatusStored
	record.ObjectKey = mdKey
	_, _ = s.repo.InsertRecord(ctx, record)
	s.stored.Add(1)
	s.lastSuccess.Store(time.Now().UnixNano())
	logger.WriteSinkEvent("info", "service.prompt_archive", "prompt_archive_stored", map[string]any{
		"request_id": env.RequestID,
		"user_id":    env.UserID,
		"group_id":   env.GroupID,
		"object_key": mdKey,
	})
	return nil
}

func (s *PromptArchiveService) normalizeAttachment(ctx context.Context, record *PromptArchivePersistedRecord, attachment PromptArchiveAttachment) (PromptArchiveAttachment, error) {
	out := attachment
	mimeType, rawData, err := s.resolvePromptArchiveAttachmentData(ctx, out)
	if err != nil {
		return out, err
	}
	if len(rawData) == 0 {
		return out, nil
	}
	if out.MIMEType == "" {
		out.MIMEType = mimeType
	}
	out.SizeBytes = int64(len(rawData))
	sum := sha256.Sum256(rawData)
	out.SHA256 = hex.EncodeToString(sum[:])

	if promptArchiveShouldEnforceInlineSizeLimit(out.SourceType) && len(rawData) > s.cfg.InlineDataMaxBytes {
		out.SourceURL = strings.TrimSpace(out.SourceURL)
		return out, nil
	}

	objectKey := buildPromptArchiveAttachmentObjectKey(record, out)
	if _, err := s.store.UploadBytes(ctx, objectKey, rawData, firstNonEmptyPromptArchiveString(out.MIMEType, "application/octet-stream")); err != nil {
		return out, err
	}
	out.ObjectKey = objectKey
	return out, nil
}

func (s *PromptArchiveService) resolvePromptArchiveAttachmentData(ctx context.Context, attachment PromptArchiveAttachment) (string, []byte, error) {
	switch attachment.SourceType {
	case PromptArchiveAttachmentSourceDataURI, PromptArchiveAttachmentSourceInline:
		return decodePromptArchiveInlineData(strings.TrimSpace(attachment.SourceURL), attachment.SourceType)
	case PromptArchiveAttachmentSourceURL:
		return fetchPromptArchiveURLData(ctx, strings.TrimSpace(attachment.SourceURL), attachment.MIMEType)
	default:
		return "", nil, nil
	}
}

func promptArchiveShouldEnforceInlineSizeLimit(sourceType PromptArchiveAttachmentSourceType) bool {
	return sourceType == PromptArchiveAttachmentSourceDataURI || sourceType == PromptArchiveAttachmentSourceInline
}

func decodePromptArchiveInlineData(raw string, sourceType PromptArchiveAttachmentSourceType) (string, []byte, error) {
	switch sourceType {
	case PromptArchiveAttachmentSourceDataURI:
		prefix, data, found := strings.Cut(raw, ",")
		if !found {
			return "", nil, fmt.Errorf("invalid data uri")
		}
		mimeType := "application/octet-stream"
		if strings.HasPrefix(prefix, "data:") {
			meta := strings.TrimPrefix(prefix, "data:")
			if mediaType, _, ok := strings.Cut(meta, ";"); ok && strings.TrimSpace(mediaType) != "" {
				mimeType = strings.TrimSpace(mediaType)
			}
		}
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(data))
		if err != nil {
			return "", nil, fmt.Errorf("decode data uri: %w", err)
		}
		return mimeType, decoded, nil
	case PromptArchiveAttachmentSourceInline:
		payload := strings.TrimSpace(raw)
		if payload == "" {
			return "", nil, fmt.Errorf("empty inline data")
		}
		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return "", nil, fmt.Errorf("decode inline data: %w", err)
		}
		return "application/octet-stream", decoded, nil
	default:
		return "", nil, fmt.Errorf("unsupported source type")
	}
}

func fetchPromptArchiveURLData(ctx context.Context, rawURL string, fallbackMIME string) (string, []byte, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", nil, nil
	}
	if strings.HasPrefix(rawURL, "file://") {
		path := strings.TrimPrefix(rawURL, "file://")
		data, err := os.ReadFile(path)
		if err != nil {
			return "", nil, fmt.Errorf("read file url: %w", err)
		}
		mimeType := firstNonEmptyPromptArchiveString(fallbackMIME, mimeTypeFromExtension(path))
		return mimeType, data, nil
	}
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return "", nil, fmt.Errorf("build attachment request: %w", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", nil, fmt.Errorf("fetch attachment url: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return "", nil, fmt.Errorf("fetch attachment url status=%d", resp.StatusCode)
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", nil, fmt.Errorf("read attachment body: %w", err)
		}
		mimeType := firstNonEmptyPromptArchiveString(resp.Header.Get("Content-Type"), fallbackMIME, mimeTypeFromExtension(rawURL))
		return strings.TrimSpace(strings.Split(mimeType, ";")[0]), data, nil
	}
	return "", nil, nil
}

func mimeTypeFromExtension(raw string) string {
	switch strings.ToLower(filepath.Ext(raw)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	default:
		return ""
	}
}

func buildPromptArchiveSummary(text string) string {
	text = strings.TrimSpace(text)
	if len(text) <= 120 {
		return text
	}
	return strings.TrimSpace(text[:120]) + "..."
}

func buildPromptArchiveMarkdown(record *PromptArchivePersistedRecord) string {
	var b strings.Builder
	_, _ = b.WriteString("---\n")
	metadata := map[string]any{
		"request_id":        record.RequestID,
		"client_request_id": record.ClientRequestID,
		"session_id":        record.SessionID,
		"user_id":           record.UserID,
		"username":          record.UsernameSnapshot,
		"email":             record.EmailSnapshot,
		"api_key_id":        record.APIKeyID,
		"group_id":          record.GroupID,
		"protocol":          record.Protocol,
		"endpoint":          record.Endpoint,
		"model":             record.Model,
		"created_at":        record.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
	raw, _ := json.MarshalIndent(metadata, "", "  ")
	_, _ = b.Write(raw)
	_, _ = b.WriteString("\n---\n\n## System\n")
	_, _ = b.WriteString(strings.TrimSpace(record.SystemPrompt))
	_, _ = b.WriteString("\n\n## User Prompt\n")
	_, _ = b.WriteString(strings.TrimSpace(record.UserPromptText))
	_, _ = b.WriteString("\n\n## Attachments\n")
	if len(record.Attachments) == 0 {
		_, _ = b.WriteString("None\n")
		return b.String()
	}
	for _, attachment := range record.Attachments {
		_, _ = b.WriteString("- kind: ")
		_, _ = b.WriteString(string(attachment.Kind))
		if attachment.ObjectKey != "" {
			_, _ = b.WriteString(", object_key: ")
			_, _ = b.WriteString(attachment.ObjectKey)
		} else if attachment.SourceURL != "" {
			_, _ = b.WriteString(", source_url: ")
			_, _ = b.WriteString(attachment.SourceURL)
		}
		if attachment.MIMEType != "" {
			_, _ = b.WriteString(", mime_type: ")
			_, _ = b.WriteString(attachment.MIMEType)
		}
		_, _ = b.WriteString("\n")
	}
	return b.String()
}

func buildPromptArchiveMarkdownObjectKey(record *PromptArchivePersistedRecord) string {
	ts := record.CreatedAt.UTC().Format("20060102T150405.000000000Z")
	return fmt.Sprintf("%s/%s_%s.md", buildPromptArchiveSessionPrefix(record), ts, firstNonEmptyPromptArchiveString(record.RequestID, "unknown-request"))
}

func buildPromptArchiveAttachmentObjectKey(record *PromptArchivePersistedRecord, attachment PromptArchiveAttachment) string {
	ts := record.CreatedAt.UTC().Format("20060102T150405.000000000Z")
	ext := promptArchiveExtensionFromMIME(attachment.MIMEType)
	return fmt.Sprintf("%s/assets/%s_%03d_%s.%s", buildPromptArchiveSessionPrefix(record), ts, attachment.Sequence, firstNonEmptyPromptArchiveString(attachment.SHA256, "unknown"), ext)
}

func promptArchiveExtensionFromMIME(mimeType string) string {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	case "image/webp":
		return "webp"
	case "video/mp4":
		return "mp4"
	default:
		return "bin"
	}
}

func firstNonEmptyPromptArchiveString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func buildPromptArchiveSessionPrefix(record *PromptArchivePersistedRecord) string {
	datePath := record.CreatedAt.UTC().Format("2006/01/02")
	userSegment := buildPromptArchiveUserSegment(record)
	sessionSegment := "session_" + sanitizePromptArchivePathSegment(firstNonEmptyPromptArchiveString(record.SessionID, "unknown-session"))
	return fmt.Sprintf("prompt-archive/%s/%s/group_%d/%s", datePath, userSegment, record.GroupID, sessionSegment)
}

func buildPromptArchiveUserSegment(record *PromptArchivePersistedRecord) string {
	identity := firstNonEmptyPromptArchiveString(record.UsernameSnapshot, record.EmailSnapshot, "unknown-user")
	slug := sanitizePromptArchivePathSegment(identity)
	return fmt.Sprintf("user_%d_%s", record.UserID, slug)
}

func sanitizePromptArchivePathSegment(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return "unknown"
	}

	var b strings.Builder
	lastDash := false
	for _, r := range raw {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r):
			_, _ = b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				_ = b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "unknown"
	}
	return out
}

func BuildPromptArchiveEnvelopeFromParsedRequest(apiKey *APIKey, endpoint, protocol string, parsed *ParsedRequest, now time.Time) *PromptArchiveEnvelope {
	if apiKey == nil || apiKey.User == nil || parsed == nil {
		return nil
	}
	sessionID := strings.TrimSpace(parsed.MetadataUserID)
	if parsedUserID := ParseMetadataUserID(parsed.MetadataUserID); parsedUserID != nil {
		sessionID = parsedUserID.SessionID
	}
	groupID := int64(0)
	if apiKey.GroupID != nil {
		groupID = *apiKey.GroupID
	}
	var messages []any
	_ = parsed.DecodeMessages(&messages)
	system, _ := parsed.SystemValue()
	return &PromptArchiveEnvelope{
		RequestID:        "",
		ClientRequestID:  "",
		SessionID:        sessionID,
		UserID:           apiKey.User.ID,
		UsernameSnapshot: apiKey.User.Username,
		EmailSnapshot:    apiKey.User.Email,
		APIKeyID:         apiKey.ID,
		GroupID:          groupID,
		Protocol:         protocol,
		Endpoint:         endpoint,
		Model:            parsed.Model,
		SystemPrompt:     extractPromptArchiveSystemText(system),
		UserPromptText:   extractPromptArchiveUserTextFromMessages(messages),
		CreatedAt:        normalizePromptArchiveTime(now),
	}
}

func EnsurePromptArchiveSessionID(env *PromptArchiveEnvelope, fallback string) *PromptArchiveEnvelope {
	if env == nil {
		return nil
	}
	if strings.TrimSpace(env.SessionID) == "" {
		env.SessionID = strings.TrimSpace(fallback)
	}
	return env
}

func BuildPromptArchiveEnvelopeFromOpenAIResponsesBody(apiKey *APIKey, endpoint string, body []byte, now time.Time) *PromptArchiveEnvelope {
	if apiKey == nil || apiKey.User == nil || len(body) == 0 {
		return nil
	}
	groupID := int64(0)
	if apiKey.GroupID != nil {
		groupID = *apiKey.GroupID
	}
	env := &PromptArchiveEnvelope{
		SessionID:        strings.TrimSpace(gjson.GetBytes(body, "prompt_cache_key").String()),
		UserID:           apiKey.User.ID,
		UsernameSnapshot: apiKey.User.Username,
		EmailSnapshot:    apiKey.User.Email,
		APIKeyID:         apiKey.ID,
		GroupID:          groupID,
		Protocol:         "openai",
		Endpoint:         endpoint,
		Model:            strings.TrimSpace(gjson.GetBytes(body, "model").String()),
		SystemPrompt:     strings.TrimSpace(gjson.GetBytes(body, "system").String()),
		UserPromptText:   extractPromptArchiveOpenAIInputText(body),
		Attachments:      extractPromptArchiveOpenAIAttachments(body),
		CreatedAt:        normalizePromptArchiveTime(now),
	}
	return env
}

func BuildPromptArchiveEnvelopeFromGeminiBody(apiKey *APIKey, endpoint, model, sessionID string, body []byte, now time.Time) *PromptArchiveEnvelope {
	if apiKey == nil || apiKey.User == nil || len(body) == 0 {
		return nil
	}
	groupID := int64(0)
	if apiKey.GroupID != nil {
		groupID = *apiKey.GroupID
	}
	env := &PromptArchiveEnvelope{
		SessionID:        strings.TrimSpace(sessionID),
		UserID:           apiKey.User.ID,
		UsernameSnapshot: apiKey.User.Username,
		EmailSnapshot:    apiKey.User.Email,
		APIKeyID:         apiKey.ID,
		GroupID:          groupID,
		Protocol:         "gemini",
		Endpoint:         endpoint,
		Model:            strings.TrimSpace(model),
		UserPromptText:   extractPromptArchiveGeminiText(body),
		Attachments:      extractPromptArchiveGeminiAttachments(body),
		CreatedAt:        normalizePromptArchiveTime(now),
	}
	return env
}

func normalizePromptArchiveTime(now time.Time) time.Time {
	if now.IsZero() {
		return time.Now().UTC()
	}
	return now.UTC()
}

func extractPromptArchiveSystemText(system any) string {
	switch v := system.(type) {
	case string:
		return strings.TrimSpace(v)
	case []any:
		var texts []string
		for _, part := range v {
			partMap, ok := part.(map[string]any)
			if !ok {
				continue
			}
			if text, ok := partMap["text"].(string); ok && strings.TrimSpace(text) != "" {
				texts = append(texts, strings.TrimSpace(text))
			}
		}
		return strings.Join(texts, "\n")
	default:
		return ""
	}
}

func extractPromptArchiveUserTextFromMessages(messages []any) string {
	for _, msg := range messages {
		msgMap, ok := msg.(map[string]any)
		if !ok {
			continue
		}
		if role, _ := msgMap["role"].(string); strings.TrimSpace(role) != "user" {
			continue
		}
		if content, ok := msgMap["content"]; ok {
			return strings.TrimSpace(extractPromptArchiveTextFromContent(content))
		}
	}
	return ""
}

func extractPromptArchiveTextFromContent(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var texts []string
		for _, part := range v {
			partMap, ok := part.(map[string]any)
			if !ok {
				continue
			}
			if partType, _ := partMap["type"].(string); partType == "text" || partType == "input_text" {
				if text, ok := partMap["text"].(string); ok && strings.TrimSpace(text) != "" {
					texts = append(texts, strings.TrimSpace(text))
				}
			}
		}
		return strings.Join(texts, "\n")
	default:
		return ""
	}
}

func extractPromptArchiveOpenAIInputText(body []byte) string {
	input := gjson.GetBytes(body, "input")
	if !input.Exists() || !input.IsArray() {
		return ""
	}
	var texts []string
	for _, item := range input.Array() {
		content := item.Get("content")
		if content.IsArray() {
			for _, part := range content.Array() {
				partType := strings.TrimSpace(part.Get("type").String())
				if (partType == "input_text" || partType == "text") && strings.TrimSpace(part.Get("text").String()) != "" {
					texts = append(texts, strings.TrimSpace(part.Get("text").String()))
				}
			}
		}
	}
	return strings.Join(texts, "\n")
}

func extractPromptArchiveOpenAIAttachments(body []byte) []PromptArchiveAttachment {
	input := gjson.GetBytes(body, "input")
	if !input.Exists() || !input.IsArray() {
		return nil
	}
	var attachments []PromptArchiveAttachment
	sequence := 1
	for _, item := range input.Array() {
		content := item.Get("content")
		if !content.IsArray() {
			continue
		}
		for _, part := range content.Array() {
			partType := strings.TrimSpace(part.Get("type").String())
			switch partType {
			case "input_image":
				imageURL := strings.TrimSpace(part.Get("image_url").String())
				if imageURL == "" {
					continue
				}
				sourceType := PromptArchiveAttachmentSourceURL
				if strings.HasPrefix(imageURL, "data:") {
					sourceType = PromptArchiveAttachmentSourceDataURI
				}
				attachments = append(attachments, PromptArchiveAttachment{
					Kind:       PromptArchiveAttachmentKindImage,
					MIMEType:   inferPromptArchiveMIMEFromURL(imageURL),
					SourceType: sourceType,
					SourceURL:  imageURL,
					Sequence:   sequence,
				})
				sequence++
			case "input_video":
				videoURL := strings.TrimSpace(part.Get("video_url").String())
				if videoURL == "" {
					continue
				}
				attachments = append(attachments, PromptArchiveAttachment{
					Kind:       PromptArchiveAttachmentKindVideo,
					MIMEType:   inferPromptArchiveMIMEFromURL(videoURL),
					SourceType: PromptArchiveAttachmentSourceURL,
					SourceURL:  videoURL,
					Sequence:   sequence,
				})
				sequence++
			case "image_url":
				imageURL := strings.TrimSpace(part.Get("image_url.url").String())
				if imageURL == "" {
					imageURL = strings.TrimSpace(part.Get("image_url").String())
				}
				if imageURL == "" {
					continue
				}
				attachments = append(attachments, PromptArchiveAttachment{
					Kind:       PromptArchiveAttachmentKindImage,
					MIMEType:   inferPromptArchiveMIMEFromURL(imageURL),
					SourceType: PromptArchiveAttachmentSourceURL,
					SourceURL:  imageURL,
					Sequence:   sequence,
				})
				sequence++
			}
		}
	}
	return attachments
}

func extractPromptArchiveGeminiText(body []byte) string {
	contents := gjson.GetBytes(body, "contents")
	if !contents.Exists() || !contents.IsArray() {
		return ""
	}
	var texts []string
	for _, item := range contents.Array() {
		if strings.TrimSpace(item.Get("role").String()) != "user" {
			continue
		}
		for _, part := range item.Get("parts").Array() {
			if text := strings.TrimSpace(part.Get("text").String()); text != "" {
				texts = append(texts, text)
			}
		}
	}
	return strings.Join(texts, "\n")
}

func extractPromptArchiveGeminiAttachments(body []byte) []PromptArchiveAttachment {
	contents := gjson.GetBytes(body, "contents")
	if !contents.Exists() || !contents.IsArray() {
		return nil
	}
	var attachments []PromptArchiveAttachment
	sequence := 1
	for _, item := range contents.Array() {
		if strings.TrimSpace(item.Get("role").String()) != "user" {
			continue
		}
		for _, part := range item.Get("parts").Array() {
			inlineData := part.Get("inline_data")
			if inlineData.Exists() {
				data := strings.TrimSpace(inlineData.Get("data").String())
				if data == "" {
					continue
				}
				attachments = append(attachments, PromptArchiveAttachment{
					Kind:       PromptArchiveAttachmentKindImage,
					MIMEType:   strings.TrimSpace(inlineData.Get("mime_type").String()),
					SourceType: PromptArchiveAttachmentSourceInline,
					SourceURL:  data,
					Sequence:   sequence,
				})
				sequence++
			}
			fileData := part.Get("file_data")
			if fileData.Exists() {
				fileURL := strings.TrimSpace(fileData.Get("file_uri").String())
				if fileURL == "" {
					continue
				}
				kind := PromptArchiveAttachmentKindImage
				mimeType := strings.TrimSpace(fileData.Get("mime_type").String())
				if strings.HasPrefix(strings.ToLower(mimeType), "video/") || strings.HasSuffix(strings.ToLower(fileURL), ".mp4") {
					kind = PromptArchiveAttachmentKindVideo
				}
				attachments = append(attachments, PromptArchiveAttachment{
					Kind:       kind,
					MIMEType:   mimeType,
					SourceType: PromptArchiveAttachmentSourceURL,
					SourceURL:  fileURL,
					Sequence:   sequence,
				})
				sequence++
			}
		}
	}
	return attachments
}

func inferPromptArchiveMIMEFromURL(raw string) string {
	if strings.HasPrefix(raw, "data:") {
		prefix, _, found := strings.Cut(raw, ",")
		if !found {
			return ""
		}
		if strings.HasPrefix(prefix, "data:") {
			meta := strings.TrimPrefix(prefix, "data:")
			if mediaType, _, ok := strings.Cut(meta, ";"); ok {
				return strings.TrimSpace(mediaType)
			}
			return strings.TrimSpace(meta)
		}
	}
	switch {
	case strings.Contains(strings.ToLower(raw), ".png"):
		return "image/png"
	case strings.Contains(strings.ToLower(raw), ".jpg"), strings.Contains(strings.ToLower(raw), ".jpeg"):
		return "image/jpeg"
	case strings.Contains(strings.ToLower(raw), ".webp"):
		return "image/webp"
	case strings.Contains(strings.ToLower(raw), ".mp4"):
		return "video/mp4"
	default:
		return ""
	}
}
