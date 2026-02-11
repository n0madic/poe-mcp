package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/n0madic/go-poe/models"
)

const cacheTTL = 15 * time.Minute

// SearchModelsArgs defines the input schema for the search_models tool.
type SearchModelsArgs struct {
	Query    string `json:"query,omitempty" jsonschema:"Search query — matches model ID, display name, description, and owner (case-insensitive substring match)"`
	OwnedBy  string `json:"owned_by,omitempty" jsonschema:"Filter by owner/provider (e.g. OpenAI, Anthropic, Google, Meta)"`
	Modality string `json:"modality,omitempty" jsonschema:"Filter by modality substring (e.g. text, image, video)"`
}

// modelCache provides an in-memory cache for the Poe model catalog.
type modelCache struct {
	mu        sync.RWMutex
	models    []models.Model
	fetchedAt time.Time
}

var cache = &modelCache{}

func (c *modelCache) get(ctx context.Context) ([]models.Model, error) {
	c.mu.RLock()
	if len(c.models) > 0 && time.Since(c.fetchedAt) < cacheTTL {
		defer c.mu.RUnlock()
		return c.models, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock.
	if len(c.models) > 0 && time.Since(c.fetchedAt) < cacheTTL {
		return c.models, nil
	}

	fetched, err := models.Fetch(ctx, nil)
	if err != nil {
		return nil, err
	}

	c.models = fetched
	c.fetchedAt = time.Now()
	return c.models, nil
}

func registerSearchModels(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_models",
		Description: "Search and filter the Poe.com model catalog by name, owner, or modality",
	}, handleSearchModels)
}

func handleSearchModels(ctx context.Context, req *mcp.CallToolRequest, args SearchModelsArgs) (*mcp.CallToolResult, any, error) {
	all, err := cache.get(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error fetching models: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	matched := filterModels(all, args)

	if len(matched) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No models found matching the given criteria."},
			},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatModels(matched)},
		},
	}, nil, nil
}

// filterModels filters the model list by the search criteria in args.
func filterModels(all []models.Model, args SearchModelsArgs) []models.Model {
	query := strings.ToLower(args.Query)
	ownedBy := strings.ToLower(args.OwnedBy)
	modality := strings.ToLower(args.Modality)

	var result []models.Model
	for _, m := range all {
		if query != "" && !matchesQuery(m, query) {
			continue
		}
		if ownedBy != "" && strings.ToLower(m.OwnedBy) != ownedBy {
			continue
		}
		if modality != "" && !strings.Contains(strings.ToLower(m.Architecture.Modality), modality) {
			continue
		}
		result = append(result, m)
	}
	return result
}

// matchesQuery checks if all words in the query appear somewhere across the model's searchable fields.
func matchesQuery(m models.Model, query string) bool {
	combined := strings.ToLower(strings.Join([]string{
		m.ID, m.Metadata.DisplayName, m.Description, m.OwnedBy,
	}, " "))
	for _, word := range strings.Fields(query) {
		if !strings.Contains(combined, word) {
			return false
		}
	}
	return true
}

// formatModels formats a slice of models as readable text.
func formatModels(matched []models.Model) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d model(s):\n\n", len(matched))
	for i, m := range matched {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(formatModel(m))
	}
	return sb.String()
}

// formatModel formats a single model as a readable text block.
func formatModel(m models.Model) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "## %s\n", m.ID)
	if m.Metadata.DisplayName != "" && m.Metadata.DisplayName != m.ID {
		fmt.Fprintf(&sb, "Display Name: %s\n", m.Metadata.DisplayName)
	}
	fmt.Fprintf(&sb, "Owner: %s\n", m.OwnedBy)
	if m.Architecture.Modality != "" {
		fmt.Fprintf(&sb, "Modality: %s\n", m.Architecture.Modality)
	}
	if m.Description != "" {
		fmt.Fprintf(&sb, "Description: %s\n", m.Description)
	}
	if m.ContextWindow != nil {
		fmt.Fprintf(&sb, "Context Length: %d tokens\n", m.ContextWindow.ContextLength)
		if m.ContextWindow.MaxOutputTokens != nil {
			fmt.Fprintf(&sb, "Max Output: %d tokens\n", *m.ContextWindow.MaxOutputTokens)
		}
	}
	if m.Pricing != nil {
		sb.WriteString("Pricing:")
		if m.Pricing.Prompt != nil {
			fmt.Fprintf(&sb, " prompt=%s", *m.Pricing.Prompt)
		}
		if m.Pricing.Completion != nil {
			fmt.Fprintf(&sb, " completion=%s", *m.Pricing.Completion)
		}
		if m.Pricing.Request != nil {
			fmt.Fprintf(&sb, " request=%s", *m.Pricing.Request)
		}
		if m.Pricing.Image != nil {
			fmt.Fprintf(&sb, " image=%s", *m.Pricing.Image)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
