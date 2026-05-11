package service

import (
	"context"
	"time"
)

type PromptArchiveAttachmentKind string
type PromptArchiveAttachmentSourceType string
type PromptArchiveRecordStatus string
type PromptArchiveSubmitMode string

const (
	PromptArchiveAttachmentKindImage PromptArchiveAttachmentKind = "image"
	PromptArchiveAttachmentKindVideo PromptArchiveAttachmentKind = "video"

	PromptArchiveAttachmentSourceURL      PromptArchiveAttachmentSourceType = "url"
	PromptArchiveAttachmentSourceDataURI  PromptArchiveAttachmentSourceType = "data_uri"
	PromptArchiveAttachmentSourceInline   PromptArchiveAttachmentSourceType = "inline_data"

	PromptArchiveRecordStatusPending PromptArchiveRecordStatus = "pending"
	PromptArchiveRecordStatusStored  PromptArchiveRecordStatus = "stored"
	PromptArchiveRecordStatusFailed  PromptArchiveRecordStatus = "failed"
	PromptArchiveRecordStatusDropped PromptArchiveRecordStatus = "dropped"

	PromptArchiveSubmitModeEnqueued PromptArchiveSubmitMode = "enqueued"
	PromptArchiveSubmitModeDropped  PromptArchiveSubmitMode = "dropped"
)

type PromptArchiveAttachment struct {
	Kind       PromptArchiveAttachmentKind
	MIMEType   string
	SourceType PromptArchiveAttachmentSourceType
	SourceURL  string
	ObjectKey  string
	SHA256     string
	SizeBytes  int64
	Sequence   int
}

type PromptArchiveEnvelope struct {
	RequestID        string
	ClientRequestID  string
	SessionID        string
	UserID           int64
	UsernameSnapshot string
	EmailSnapshot    string
	APIKeyID         int64
	GroupID          int64
	Protocol         string
	Endpoint         string
	Model            string
	SystemPrompt     string
	UserPromptText   string
	PromptSummary    string
	Attachments      []PromptArchiveAttachment
	CreatedAt        time.Time
}

type PromptArchiveSettings struct {
	Enabled         bool      `json:"enabled"`
	AllGroups       bool      `json:"all_groups"`
	GroupIDs        []int64   `json:"group_ids"`
	Bucket          string    `json:"bucket"`
	UpdatedAt       time.Time `json:"updated_at"`
	UpdatedByUserID int64     `json:"updated_by_user_id"`
}

type PromptArchivePersistedRecord struct {
	RequestID        string
	ClientRequestID  string
	SessionID        string
	UserID           int64
	UsernameSnapshot string
	EmailSnapshot    string
	APIKeyID         int64
	GroupID          int64
	Protocol         string
	Endpoint         string
	Model            string
	SystemPrompt     string
	UserPromptText   string
	PromptSummary    string
	ObjectKey        string
	Status           PromptArchiveRecordStatus
	ErrorMessage     string
	Attachments      []PromptArchiveAttachment
	CreatedAt        time.Time
}

type PromptArchiveRecordFilter struct {
	Page            int
	PageSize        int
	Query           string
	Username        string
	Email           string
	SessionID       string
	Model           string
	Protocol        string
	GroupID         *int64
	UserID          *int64
	APIKeyID        *int64
	Status          string
	StartTime       *time.Time
	EndTime         *time.Time
}

type PromptArchiveRecordList struct {
	Items    []*PromptArchiveRecordDetail
	Total    int
	Page     int
	PageSize int
}

type PromptArchiveRecordDetail struct {
	ID int64
	PromptArchivePersistedRecord
	PresignedURL string
}

type PromptArchiveSessionDetail struct {
	SessionID string
	GroupID   int64
	Records   []*PromptArchiveRecordDetail
}

type PromptArchiveHealth struct {
	QueueDepth       int64     `json:"queue_depth"`
	QueueCapacity    int64     `json:"queue_capacity"`
	DroppedCount     uint64    `json:"dropped_count"`
	FailedCount      uint64    `json:"failed_count"`
	StoredCount      uint64    `json:"stored_count"`
	LastSuccessAt    time.Time `json:"last_success_at"`
	LastFailureAt    time.Time `json:"last_failure_at"`
	LastFailureError string    `json:"last_failure_error"`
}

type PromptArchiveRepository interface {
	GetSettings(ctx context.Context) (*PromptArchiveSettings, error)
	UpsertSettings(ctx context.Context, settings *PromptArchiveSettings, operatorID int64) (*PromptArchiveSettings, error)
	InsertRecord(ctx context.Context, record *PromptArchivePersistedRecord) (int64, error)
	ListRecords(ctx context.Context, filter *PromptArchiveRecordFilter) (*PromptArchiveRecordList, error)
	GetRecordByID(ctx context.Context, id int64) (*PromptArchiveRecordDetail, error)
	GetSessionByID(ctx context.Context, sessionID string, groupID int64) (*PromptArchiveSessionDetail, error)
}

type PromptArchiveObjectStore interface {
	UploadBytes(ctx context.Context, key string, body []byte, contentType string) (int64, error)
	PresignURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

type PromptArchiveObjectStoreConfig struct {
	Endpoint       string
	Bucket         string
	Region         string
	AccessKeyID    string
	SecretAccessKey string
	ForcePathStyle bool
}
