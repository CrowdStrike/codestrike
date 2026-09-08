package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/rs/zerolog"

	"github.com/CrowdStrike/codestrike/internal/api"
	"github.com/CrowdStrike/codestrike/internal/config"
)

func setupContainer(t *testing.T) *restful.Container {
	t.Helper()
	log := zerolog.Nop()
	handler := api.NewHandler(nil, &config.Config{}, "", &log)
	container := restful.NewContainer()
	api.RegisterRoutes(container, handler)
	return container
}

func TestHealthz(t *testing.T) {
	container := setupContainer(t)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
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

	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHealthz_ContentType(t *testing.T) {
	container := setupContainer(t)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", ct)
	}
}
