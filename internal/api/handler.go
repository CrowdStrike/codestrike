package api

import (
	"net/http"

	"github.com/CrowdStrike/codestrike/internal/config"
	"github.com/CrowdStrike/codestrike/internal/llm"
	"github.com/emicklei/go-restful/v3"
	"github.com/rs/zerolog"
)

type Handler struct {
	llmClient llm.LLMClient
	appConfig config.Config
	ghToken   string
	logger    *zerolog.Logger
}

func NewHandler(llmClient llm.LLMClient, appConfig *config.Config, ghToken string, log *zerolog.Logger) *Handler {
	return &Handler{
		llmClient: llmClient,
		appConfig: *appConfig,
		ghToken:   ghToken,
		logger:    log,
	}
}

// Healthz handles GET /healthz
func (h *Handler) Healthz(req *restful.Request, resp *restful.Response) {
	healthResponse := HealthResponse{
		Status:  "ok",
		Version: "1.0.0",
	}

	resp.WriteHeaderAndEntity(http.StatusOK, healthResponse)
}
