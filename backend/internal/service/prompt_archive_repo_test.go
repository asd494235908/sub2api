package service

import (
	"context"
	"testing"
)

func TestPromptArchiveService_UpdateSettingsCachesAndReturnsUpdated(t *testing.T) {
	repo := &promptArchiveRepoStub{}
	svc := NewPromptArchiveService(repo, nil, nil)

	updated, err := svc.UpdateSettings(context.Background(), &PromptArchiveSettings{
		Enabled:   true,
		AllGroups: false,
		GroupIDs:  []int64{1, 2},
		Bucket:    "archive-bucket",
	}, 99)
	if err != nil {
		t.Fatalf("UpdateSettings error: %v", err)
	}
	if !updated.Enabled {
		t.Fatalf("updated.Enabled should be true")
	}
	got, err := svc.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("GetSettings error: %v", err)
	}
	if len(got.GroupIDs) != 2 {
		t.Fatalf("group_ids=%v", got.GroupIDs)
	}
}

func TestPromptArchiveService_CaptureSkipsWhenGroupNotEnabled(t *testing.T) {
	repo := &promptArchiveRepoStub{
		settings: &PromptArchiveSettings{
			Enabled:   true,
			AllGroups: false,
			GroupIDs:  []int64{7},
		},
	}
	svc := NewPromptArchiveService(repo, nil, nil)
	env := &PromptArchiveEnvelope{
		RequestID:       "req-skip",
		ClientRequestID: "creq-skip",
		GroupID:         9,
	}
	if mode := svc.Capture(context.Background(), env); mode != PromptArchiveSubmitModeDropped {
		t.Fatalf("capture mode=%s, want dropped", mode)
	}
}
