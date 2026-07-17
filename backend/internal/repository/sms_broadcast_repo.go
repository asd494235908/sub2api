package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

const smsBroadcastTemplateVarRowsKey = "__template_var_rows"

type smsBroadcastRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

func NewSMSBroadcastRepository(client *dbent.Client, sqlDB *sql.DB) service.SMSBroadcastRepository {
	return &smsBroadcastRepository{client: client, sql: sqlDB}
}

func (r *smsBroadcastRepository) exec(ctx context.Context) sqlQueryExecutor {
	return txAwareSQLExecutor(ctx, r.sql, r.client)
}

func (r *smsBroadcastRepository) CreateCampaign(ctx context.Context, campaign *service.SMSBroadcastCampaign) error {
	if campaign == nil {
		return nil
	}
	exec := r.exec(ctx)
	if exec == nil {
		return errors.New("sms broadcast repository is not configured")
	}
	audienceJSON, err := json.Marshal(campaign.Audience)
	if err != nil {
		return err
	}
	hasVarRowsColumn, err := r.hasTemplateVarRowsColumn(ctx, exec)
	if err != nil {
		return err
	}
	varsJSON, err := marshalSMSBroadcastTemplateVars(campaign.TemplateVars, campaign.TemplateVarRows, !hasVarRowsColumn)
	if err != nil {
		return err
	}
	varRowsJSON, err := json.Marshal(campaign.TemplateVarRows)
	if err != nil {
		return err
	}
	if !hasVarRowsColumn {
		query := `
			INSERT INTO sms_broadcast_campaigns (
				title, mode, template_id, body, rendered_body, status, audience, template_vars,
				total_recipients, sent_count, failed_count, skipped_count,
				error_message, created_by, updated_by, started_at, finished_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
			RETURNING id, created_at, updated_at
		`
		if err := scanSingleRow(ctx, exec, query, []any{
			campaign.Title,
			string(campaign.Mode),
			campaign.TemplateID,
			campaign.Body,
			campaign.RenderedBody,
			string(campaign.Status),
			audienceJSON,
			varsJSON,
			campaign.TotalRecipients,
			campaign.SentCount,
			campaign.FailedCount,
			campaign.SkippedCount,
			campaign.ErrorMessage,
			campaign.CreatedBy,
			campaign.UpdatedBy,
			campaign.StartedAt,
			campaign.FinishedAt,
		}, &campaign.ID, &campaign.CreatedAt, &campaign.UpdatedAt); err != nil {
			return err
		}
		return nil
	}
	query := `
		INSERT INTO sms_broadcast_campaigns (
			title, mode, template_id, body, rendered_body, status, audience, template_vars, template_var_rows,
			total_recipients, sent_count, failed_count, skipped_count,
			error_message, created_by, updated_by, started_at, finished_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at, updated_at
	`
	if err := scanSingleRow(ctx, exec, query, []any{
		campaign.Title,
		string(campaign.Mode),
		campaign.TemplateID,
		campaign.Body,
		campaign.RenderedBody,
		string(campaign.Status),
		audienceJSON,
		varsJSON,
		varRowsJSON,
		campaign.TotalRecipients,
		campaign.SentCount,
		campaign.FailedCount,
		campaign.SkippedCount,
		campaign.ErrorMessage,
		campaign.CreatedBy,
		campaign.UpdatedBy,
		campaign.StartedAt,
		campaign.FinishedAt,
	}, &campaign.ID, &campaign.CreatedAt, &campaign.UpdatedAt); err != nil {
		return err
	}
	return nil
}

func (r *smsBroadcastRepository) UpdateCampaign(ctx context.Context, campaign *service.SMSBroadcastCampaign) error {
	if campaign == nil {
		return nil
	}
	exec := r.exec(ctx)
	if exec == nil {
		return errors.New("sms broadcast repository is not configured")
	}
	audienceJSON, err := json.Marshal(campaign.Audience)
	if err != nil {
		return err
	}
	hasVarRowsColumn, err := r.hasTemplateVarRowsColumn(ctx, exec)
	if err != nil {
		return err
	}
	varsJSON, err := marshalSMSBroadcastTemplateVars(campaign.TemplateVars, campaign.TemplateVarRows, !hasVarRowsColumn)
	if err != nil {
		return err
	}
	varRowsJSON, err := json.Marshal(campaign.TemplateVarRows)
	if err != nil {
		return err
	}
	if !hasVarRowsColumn {
		query := `
			UPDATE sms_broadcast_campaigns
			SET title = $1,
				mode = $2,
				template_id = $3,
				body = $4,
				rendered_body = $5,
				status = $6,
				audience = $7,
				template_vars = $8,
				total_recipients = $9,
				sent_count = $10,
				failed_count = $11,
				skipped_count = $12,
				error_message = $13,
				updated_by = $14,
				started_at = $15,
				finished_at = $16,
				updated_at = NOW()
			WHERE id = $17
			RETURNING updated_at
		`
		if err := scanSingleRow(ctx, exec, query, []any{
			campaign.Title,
			string(campaign.Mode),
			campaign.TemplateID,
			campaign.Body,
			campaign.RenderedBody,
			string(campaign.Status),
			audienceJSON,
			varsJSON,
			campaign.TotalRecipients,
			campaign.SentCount,
			campaign.FailedCount,
			campaign.SkippedCount,
			campaign.ErrorMessage,
			campaign.UpdatedBy,
			campaign.StartedAt,
			campaign.FinishedAt,
			campaign.ID,
		}, &campaign.UpdatedAt); err != nil {
			return translatePersistenceError(err, service.ErrSMSBroadcastCampaignNotFound, nil)
		}
		return nil
	}
	query := `
		UPDATE sms_broadcast_campaigns
		SET title = $1,
			mode = $2,
			template_id = $3,
			body = $4,
			rendered_body = $5,
			status = $6,
			audience = $7,
			template_vars = $8,
			template_var_rows = $9,
			total_recipients = $10,
			sent_count = $11,
			failed_count = $12,
			skipped_count = $13,
			error_message = $14,
			updated_by = $15,
			started_at = $16,
			finished_at = $17,
			updated_at = NOW()
		WHERE id = $18
		RETURNING updated_at
	`
	if err := scanSingleRow(ctx, exec, query, []any{
		campaign.Title,
		string(campaign.Mode),
		campaign.TemplateID,
		campaign.Body,
		campaign.RenderedBody,
		string(campaign.Status),
		audienceJSON,
		varsJSON,
		varRowsJSON,
		campaign.TotalRecipients,
		campaign.SentCount,
		campaign.FailedCount,
		campaign.SkippedCount,
		campaign.ErrorMessage,
		campaign.UpdatedBy,
		campaign.StartedAt,
		campaign.FinishedAt,
		campaign.ID,
	}, &campaign.UpdatedAt); err != nil {
		return translatePersistenceError(err, service.ErrSMSBroadcastCampaignNotFound, nil)
	}
	return nil
}

func (r *smsBroadcastRepository) GetCampaignByID(ctx context.Context, id int64) (*service.SMSBroadcastCampaign, error) {
	exec := r.exec(ctx)
	if exec == nil {
		return nil, errors.New("sms broadcast repository is not configured")
	}
	hasVarRowsColumn, err := r.hasTemplateVarRowsColumn(ctx, exec)
	if err != nil {
		return nil, err
	}
	var query string
	if hasVarRowsColumn {
		query = `
		SELECT id, title, mode, template_id, body, rendered_body, status, audience,
			template_vars, total_recipients, sent_count, failed_count, skipped_count,
			error_message, created_by, updated_by, started_at, finished_at,
			created_at, updated_at, COALESCE(template_var_rows, '[]'::jsonb)
		FROM sms_broadcast_campaigns
		WHERE id = $1
	`
	} else {
		query = `
		SELECT id, title, mode, template_id, body, rendered_body, status, audience,
			template_vars, total_recipients, sent_count, failed_count, skipped_count,
			error_message, created_by, updated_by, started_at, finished_at,
			created_at, updated_at, '[]'::jsonb
		FROM sms_broadcast_campaigns
		WHERE id = $1
	`
	}
	campaign, err := scanSMSBroadcastCampaign(ctx, exec, query, []any{id})
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrSMSBroadcastCampaignNotFound, nil)
	}
	return campaign, nil
}

func (r *smsBroadcastRepository) ListCampaigns(ctx context.Context, params pagination.PaginationParams) ([]service.SMSBroadcastCampaign, *pagination.PaginationResult, error) {
	exec := r.exec(ctx)
	if exec == nil {
		return nil, nil, errors.New("sms broadcast repository is not configured")
	}
	var total int64
	if err := scanSingleRow(ctx, exec, "SELECT COUNT(*) FROM sms_broadcast_campaigns", nil, &total); err != nil {
		return nil, nil, err
	}
	if total == 0 {
		return []service.SMSBroadcastCampaign{}, paginationResultFromTotal(0, params), nil
	}
	hasVarRowsColumn, err := r.hasTemplateVarRowsColumn(ctx, exec)
	if err != nil {
		return nil, nil, err
	}
	var query string
	if hasVarRowsColumn {
		query = `
		SELECT id, title, mode, template_id, body, rendered_body, status, audience,
			template_vars, total_recipients, sent_count, failed_count, skipped_count,
			error_message, created_by, updated_by, started_at, finished_at,
			created_at, updated_at, COALESCE(template_var_rows, '[]'::jsonb)
		FROM sms_broadcast_campaigns
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`
	} else {
		query = `
		SELECT id, title, mode, template_id, body, rendered_body, status, audience,
			template_vars, total_recipients, sent_count, failed_count, skipped_count,
			error_message, created_by, updated_by, started_at, finished_at,
			created_at, updated_at, '[]'::jsonb
		FROM sms_broadcast_campaigns
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`
	}
	rows, err := exec.QueryContext(ctx, query, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.SMSBroadcastCampaign, 0)
	for rows.Next() {
		item, err := scanSMSBroadcastCampaignFromRows(rows)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *smsBroadcastRepository) AppendRecipients(ctx context.Context, campaignID int64, recipients []service.SMSBroadcastRecipient) error {
	if len(recipients) == 0 {
		return nil
	}
	exec := r.exec(ctx)
	if exec == nil {
		return errors.New("sms broadcast repository is not configured")
	}
	query := `
		INSERT INTO sms_broadcast_recipients (
			campaign_id, user_id, phone_number, raw_phone, rendered_body, status, error_message, sent_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (campaign_id, phone_number) DO NOTHING
	`
	for i := range recipients {
		status := recipients[i].Status
		if status == "" {
			status = "queued"
		}
		if _, err := exec.ExecContext(ctx, query,
			campaignID,
			recipients[i].UserID,
			recipients[i].PhoneNumber,
			recipients[i].RawPhone,
			recipients[i].RenderedBody,
			status,
			recipients[i].ErrorMessage,
			recipients[i].SentAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *smsBroadcastRepository) ListRecipients(ctx context.Context, campaignID int64) ([]service.SMSBroadcastRecipient, error) {
	exec := r.exec(ctx)
	if exec == nil {
		return nil, errors.New("sms broadcast repository is not configured")
	}
	query := `
		SELECT user_id, phone_number, raw_phone, rendered_body, status, error_message, sent_at, created_at, updated_at
		FROM sms_broadcast_recipients
		WHERE campaign_id = $1
		ORDER BY id ASC
	`
	rows, err := exec.QueryContext(ctx, query, campaignID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.SMSBroadcastRecipient, 0)
	for rows.Next() {
		var item service.SMSBroadcastRecipient
		var errMsg sql.NullString
		var sentAt sql.NullTime
		if err := rows.Scan(
			&item.UserID,
			&item.PhoneNumber,
			&item.RawPhone,
			&item.RenderedBody,
			&item.Status,
			&errMsg,
			&sentAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if errMsg.Valid {
			item.ErrorMessage = &errMsg.String
		}
		if sentAt.Valid {
			item.SentAt = &sentAt.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *smsBroadcastRepository) ListRecipientsPaginated(ctx context.Context, campaignID int64, params pagination.PaginationParams, status string) ([]service.SMSBroadcastRecipient, *pagination.PaginationResult, error) {
	exec := r.exec(ctx)
	if exec == nil {
		return nil, nil, errors.New("sms broadcast repository is not configured")
	}
	where := "WHERE campaign_id = $1"
	args := []any{campaignID}
	if strings.TrimSpace(status) != "" {
		where += " AND status = $2"
		args = append(args, strings.TrimSpace(status))
	}
	var total int64
	countQuery := "SELECT COUNT(*) FROM sms_broadcast_recipients " + where
	if err := scanSingleRow(ctx, exec, countQuery, args, &total); err != nil {
		return nil, nil, err
	}
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	listArgs := append(append([]any(nil), args...), pageSize, offset)
	query := `
		SELECT user_id, phone_number, raw_phone, rendered_body, status, error_message, sent_at, created_at, updated_at
		FROM sms_broadcast_recipients
		` + where + `
		ORDER BY id ASC
		LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
	rows, err := exec.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.SMSBroadcastRecipient, 0)
	for rows.Next() {
		var item service.SMSBroadcastRecipient
		var errMsg sql.NullString
		var sentAt sql.NullTime
		if err := rows.Scan(
			&item.UserID,
			&item.PhoneNumber,
			&item.RawPhone,
			&item.RenderedBody,
			&item.Status,
			&errMsg,
			&sentAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}
		if errMsg.Valid {
			item.ErrorMessage = &errMsg.String
		}
		if sentAt.Valid {
			item.SentAt = &sentAt.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	result := paginationResultFromTotal(total, params)
	return items, result, nil
}

func (r *smsBroadcastRepository) UpdateRecipient(ctx context.Context, campaignID int64, recipient *service.SMSBroadcastRecipient) error {
	if recipient == nil {
		return nil
	}
	exec := r.exec(ctx)
	if exec == nil {
		return errors.New("sms broadcast repository is not configured")
	}
	query := `
		UPDATE sms_broadcast_recipients
		SET rendered_body = $1,
			status = $2,
			error_message = $3,
			sent_at = $4,
			updated_at = NOW()
		WHERE campaign_id = $5 AND phone_number = $6
	`
	_, err := exec.ExecContext(ctx, query, recipient.RenderedBody, recipient.Status, recipient.ErrorMessage, recipient.SentAt, campaignID, recipient.PhoneNumber)
	return err
}

type smsBroadcastRowScanner interface {
	Scan(dest ...any) error
}

func scanSMSBroadcastCampaign(ctx context.Context, exec sqlQueryExecutor, query string, args []any) (*service.SMSBroadcastCampaign, error) {
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}
	return scanSMSBroadcastCampaignFromRows(rows)
}

func scanSMSBroadcastCampaignFromRows(row smsBroadcastRowScanner) (*service.SMSBroadcastCampaign, error) {
	var item service.SMSBroadcastCampaign
	var mode string
	var status string
	var audienceJSON []byte
	var varsJSON []byte
	var varRowsJSON []byte
	var errMsg sql.NullString
	var createdBy sql.NullInt64
	var updatedBy sql.NullInt64
	var startedAt sql.NullTime
	var finishedAt sql.NullTime
	if err := row.Scan(
		&item.ID,
		&item.Title,
		&mode,
		&item.TemplateID,
		&item.Body,
		&item.RenderedBody,
		&status,
		&audienceJSON,
		&varsJSON,
		&item.TotalRecipients,
		&item.SentCount,
		&item.FailedCount,
		&item.SkippedCount,
		&errMsg,
		&createdBy,
		&updatedBy,
		&startedAt,
		&finishedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&varRowsJSON,
	); err != nil {
		return nil, err
	}
	item.Mode = service.SMSBroadcastMode(mode)
	item.Status = service.SMSBroadcastStatus(status)
	if len(audienceJSON) > 0 {
		if err := json.Unmarshal(audienceJSON, &item.Audience); err != nil {
			return nil, fmt.Errorf("parse sms broadcast audience: %w", err)
		}
	}
	if len(varsJSON) > 0 {
		varRows, err := unmarshalSMSBroadcastTemplateVars(varsJSON, &item.TemplateVars)
		if err != nil {
			return nil, fmt.Errorf("parse sms broadcast template vars: %w", err)
		}
		item.TemplateVarRows = varRows
	}
	if len(varRowsJSON) > 0 {
		var rows []service.SMSBroadcastTemplateVarRow
		if err := json.Unmarshal(varRowsJSON, &rows); err != nil {
			return nil, fmt.Errorf("parse sms broadcast template var rows: %w", err)
		}
		if len(rows) > 0 {
			item.TemplateVarRows = rows
		}
	}
	if len(item.TemplateVarRows) == 0 && len(item.TemplateVars) > 0 {
		keys := make([]string, 0, len(item.TemplateVars))
		for key := range item.TemplateVars {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			item.TemplateVarRows = append(item.TemplateVarRows, service.SMSBroadcastTemplateVarRow{Key: key, Value: item.TemplateVars[key]})
		}
	}
	if errMsg.Valid {
		item.ErrorMessage = &errMsg.String
	}
	if createdBy.Valid {
		v := createdBy.Int64
		item.CreatedBy = &v
	}
	if updatedBy.Valid {
		v := updatedBy.Int64
		item.UpdatedBy = &v
	}
	if startedAt.Valid {
		item.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		item.FinishedAt = &finishedAt.Time
	}
	return &item, nil
}

func (r *smsBroadcastRepository) hasTemplateVarRowsColumn(ctx context.Context, exec sqlQueryExecutor) (bool, error) {
	var exists bool
	err := scanSingleRow(ctx, exec, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_name = $1 AND column_name = $2
		)
	`, []any{"sms_broadcast_campaigns", "template_var_rows"}, &exists)
	return exists, err
}

func marshalSMSBroadcastTemplateVars(vars map[string]string, rows []service.SMSBroadcastTemplateVarRow, includeRows bool) ([]byte, error) {
	if !includeRows {
		if vars == nil {
			vars = map[string]string{}
		}
		return json.Marshal(vars)
	}
	payload := make(map[string]any, len(vars)+1)
	for key, value := range vars {
		if key == smsBroadcastTemplateVarRowsKey {
			continue
		}
		payload[key] = value
	}
	if len(rows) > 0 {
		payload[smsBroadcastTemplateVarRowsKey] = rows
	}
	return json.Marshal(payload)
}

func unmarshalSMSBroadcastTemplateVars(data []byte, out *map[string]string) ([]service.SMSBroadcastTemplateVarRow, error) {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	vars := make(map[string]string, len(raw))
	var rows []service.SMSBroadcastTemplateVarRow
	for key, value := range raw {
		if key == smsBroadcastTemplateVarRowsKey {
			if err := json.Unmarshal(value, &rows); err != nil {
				return nil, err
			}
			continue
		}
		var str string
		if err := json.Unmarshal(value, &str); err != nil {
			return nil, err
		}
		vars[key] = str
	}
	*out = vars
	return rows, nil
}
