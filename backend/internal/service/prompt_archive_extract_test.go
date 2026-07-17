package service

import (
	"testing"
	"time"
)

func TestBuildPromptArchiveEnvelopeFromParsedRequest(t *testing.T) {
	body := []byte(`{
		"model":"claude-sonnet-4-5",
		"metadata":{"user_id":"user_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_account__session_123e4567-e89b-12d3-a456-426614174000"},
		"system":[{"type":"text","text":"You are helpful"}],
		"messages":[{"role":"user","content":[{"type":"text","text":"hello world"}]}]
	}`)
	parsed, err := ParseGatewayRequest(NewRequestBodyRef(body), "")
	if err != nil {
		t.Fatalf("ParseGatewayRequest error: %v", err)
	}
	apiKey := &APIKey{
		ID:      41,
		UserID:  7,
		GroupID: promptArchiveInt64Ptr(9),
		User: &User{
			ID:       7,
			Username: "alice",
			Email:    "alice@example.com",
		},
	}

	env := BuildPromptArchiveEnvelopeFromParsedRequest(apiKey, "/v1/messages", "anthropic", parsed, time.Now().UTC())
	if env == nil {
		t.Fatalf("env should not be nil")
	}
	if env.SessionID != "123e4567-e89b-12d3-a456-426614174000" {
		t.Fatalf("session_id=%q", env.SessionID)
	}
	if env.UserPromptText != "hello world" {
		t.Fatalf("user prompt=%q", env.UserPromptText)
	}
	if env.SystemPrompt != "You are helpful" {
		t.Fatalf("system prompt=%q", env.SystemPrompt)
	}
}

func TestBuildPromptArchiveEnvelopeFromOpenAIResponsesBody(t *testing.T) {
	apiKey := &APIKey{
		ID:      51,
		UserID:  8,
		GroupID: promptArchiveInt64Ptr(10),
		User: &User{
			ID:       8,
			Username: "bob",
			Email:    "bob@example.com",
		},
	}
	body := []byte(`{
		"model":"gpt-5.4",
		"prompt_cache_key":"openai-session-1",
		"input":[
			{"role":"user","content":[
				{"type":"input_text","text":"describe this"},
				{"type":"input_image","image_url":"https://example.com/image.png"}
			]}
		]
	}`)

	env := BuildPromptArchiveEnvelopeFromOpenAIResponsesBody(apiKey, "/v1/responses", body, time.Now().UTC())
	if env == nil {
		t.Fatalf("env should not be nil")
	}
	if env.SessionID != "openai-session-1" {
		t.Fatalf("session_id=%q", env.SessionID)
	}
	if env.UserPromptText != "describe this" {
		t.Fatalf("user prompt=%q", env.UserPromptText)
	}
	if len(env.Attachments) != 1 {
		t.Fatalf("attachments=%d, want 1", len(env.Attachments))
	}
}

func TestBuildPromptArchiveEnvelopeFromGeminiBody(t *testing.T) {
	apiKey := &APIKey{
		ID:      61,
		UserID:  9,
		GroupID: promptArchiveInt64Ptr(11),
		User: &User{
			ID:       9,
			Username: "cathy",
			Email:    "cathy@example.com",
		},
	}
	body := []byte(`{
		"contents":[
			{"role":"user","parts":[
				{"text":"summarize this image"},
				{"inline_data":{"mime_type":"image/png","data":"aGVsbG8="}}
			]}
		]
	}`)

	env := BuildPromptArchiveEnvelopeFromGeminiBody(apiKey, "/v1beta/models/gemini:generateContent", "gemini-2.5-pro", "gemini-session-1", body, time.Now().UTC())
	if env == nil {
		t.Fatalf("env should not be nil")
	}
	if env.SessionID != "gemini-session-1" {
		t.Fatalf("session_id=%q", env.SessionID)
	}
	if env.UserPromptText != "summarize this image" {
		t.Fatalf("user prompt=%q", env.UserPromptText)
	}
	if len(env.Attachments) != 1 {
		t.Fatalf("attachments=%d, want 1", len(env.Attachments))
	}
}

func promptArchiveInt64Ptr(v int64) *int64 {
	return &v
}
