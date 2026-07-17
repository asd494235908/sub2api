package service

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
)

func ProvidePromptArchiveService(repo PromptArchiveRepository, store PromptArchiveObjectStore, cfg *config.Config) *PromptArchiveService {
	svc := NewPromptArchiveService(repo, store, cfg)
	svc.Start()
	return svc
}
