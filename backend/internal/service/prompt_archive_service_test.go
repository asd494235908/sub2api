package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

type promptArchiveRepoStub struct {
	mu       sync.Mutex
	settings *PromptArchiveSettings
	records  []*PromptArchivePersistedRecord
}

func (s *promptArchiveRepoStub) GetSettings(context.Context) (*PromptArchiveSettings, error) {
	if s.settings == nil {
		return DefaultPromptArchiveSettings(), nil
	}
	cp := *s.settings
	cp.GroupIDs = append([]int64(nil), cp.GroupIDs...)
	return &cp, nil
}

func (s *promptArchiveRepoStub) UpsertSettings(_ context.Context, settings *PromptArchiveSettings, operatorID int64) (*PromptArchiveSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *settings
	cp.UpdatedByUserID = operatorID
	cp.UpdatedAt = time.Now().UTC()
	s.settings = &cp
	return &cp, nil
}

func (s *promptArchiveRepoStub) InsertRecord(_ context.Context, record *PromptArchivePersistedRecord) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *record
	cp.Attachments = append([]PromptArchiveAttachment(nil), record.Attachments...)
	s.records = append(s.records, &cp)
	return int64(len(s.records)), nil
}

func (s *promptArchiveRepoStub) ListRecords(context.Context, *PromptArchiveRecordFilter) (*PromptArchiveRecordList, error) {
	return &PromptArchiveRecordList{}, nil
}

func (s *promptArchiveRepoStub) GetRecordByID(context.Context, int64) (*PromptArchiveRecordDetail, error) {
	return nil, nil
}

func (s *promptArchiveRepoStub) GetSessionByID(context.Context, string, int64) (*PromptArchiveSessionDetail, error) {
	return nil, nil
}

type promptArchiveStoreStub struct {
	uploadErr error
	keys      []string
	bodies    []string
}

func (s *promptArchiveStoreStub) Upload(_ context.Context, key string, body strings.Reader, contentType string) (int64, error) {
	_ = contentType
	return 0, errors.New("unimplemented")
}

func (s *promptArchiveStoreStub) UploadBytes(_ context.Context, key string, body []byte, _ string) (int64, error) {
	if s.uploadErr != nil {
		return 0, s.uploadErr
	}
	s.keys = append(s.keys, key)
	s.bodies = append(s.bodies, string(body))
	return int64(len(body)), nil
}

func (s *promptArchiveStoreStub) PresignURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://example.com/" + key, nil
}

func TestPromptArchiveService_EnqueueDropsWhenQueueFull(t *testing.T) {
	repo := &promptArchiveRepoStub{
		settings: &PromptArchiveSettings{Enabled: true, AllGroups: true},
	}
	cfg := &config.Config{
		Archive: config.ArchiveConfig{
			Enabled:             true,
			QueueCapacity:       1,
			WorkerCount:         1,
			BatchSize:           1,
			FlushIntervalSeconds: 2,
			OverflowPolicy:      config.PromptArchiveOverflowPolicyDropAndLog,
			InlineDataMaxBytes:  1024,
		},
	}
	svc := NewPromptArchiveService(repo, nil, cfg)

	env := &PromptArchiveEnvelope{
		RequestID:        "req-1",
		ClientRequestID:  "creq-1",
		SessionID:        "sess-1",
		UserID:           1,
		UsernameSnapshot: "u1",
		EmailSnapshot:    "u1@example.com",
		APIKeyID:         11,
		GroupID:          1,
		Protocol:         "anthropic",
		Endpoint:         "/v1/messages",
		Model:            "claude-sonnet",
		UserPromptText:   "hello",
		CreatedAt:        time.Now().UTC(),
	}

	if mode := svc.Enqueue(env); mode != PromptArchiveSubmitModeEnqueued {
		t.Fatalf("first enqueue mode=%s", mode)
	}
	if mode := svc.Enqueue(env); mode != PromptArchiveSubmitModeDropped {
		t.Fatalf("second enqueue mode=%s, want dropped", mode)
	}

	health := svc.Health()
	if health.DroppedCount != 1 {
		t.Fatalf("dropped=%d, want 1", health.DroppedCount)
	}
}

func TestPromptArchiveService_ProcessPersistsStoredRecord(t *testing.T) {
	repo := &promptArchiveRepoStub{
		settings: &PromptArchiveSettings{Enabled: true, AllGroups: true},
	}
	store := &promptArchiveStoreStub{}
	cfg := &config.Config{
		Archive: config.ArchiveConfig{
			Enabled:             true,
			QueueCapacity:       8,
			WorkerCount:         1,
			BatchSize:           1,
			FlushIntervalSeconds: 1,
			OverflowPolicy:      config.PromptArchiveOverflowPolicyDropAndLog,
			InlineDataMaxBytes:  1024 * 1024,
		},
	}
	svc := NewPromptArchiveService(repo, store, cfg)

	env := &PromptArchiveEnvelope{
		RequestID:        "req-2",
		ClientRequestID:  "creq-2",
		SessionID:        "sess-2",
		UserID:           2,
		UsernameSnapshot: "u2",
		EmailSnapshot:    "u2@example.com",
		APIKeyID:         12,
		GroupID:          2,
		Protocol:         "openai",
		Endpoint:         "/v1/responses",
		Model:            "gpt-5.4",
		SystemPrompt:     "system",
		UserPromptText:   "draw cat",
		Attachments: []PromptArchiveAttachment{
			{
				Kind:       PromptArchiveAttachmentKindImage,
				MIMEType:   "image/png",
				SourceType: PromptArchiveAttachmentSourceDataURI,
				SourceURL:  "data:image/png;base64,aGVsbG8=",
				Sequence:   1,
			},
		},
		CreatedAt: time.Now().UTC(),
	}

	if err := svc.processEnvelope(context.Background(), env); err != nil {
		t.Fatalf("processEnvelope error: %v", err)
	}

	if len(repo.records) != 1 {
		t.Fatalf("records=%d, want 1", len(repo.records))
	}
	if repo.records[0].Status != PromptArchiveRecordStatusStored {
		t.Fatalf("status=%s, want stored", repo.records[0].Status)
	}
	if repo.records[0].ObjectKey == "" {
		t.Fatalf("object key should not be empty")
	}
	if len(repo.records[0].Attachments) != 1 {
		t.Fatalf("attachments=%d, want 1", len(repo.records[0].Attachments))
	}
	if len(store.keys) < 2 {
		t.Fatalf("uploads=%d, want at least 2", len(store.keys))
	}
}

func TestPromptArchiveService_ProcessStoreFailureMarksRecordFailed(t *testing.T) {
	repo := &promptArchiveRepoStub{
		settings: &PromptArchiveSettings{Enabled: true, AllGroups: true},
	}
	store := &promptArchiveStoreStub{uploadErr: errors.New("upload failed")}
	cfg := &config.Config{
		Archive: config.ArchiveConfig{
			Enabled:             true,
			QueueCapacity:       8,
			WorkerCount:         1,
			BatchSize:           1,
			FlushIntervalSeconds: 1,
			OverflowPolicy:      config.PromptArchiveOverflowPolicyDropAndLog,
			InlineDataMaxBytes:  1024,
		},
	}
	svc := NewPromptArchiveService(repo, store, cfg)

	env := &PromptArchiveEnvelope{
		RequestID:        "req-3",
		ClientRequestID:  "creq-3",
		SessionID:        "sess-3",
		UserID:           3,
		UsernameSnapshot: "u3",
		EmailSnapshot:    "u3@example.com",
		APIKeyID:         13,
		GroupID:          3,
		Protocol:         "anthropic",
		Endpoint:         "/v1/messages",
		Model:            "claude-sonnet",
		UserPromptText:   "hello",
		CreatedAt:        time.Now().UTC(),
	}

	if err := svc.processEnvelope(context.Background(), env); err != nil {
		t.Fatalf("processEnvelope error: %v", err)
	}
	if len(repo.records) != 1 {
		t.Fatalf("records=%d, want 1", len(repo.records))
	}
	if repo.records[0].Status != PromptArchiveRecordStatusFailed {
		t.Fatalf("status=%s, want failed", repo.records[0].Status)
	}
	if repo.records[0].ErrorMessage == "" {
		t.Fatalf("failed record should include error message")
	}
}
