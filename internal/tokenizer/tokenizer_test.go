package tokenizer_test

import (
	"strings"
	"testing"

	"github.com/CrowdStrike/codestrike/internal/tokenizer"
)

func TestCountTokens(t *testing.T) {
	tok := tokenizer.New()

	tests := []struct {
		name     string
		input    string
		minCount int
		maxCount int
	}{
		{
			name:     "empty string",
			input:    "",
			minCount: 0,
			maxCount: 0,
		},
		{
			name:     "single word",
			input:    "hello",
			minCount: 1,
			maxCount: 2,
		},
		{
			name:     "short sentence",
			input:    "The quick brown fox jumps over the lazy dog.",
			minCount: 8,
			maxCount: 12,
		},
		{
			name:     "code snippet",
			input:    "func main() {\n\tfmt.Println(\"hello world\")\n}",
			minCount: 8,
			maxCount: 20,
		},
		{
			name:     "long text scales linearly",
			input:    strings.Repeat("token ", 1000),
			minCount: 900,
			maxCount: 1100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := tok.CountTokens(tt.input)
			if count < tt.minCount || count > tt.maxCount {
				t.Errorf("CountTokens() = %d, expected between %d and %d", count, tt.minCount, tt.maxCount)
			}
		})
	}
}

func TestGetModelLimits(t *testing.T) {
	tests := []struct {
		modelID       string
		wantWindow    int
		wantMaxOutput int
	}{
		{"gpt-4o", 128000, 16384},
		{"claude-sonnet-4-20250514", 200000, 8192},
		{"llama3", 8192, 2048},
		{"unknown-model", 8192, 4096},
	}

	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			limits := tokenizer.GetModelLimits(tt.modelID)
			if limits.ContextWindow != tt.wantWindow {
				t.Errorf("ContextWindow = %d, want %d", limits.ContextWindow, tt.wantWindow)
			}
			if limits.MaxOutputTokens != tt.wantMaxOutput {
				t.Errorf("MaxOutputTokens = %d, want %d", limits.MaxOutputTokens, tt.wantMaxOutput)
			}
		})
	}
}
