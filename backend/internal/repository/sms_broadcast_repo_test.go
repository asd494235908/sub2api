package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSMSBroadcastRepositoryCreateCampaignWithoutTemplateVarRowsColumn(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &smsBroadcastRepository{sql: db}

	campaign := &service.SMSBroadcastCampaign{
		Title:      "Maintenance",
		Mode:       service.SMSBroadcastModeTemplate,
		TemplateID: "broadcast-template",
		Status:     service.SMSBroadcastStatusDraft,
		Audience:   service.SMSBroadcastAudienceFilters{UserIDs: []int64{2}},
		TemplateVars: map[string]string{
			"code": "1213",
		},
		TemplateVarRows: []service.SMSBroadcastTemplateVarRow{
			{Key: "code", Value: "1213"},
		},
	}
	now := time.Date(2026, 5, 19, 23, 0, 0, 0, time.UTC)

	mock.ExpectQuery("information_schema\\.columns").
		WithArgs("sms_broadcast_campaigns", "template_var_rows").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("INSERT INTO sms_broadcast_campaigns").
		WithArgs(
			campaign.Title,
			string(campaign.Mode),
			campaign.TemplateID,
			campaign.Body,
			campaign.RenderedBody,
			string(campaign.Status),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			campaign.TotalRecipients,
			campaign.SentCount,
			campaign.FailedCount,
			campaign.SkippedCount,
			campaign.ErrorMessage,
			campaign.CreatedBy,
			campaign.UpdatedBy,
			campaign.StartedAt,
			campaign.FinishedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(9), now, now))

	err := repo.CreateCampaign(context.Background(), campaign)
	require.NoError(t, err)
	require.Equal(t, int64(9), campaign.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSMSBroadcastRepositoryScanCampaignRestoresRowsFromTemplateVarsCompat(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &smsBroadcastRepository{sql: db}

	rowsJSON, err := json.Marshal([]service.SMSBroadcastTemplateVarRow{{Key: "code", Value: "1213"}})
	require.NoError(t, err)
	varsJSON, err := json.Marshal(map[string]any{
		"code":                         "1213",
		smsBroadcastTemplateVarRowsKey: json.RawMessage(rowsJSON),
	})
	require.NoError(t, err)
	audienceJSON, err := json.Marshal(service.SMSBroadcastAudienceFilters{UserIDs: []int64{2}})
	require.NoError(t, err)
	now := time.Date(2026, 5, 19, 23, 0, 0, 0, time.UTC)

	mock.ExpectQuery("information_schema\\.columns").
		WithArgs("sms_broadcast_campaigns", "template_var_rows").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT id, title, mode, template_id").
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "mode", "template_id", "body", "rendered_body", "status", "audience",
			"template_vars", "total_recipients", "sent_count", "failed_count", "skipped_count",
			"error_message", "created_by", "updated_by", "started_at", "finished_at",
			"created_at", "updated_at", "template_var_rows",
		}).AddRow(
			int64(9), "Maintenance", string(service.SMSBroadcastModeTemplate), "broadcast-template", "", "",
			string(service.SMSBroadcastStatusQueued), audienceJSON, varsJSON, int64(1), int64(0), int64(0), int64(0),
			nil, nil, nil, nil, nil, now, now, []byte("[]"),
		))

	campaign, err := repo.GetCampaignByID(context.Background(), 9)
	require.NoError(t, err)
	require.Equal(t, []service.SMSBroadcastTemplateVarRow{{Key: "code", Value: "1213"}}, campaign.TemplateVarRows)
	require.Equal(t, map[string]string{"code": "1213"}, campaign.TemplateVars)
	require.NoError(t, mock.ExpectationsWereMet())
}
