package service

import "testing"

func TestExtractPromptArchiveOpenAIAttachments_IncludesVideoAndNestedImageURLShape(t *testing.T) {
	body := []byte(`{
		"input":[
			{"role":"user","content":[
				{"type":"input_image","image_url":"https://example.com/a.png"},
				{"type":"input_video","video_url":"https://example.com/demo.mp4"},
				{"type":"image_url","image_url":{"url":"https://example.com/b.jpg"}}
			]}
		]
	}`)

	attachments := extractPromptArchiveOpenAIAttachments(body)
	if len(attachments) != 3 {
		t.Fatalf("attachments=%d, want 3", len(attachments))
	}
	if attachments[1].Kind != PromptArchiveAttachmentKindVideo {
		t.Fatalf("attachments[1].kind=%s, want video", attachments[1].Kind)
	}
	if attachments[2].SourceURL != "https://example.com/b.jpg" {
		t.Fatalf("attachments[2].source_url=%q", attachments[2].SourceURL)
	}
}

func TestExtractPromptArchiveGeminiAttachments_RecognizesFileDataAndFileURI(t *testing.T) {
	body := []byte(`{
		"contents":[
			{"role":"user","parts":[
				{"inline_data":{"mime_type":"image/png","data":"aGVsbG8="}},
				{"file_data":{"mime_type":"video/mp4","file_uri":"https://example.com/v.mp4"}}
			]}
		]
	}`)

	attachments := extractPromptArchiveGeminiAttachments(body)
	if len(attachments) != 2 {
		t.Fatalf("attachments=%d, want 2", len(attachments))
	}
	if attachments[1].Kind != PromptArchiveAttachmentKindVideo {
		t.Fatalf("attachments[1].kind=%s, want video", attachments[1].Kind)
	}
	if attachments[1].SourceURL != "https://example.com/v.mp4" {
		t.Fatalf("attachments[1].source_url=%q", attachments[1].SourceURL)
	}
}
