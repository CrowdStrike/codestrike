package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/rs/zerolog"

	"github.com/CrowdStrike/codestrike/internal/api"
	"github.com/CrowdStrike/codestrike/internal/config"
	"github.com/CrowdStrike/codestrike/internal/setup"
)

func setupContainer(t *testing.T) *restful.Container {
	t.Helper()
	log := zerolog.Nop()
	handler := api.NewHandler(nil, &config.Config{}, &setup.Config{}, &log)
	container := restful.NewContainer()
	api.RegisterRoutes(container, handler)
	return container
}

func TestHealthz(t *testing.T) {
	container := setupContainer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp api.HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status %q, got %q", "ok", resp.Status)
	}

	if resp.Version == "" {
		t.Error("expected non-empty version")
	}
}

func TestHealthz_MethodNotAllowed(t *testing.T) {
	container := setupContainer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/healthz", nil)
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHealthz_ContentType(t *testing.T) {
	container := setupContainer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", ct)
	}
}

func TestReview_EmptyBody(t *testing.T) {
	container := setupContainer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/review", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var resp api.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != "pr_url is required" {
		t.Errorf("expected %q, got %q", "pr_url is required", resp.Error)
	}
}

func TestReview_InvalidPRURL(t *testing.T) {
	container := setupContainer(t)

	body := `{"pr_url": "not-a-valid-url"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/review", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var resp api.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !strings.Contains(resp.Error, "invalid PR URL") {
		t.Errorf("expected error containing %q, got %q", "invalid PR URL", resp.Error)
	}
}

func TestReview_InvalidJSON(t *testing.T) {
	container := setupContainer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/review", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var resp api.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !strings.Contains(resp.Error, "invalid request body") {
		t.Errorf("expected error containing %q, got %q", "invalid request body", resp.Error)
	}
}

func TestReview_MethodNotAllowed(t *testing.T) {
	container := setupContainer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/review", nil)
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
