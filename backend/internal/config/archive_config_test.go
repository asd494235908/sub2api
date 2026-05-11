package config

import "testing"

func TestLoadDefaultArchiveConfig(t *testing.T) {
	resetViperWithJWTSecret(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Archive.Enabled {
		t.Fatalf("Archive.Enabled = true, want false")
	}
	if cfg.Archive.QueueCapacity != 1024 {
		t.Fatalf("Archive.QueueCapacity = %d, want 1024", cfg.Archive.QueueCapacity)
	}
	if cfg.Archive.WorkerCount != 8 {
		t.Fatalf("Archive.WorkerCount = %d, want 8", cfg.Archive.WorkerCount)
	}
	if cfg.Archive.BatchSize != 32 {
		t.Fatalf("Archive.BatchSize = %d, want 32", cfg.Archive.BatchSize)
	}
	if cfg.Archive.FlushIntervalSeconds != 2 {
		t.Fatalf("Archive.FlushIntervalSeconds = %d, want 2", cfg.Archive.FlushIntervalSeconds)
	}
	if cfg.Archive.OverflowPolicy != PromptArchiveOverflowPolicyDropAndLog {
		t.Fatalf("Archive.OverflowPolicy = %q, want %q", cfg.Archive.OverflowPolicy, PromptArchiveOverflowPolicyDropAndLog)
	}
	if cfg.Archive.InlineDataMaxBytes != 1024*1024 {
		t.Fatalf("Archive.InlineDataMaxBytes = %d, want %d", cfg.Archive.InlineDataMaxBytes, 1024*1024)
	}
	if cfg.Archive.MinIO.Region != "auto" {
		t.Fatalf("Archive.MinIO.Region = %q, want auto", cfg.Archive.MinIO.Region)
	}
	if !cfg.Archive.MinIO.ForcePathStyle {
		t.Fatalf("Archive.MinIO.ForcePathStyle = false, want true")
	}
}
