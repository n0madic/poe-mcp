package main

import (
	"strings"
	"testing"

	"github.com/n0madic/go-poe/models"
)

func sampleModels() []models.Model {
	str := func(s string) *string { return &s }
	intP := func(i int) *int { return &i }
	return []models.Model{
		{
			ID:          "gpt-4o",
			Description: "OpenAI's flagship multimodal model",
			OwnedBy:     "OpenAI",
			Architecture: models.Architecture{
				Modality: "text,image->text",
			},
			Metadata: models.ModelMetadata{
				DisplayName: "GPT-4o",
			},
			ContextWindow: &models.ContextWindow{
				ContextLength:   128000,
				MaxOutputTokens: intP(16384),
			},
			Pricing: &models.Pricing{
				Prompt:     str("0.000005"),
				Completion: str("0.000015"),
			},
		},
		{
			ID:          "claude-4.5-sonnet",
			Description: "Anthropic's balanced model",
			OwnedBy:     "Anthropic",
			Architecture: models.Architecture{
				Modality: "text,image->text",
			},
			Metadata: models.ModelMetadata{
				DisplayName: "Claude 3.5 Sonnet",
			},
			ContextWindow: &models.ContextWindow{
				ContextLength:   200000,
				MaxOutputTokens: intP(8192),
			},
			Pricing: &models.Pricing{
				Prompt:     str("0.000003"),
				Completion: str("0.000015"),
			},
		},
		{
			ID:          "dall-e-3",
			Description: "OpenAI's image generation model",
			OwnedBy:     "OpenAI",
			Architecture: models.Architecture{
				Modality: "text->image",
			},
			Metadata: models.ModelMetadata{
				DisplayName: "DALL-E 3",
			},
			Pricing: &models.Pricing{
				Image: str("0.04"),
			},
		},
		{
			ID:          "gemini-2.5-pro",
			Description: "Google's advanced model with reasoning",
			OwnedBy:     "Google",
			Architecture: models.Architecture{
				Modality: "text,image,video->text",
			},
			Metadata: models.ModelMetadata{
				DisplayName: "Gemini 2.5 Pro",
			},
			ContextWindow: &models.ContextWindow{
				ContextLength:   1048576,
				MaxOutputTokens: intP(65536),
			},
			Pricing: &models.Pricing{
				Prompt:     str("0.00000125"),
				Completion: str("0.00001"),
			},
		},
	}
}

func TestFilterModelsNoFilters(t *testing.T) {
	all := sampleModels()
	result := filterModels(all, SearchModelsArgs{})
	if len(result) != len(all) {
		t.Errorf("expected %d models, got %d", len(all), len(result))
	}
}

func TestFilterModelsByQuery(t *testing.T) {
	all := sampleModels()

	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{"match by ID", "gpt", 1},
		{"match by display name", "claude", 1},
		{"match by description", "flagship", 1},
		{"match by owner", "google", 1},
		{"match multiple", "openai", 2},
		{"no match", "nonexistent", 0},
		{"case insensitive", "GPT", 1},
		{"multi-word query", "GPT 4o", 1},
		{"multi-word across fields", "openai image", 1},
		{"multi-word no match", "google codex", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterModels(all, SearchModelsArgs{Query: tt.query})
			if len(result) != tt.expected {
				t.Errorf("query=%q: expected %d models, got %d", tt.query, tt.expected, len(result))
			}
		})
	}
}

func TestFilterModelsByOwnedBy(t *testing.T) {
	all := sampleModels()

	result := filterModels(all, SearchModelsArgs{OwnedBy: "OpenAI"})
	if len(result) != 2 {
		t.Errorf("expected 2 OpenAI models, got %d", len(result))
	}

	result = filterModels(all, SearchModelsArgs{OwnedBy: "anthropic"})
	if len(result) != 1 {
		t.Errorf("expected 1 Anthropic model, got %d", len(result))
	}
}

func TestFilterModelsByModality(t *testing.T) {
	all := sampleModels()

	result := filterModels(all, SearchModelsArgs{Modality: "image"})
	if len(result) != 4 {
		t.Errorf("expected 4 models with image modality, got %d", len(result))
	}

	result = filterModels(all, SearchModelsArgs{Modality: "video"})
	if len(result) != 1 {
		t.Errorf("expected 1 model with video modality, got %d", len(result))
	}
}

func TestFilterModelsCombined(t *testing.T) {
	all := sampleModels()

	result := filterModels(all, SearchModelsArgs{Query: "openai", Modality: "image"})
	if len(result) != 2 {
		t.Errorf("expected 2 OpenAI models with image, got %d", len(result))
	}

	result = filterModels(all, SearchModelsArgs{OwnedBy: "OpenAI", Modality: "text->image"})
	if len(result) != 1 {
		t.Errorf("expected 1 model, got %d", len(result))
	}
	if result[0].ID != "dall-e-3" {
		t.Errorf("expected dall-e-3, got %s", result[0].ID)
	}
}

func TestFormatModel(t *testing.T) {
	m := sampleModels()[0] // gpt-4o
	output := formatModel(m)

	expected := []string{
		"## gpt-4o",
		"Display Name: GPT-4o",
		"Owner: OpenAI",
		"Modality: text,image->text",
		"Context Length: 128000 tokens",
		"Max Output: 16384 tokens",
		"prompt=0.000005",
		"completion=0.000015",
	}

	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("formatModel output missing %q\ngot: %s", s, output)
		}
	}
}

func TestFormatModelMinimal(t *testing.T) {
	m := models.Model{
		ID:      "test-model",
		OwnedBy: "TestOrg",
	}
	output := formatModel(m)

	if !strings.Contains(output, "## test-model") {
		t.Error("missing model ID header")
	}
	if !strings.Contains(output, "Owner: TestOrg") {
		t.Error("missing owner")
	}
	if strings.Contains(output, "Display Name:") {
		t.Error("should not show display name when same as empty")
	}
}
