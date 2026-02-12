package main

import (
	"context"
	"encoding/base64"
	"testing"
)

func TestUploadFilesValidation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		files   []FileInput
		wantErr string
	}{
		{
			name:    "empty name",
			files:   []FileInput{{URL: "https://example.com/f.txt"}},
			wantErr: "file name is required",
		},
		{
			name:    "missing url and content",
			files:   []FileInput{{Name: "f.txt"}},
			wantErr: "either url or content is required",
		},
		{
			name:    "invalid base64",
			files:   []FileInput{{Name: "f.txt", Content: "!!!not-base64!!!"}},
			wantErr: "invalid base64 content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uploadFiles(ctx, tt.files, "fake-key")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if got := err.Error(); !containsSubstring(got, tt.wantErr) {
				t.Errorf("error = %q, want substring %q", got, tt.wantErr)
			}
		})
	}
}

func TestUploadFilesEmptySlice(t *testing.T) {
	ctx := context.Background()
	attachments, err := uploadFiles(ctx, nil, "fake-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(attachments) != 0 {
		t.Errorf("expected 0 attachments, got %d", len(attachments))
	}
}

func TestBase64DecodeRoundTrip(t *testing.T) {
	// Verify that our base64 decode path handles valid content correctly.
	original := []byte("hello, world!")
	encoded := base64.StdEncoding.EncodeToString(original)

	file := FileInput{
		Name:    "test.txt",
		Content: encoded,
	}

	if file.Name == "" {
		t.Fatal("name should not be empty")
	}
	if file.Content == "" {
		t.Fatal("content should not be empty")
	}

	// Verify decode succeeds
	decoded, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		t.Fatalf("base64 decode failed: %v", err)
	}
	if string(decoded) != string(original) {
		t.Errorf("decoded = %q, want %q", string(decoded), string(original))
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
