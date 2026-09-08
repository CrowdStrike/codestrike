package api

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/rs/zerolog"

	"github.com/CrowdStrike/codestrike/internal/config"
	"github.com/CrowdStrike/codestrike/internal/llm"
	"github.com/CrowdStrike/codestrike/internal/review"
	"github.com/CrowdStrike/codestrike/internal/scm/github"
	"github.com/CrowdStrike/codestrike/internal/tokenizer"
)

type Handler struct {
	llmClient llm.LLMClient
	appConfig *config.Config
	ghToken   string
	logger    *zerolog.Logger
}

func NewHandler(llmClient llm.LLMClient, appConfig *config.Config, ghToken string, log *zerolog.Logger) *Handler {
	return &Handler{
		llmClient: llmClient,
		appConfig: appConfig,
		ghToken:   ghToken,
		logger:    log,
	}
}

// Healthz handles GET /healthz.
func (h *Handler) Healthz(req *restful.Request, resp *restful.Response) {
	healthResponse := HealthResponse{
		Status:  "ok",
		Version: "1.0.0",
	}

	_ = resp.WriteHeaderAndEntity(http.StatusOK, healthResponse)
}

// Review handles POST /api/v1/review.
func (h *Handler) Review(req *restful.Request, resp *restful.Response) {
	var body ReviewRequest
	if err := req.ReadEntity(&body); err != nil {
		_ = resp.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{
			Error: fmt.Sprintf("invalid request body: %v", err),
		})
		return
	}

	if body.PrUrl == "" {
		_ = resp.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{
			Error: "pr_url is required",
		})
		return
	}

	ref, err := review.ParsePrUrl(body.PrUrl)
	if err != nil {
		_ = resp.WriteHeaderAndEntity(http.StatusBadRequest, ErrorResponse{
			Error: fmt.Sprintf("invalid PR URL: %v", err),
		})
		return
	}

	ghClient := github.New(github.Config{
		Owner:   ref.Owner,
		Repo:    ref.Repo,
		Token:   h.ghToken,
		BaseURL: h.appConfig.GitHub.BaseURL,
	})

	tok := tokenizer.NewForModel(h.appConfig.Review.Context.TokenizerModel)
	pipeline := review.NewPipeline(ghClient, h.llmClient, h.appConfig, tok, h.logger, review.Options{
		FullContext: body.FullContext,
	})

	if err := pipeline.Run(req.Request.Context(), ref); err != nil {
		h.logger.Error().Err(err).Str("pr_url", body.PrUrl).Msg("review failed")
		_ = resp.WriteHeaderAndEntity(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("review failed: %v", err),
		})
		return
	}

	_ = resp.WriteHeaderAndEntity(http.StatusOK, ReviewResponse{
		Status: "reviewed",
		Owner:  ref.Owner,
		Repo:   ref.Repo,
		PR:     ref.Number,
	})
}
