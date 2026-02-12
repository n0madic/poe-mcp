package main

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestRunCLI_NoSubcommand(t *testing.T) {
	// When no subcommand is provided, runCLI shows help and returns nil
	err := runCLI([]string{})
	if err != nil {
		t.Errorf("Expected nil error when showing help, got: %v", err)
	}
}

func TestRunCLI_UnknownSubcommand(t *testing.T) {
	err := runCLI([]string{"badcommand"})
	if err == nil {
		t.Error("Expected error for unknown subcommand")
	}
	if !strings.Contains(err.Error(), "unknown subcommand") {
		t.Errorf("Expected 'unknown subcommand' error, got: %v", err)
	}
}

func TestRunQuery_MissingArgs(t *testing.T) {
	// Save and restore original POE_API_KEY
	origKey := os.Getenv("POE_API_KEY")
	defer os.Setenv("POE_API_KEY", origKey)

	// Set a dummy API key to bypass the API key check
	os.Setenv("POE_API_KEY", "test-key")

	tests := []struct {
		name string
		args []string
	}{
		{"no args", []string{}},
		{"only bot", []string{"GPT-4o"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runQuery(tt.args)
			if err == nil {
				t.Error("Expected error for missing query arguments")
			}
			if !strings.Contains(err.Error(), "usage:") {
				t.Errorf("Expected usage error, got: %v", err)
			}
		})
	}
}

func TestUploadCLIFiles_InvalidPath(t *testing.T) {
	ctx := context.Background()
	_, err := uploadCLIFiles(ctx, []string{"/no/such/file.txt"}, nil, "fake-key")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "open") {
		t.Errorf("expected open error, got: %v", err)
	}
}

func TestUploadCLIFiles_EmptySlices(t *testing.T) {
	ctx := context.Background()
	attachments, err := uploadCLIFiles(ctx, nil, nil, "fake-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(attachments) != 0 {
		t.Errorf("expected 0 attachments, got %d", len(attachments))
	}
}

func TestRunQuery_MissingAPIKey(t *testing.T) {
	// Save and restore original POE_API_KEY
	origKey := os.Getenv("POE_API_KEY")
	defer os.Setenv("POE_API_KEY", origKey)

	// Unset API key
	os.Unsetenv("POE_API_KEY")

	err := runQuery([]string{"GPT-4o", "Hello"})
	if err == nil {
		t.Error("Expected error when POE_API_KEY is not set")
	}
	if !strings.Contains(err.Error(), "POE_API_KEY") {
		t.Errorf("Expected POE_API_KEY error, got: %v", err)
	}
}
