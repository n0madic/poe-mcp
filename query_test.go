package main

import (
	"context"
	"strings"
	"testing"
)

func TestUploadFilesValidation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		files   []string
		wantErr string
	}{
		{
			name:    "non-existent local file",
			files:   []string{"/no/such/file.txt"},
			wantErr: "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uploadFiles(ctx, tt.files, "fake-key")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantErr)
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
