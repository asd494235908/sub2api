package repository

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type promptArchiveObjectStore struct {
	store service.BackupObjectStore
}

func ProvidePromptArchiveObjectStore(cfg *config.Config, factory service.BackupObjectStoreFactory) (service.PromptArchiveObjectStore, error) {
	if cfg == nil || !cfg.Archive.Enabled {
		return nil, nil
	}
	return NewPromptArchiveObjectStoreFactory(factory, &service.PromptArchiveObjectStoreConfig{
		Endpoint:        cfg.Archive.MinIO.Endpoint,
		Bucket:          cfg.Archive.MinIO.Bucket,
		Region:          cfg.Archive.MinIO.Region,
		AccessKeyID:     cfg.Archive.MinIO.AccessKey,
		SecretAccessKey: cfg.Archive.MinIO.SecretKey,
		ForcePathStyle:  cfg.Archive.MinIO.ForcePathStyle,
	})
}

func NewPromptArchiveObjectStoreFactory(factory service.BackupObjectStoreFactory, cfg *service.PromptArchiveObjectStoreConfig) (service.PromptArchiveObjectStore, error) {
	if factory == nil {
		return nil, fmt.Errorf("nil backup object store factory")
	}
	if cfg == nil {
		return nil, fmt.Errorf("nil prompt archive object store config")
	}
	store, err := factory(context.Background(), &service.BackupS3Config{
		Endpoint:        cfg.Endpoint,
		Bucket:          cfg.Bucket,
		Region:          cfg.Region,
		AccessKeyID:     cfg.AccessKeyID,
		SecretAccessKey: cfg.SecretAccessKey,
		ForcePathStyle:  cfg.ForcePathStyle,
	})
	if err != nil {
		return nil, err
	}
	return &promptArchiveObjectStore{store: store}, nil
}

func (s *promptArchiveObjectStore) UploadBytes(ctx context.Context, key string, body []byte, contentType string) (int64, error) {
	if s == nil || s.store == nil {
		return 0, fmt.Errorf("prompt archive object store not configured")
	}
	return s.store.Upload(ctx, key, bytes.NewReader(body), contentType)
}

func (s *promptArchiveObjectStore) PresignURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if s == nil || s.store == nil {
		return "", fmt.Errorf("prompt archive object store not configured")
	}
	return s.store.PresignURL(ctx, key, expiry)
}
