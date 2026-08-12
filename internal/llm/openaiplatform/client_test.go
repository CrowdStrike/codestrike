package openaiplatform

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CrowdStrike/codestrike/internal/llm"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		apiKey  string
		modelID string
		wantErr bool
	}{
		{
			name:    "valid configuration",
			baseURL: "https://api.openai.com/v1",
			apiKey:  "sk-proj-test-key",
			modelID: "gpt-4o-mini",
			wantErr: false,
		},
		{
			name:    "empty API key still creates client",
			baseURL: "https://api.openai.com/v1",
			apiKey:  "",
			modelID: "gpt-4o-mini",
			wantErr: false,
		},
		{
			name:    "empty model ID still creates client",
			baseURL: "https://api.openai.com/v1",
			apiKey:  "sk-proj-test-key",
			modelID: "",
			wantErr: false,
		},
		{
			name:    "both empty still creates client",
			baseURL: "https://api.openai.com/v1",
			apiKey:  "",
			modelID: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client, err := NewClient(ctx, tt.baseURL, tt.apiKey, tt.modelID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewClient() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewClient() unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("NewClient() returned nil client")
				return
			}

			// Verify client configuration
			if client.ModelID != tt.modelID {
				t.Errorf("ModelID = %v, want %v", client.ModelID, tt.modelID)
			}

			if client.MaxRetries != 3 {
				t.Errorf("MaxRetries = %v, want 3", client.MaxRetries)
			}

			if client.InitialDelay != 100*time.Millisecond {
				t.Errorf("InitialDelay = %v, want 100ms", client.InitialDelay)
			}

			if client.MaxDelay != 12*time.Second {
				t.Errorf("MaxDelay = %v, want 12s", client.MaxDelay)
			}

			if client.Client == nil {
				t.Error("Client.Client (OpenAI client) is nil")
			}
		})
	}
}

func TestNewClient_ContextCancellation(t *testing.T) {
	// Test that cancelled context doesn't prevent client creation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client, err := NewClient(ctx, "https://api.openai.com/v1", "sk-test", "gpt-4o-mini")
	if err != nil {
		t.Errorf("NewClient() with cancelled context should still succeed, got error: %v", err)
	}

	if client == nil {
		t.Error("NewClient() returned nil client even with cancelled context")
	}
}

func TestClient_UsesCompatibleEndpointAndAPIKey(t *testing.T) {
	const apiKey = "compatible-provider-key"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("request path = %q, want /v1/chat/completions", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+apiKey {
			t.Errorf("Authorization = %q, want Bearer token", got)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"chatcmpl-test","object":"chat.completion","created":1,"model":"test-model","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`)
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), server.URL+"/v1", apiKey, "test-model")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	response, err := client.InvokeModel(context.Background(), llm.LLMRequest{
		Prompt:      "test",
		MaxTokens:   10,
		Temperature: 0,
	})
	if err != nil {
		t.Fatalf("InvokeModel() error = %v", err)
	}
	if response.Content != "ok" {
		t.Errorf("response content = %q, want ok", response.Content)
	}
}
