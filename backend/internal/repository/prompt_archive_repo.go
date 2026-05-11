package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type promptArchiveRepository struct {
	db *sql.DB
}

func NewPromptArchiveRepository(db *sql.DB) service.PromptArchiveRepository {
	return &promptArchiveRepository{db: db}
}

func (r *promptArchiveRepository) GetSettings(ctx context.Context) (*service.PromptArchiveSettings, error) {
	if r == nil || r.db == nil {
		return service.DefaultPromptArchiveSettings(), nil
	}
	const query = `
SELECT enabled, all_groups, group_ids, bucket, updated_at, updated_by_user_id
FROM prompt_archive.archive_settings
ORDER BY id DESC
LIMIT 1`
	row := r.db.QueryRowContext(ctx, query)
	var (
		enabled         bool
		allGroups       bool
		groupIDsJSON    string
		bucket          sql.NullString
		updatedAt       sql.NullTime
		updatedByUserID sql.NullInt64
	)
	err := row.Scan(&enabled, &allGroups, &groupIDsJSON, &bucket, &updatedAt, &updatedByUserID)
	if err == sql.ErrNoRows {
		return service.DefaultPromptArchiveSettings(), nil
	}
	if err != nil {
		return nil, err
	}
	out := &service.PromptArchiveSettings{
		Enabled:   enabled,
		AllGroups: allGroups,
		Bucket:    strings.TrimSpace(bucket.String),
	}
	if updatedAt.Valid {
		out.UpdatedAt = updatedAt.Time.UTC()
	}
	if updatedByUserID.Valid {
		out.UpdatedByUserID = updatedByUserID.Int64
	}
	if strings.TrimSpace(groupIDsJSON) != "" {
		_ = json.Unmarshal([]byte(groupIDsJSON), &out.GroupIDs)
	}
	if out.GroupIDs == nil {
		out.GroupIDs = []int64{}
	}
	return out, nil
}

func (r *promptArchiveRepository) UpsertSettings(ctx context.Context, settings *service.PromptArchiveSettings, operatorID int64) (*service.PromptArchiveSettings, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil prompt archive repository")
	}
	if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}
	groupIDsJSON, err := json.Marshal(settings.GroupIDs)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	const query = `
INSERT INTO prompt_archive.archive_settings (
  singleton_key, enabled, all_groups, group_ids, bucket, updated_at, updated_by_user_id
) VALUES (
  1, $1, $2, $3, $4, $5, $6
)
ON CONFLICT (singleton_key) DO UPDATE SET
  enabled = EXCLUDED.enabled,
  all_groups = EXCLUDED.all_groups,
  group_ids = EXCLUDED.group_ids,
  bucket = EXCLUDED.bucket,
  updated_at = EXCLUDED.updated_at,
  updated_by_user_id = EXCLUDED.updated_by_user_id`
	if _, err := r.db.ExecContext(ctx, query, settings.Enabled, settings.AllGroups, string(groupIDsJSON), strings.TrimSpace(settings.Bucket), now, operatorID); err != nil {
		return nil, err
	}
	out := *settings
	out.UpdatedAt = now
	out.UpdatedByUserID = operatorID
	return &out, nil
}

func (r *promptArchiveRepository) InsertRecord(ctx context.Context, record *service.PromptArchivePersistedRecord) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("nil prompt archive repository")
	}
	if record == nil {
		return 0, fmt.Errorf("record cannot be nil")
	}
	attachmentsJSON, err := json.Marshal(record.Attachments)
	if err != nil {
		return 0, err
	}
	const query = `
INSERT INTO prompt_archive.ai_design_records (
  request_id, client_request_id, session_id, user_id, username_snapshot, email_snapshot,
  api_key_id, group_id, protocol, endpoint, model, system_prompt, user_prompt_text,
  prompt_summary, object_key, status, error_message, attachments_json, created_at
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19
) RETURNING id`
	var id int64
	err = r.db.QueryRowContext(
		ctx,
		query,
		record.RequestID,
		record.ClientRequestID,
		record.SessionID,
		record.UserID,
		record.UsernameSnapshot,
		record.EmailSnapshot,
		record.APIKeyID,
		record.GroupID,
		record.Protocol,
		record.Endpoint,
		record.Model,
		record.SystemPrompt,
		record.UserPromptText,
		record.PromptSummary,
		record.ObjectKey,
		string(record.Status),
		record.ErrorMessage,
		string(attachmentsJSON),
		record.CreatedAt.UTC(),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *promptArchiveRepository) ListRecords(ctx context.Context, filter *service.PromptArchiveRecordFilter) (*service.PromptArchiveRecordList, error) {
	if r == nil || r.db == nil {
		return &service.PromptArchiveRecordList{}, nil
	}
	if filter == nil {
		filter = &service.PromptArchiveRecordFilter{}
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	where, args := buildPromptArchiveRecordsWhere(filter)
	var total int
	countSQL := "SELECT COUNT(*) FROM prompt_archive.ai_design_records r " + where
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	selectSQL := `
SELECT id, request_id, client_request_id, session_id, user_id, username_snapshot, email_snapshot,
       api_key_id, group_id, protocol, endpoint, model, system_prompt, user_prompt_text,
       prompt_summary, object_key, status, error_message, attachments_json, created_at
FROM prompt_archive.ai_design_records r ` + where + `
ORDER BY created_at DESC
LIMIT $` + itoa(len(args)-1) + ` OFFSET $` + itoa(len(args))
	rows, err := r.db.QueryContext(ctx, selectSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*service.PromptArchiveRecordDetail, 0, pageSize)
	for rows.Next() {
		item, scanErr := scanPromptArchiveRecord(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &service.PromptArchiveRecordList{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (r *promptArchiveRepository) GetRecordByID(ctx context.Context, id int64) (*service.PromptArchiveRecordDetail, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil prompt archive repository")
	}
	const query = `
SELECT id, request_id, client_request_id, session_id, user_id, username_snapshot, email_snapshot,
       api_key_id, group_id, protocol, endpoint, model, system_prompt, user_prompt_text,
       prompt_summary, object_key, status, error_message, attachments_json, created_at
FROM prompt_archive.ai_design_records
WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	return scanPromptArchiveRecord(row)
}

func (r *promptArchiveRepository) GetSessionByID(ctx context.Context, sessionID string, groupID int64) (*service.PromptArchiveSessionDetail, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil prompt archive repository")
	}
	const query = `
SELECT id, request_id, client_request_id, session_id, user_id, username_snapshot, email_snapshot,
       api_key_id, group_id, protocol, endpoint, model, system_prompt, user_prompt_text,
       prompt_summary, object_key, status, error_message, attachments_json, created_at
FROM prompt_archive.ai_design_records
WHERE session_id = $1 AND group_id = $2
ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, strings.TrimSpace(sessionID), groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := &service.PromptArchiveSessionDetail{
		SessionID: strings.TrimSpace(sessionID),
		GroupID:   groupID,
		Records:   []*service.PromptArchiveRecordDetail{},
	}
	for rows.Next() {
		item, scanErr := scanPromptArchiveRecord(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out.Records = append(out.Records, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

type promptArchiveScanner interface {
	Scan(dest ...any) error
}

func scanPromptArchiveRecord(row promptArchiveScanner) (*service.PromptArchiveRecordDetail, error) {
	var (
		item            service.PromptArchiveRecordDetail
		status          string
		attachmentsJSON string
	)
	err := row.Scan(
		&item.ID,
		&item.RequestID,
		&item.ClientRequestID,
		&item.SessionID,
		&item.UserID,
		&item.UsernameSnapshot,
		&item.EmailSnapshot,
		&item.APIKeyID,
		&item.GroupID,
		&item.Protocol,
		&item.Endpoint,
		&item.Model,
		&item.SystemPrompt,
		&item.UserPromptText,
		&item.PromptSummary,
		&item.ObjectKey,
		&status,
		&item.ErrorMessage,
		&attachmentsJSON,
		&item.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	item.Status = service.PromptArchiveRecordStatus(status)
	if strings.TrimSpace(attachmentsJSON) != "" {
		_ = json.Unmarshal([]byte(attachmentsJSON), &item.Attachments)
	}
	if item.Attachments == nil {
		item.Attachments = []service.PromptArchiveAttachment{}
	}
	item.CreatedAt = item.CreatedAt.UTC()
	return &item, nil
}

func buildPromptArchiveRecordsWhere(filter *service.PromptArchiveRecordFilter) (string, []any) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	if q := strings.TrimSpace(filter.Query); q != "" {
		args = append(args, "%"+q+"%")
		n := itoa(len(args))
		clauses = append(clauses, "(r.username_snapshot ILIKE $"+n+" OR r.email_snapshot ILIKE $"+n+" OR r.session_id ILIKE $"+n+" OR r.model ILIKE $"+n+" OR r.prompt_summary ILIKE $"+n+")")
	}
	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}
