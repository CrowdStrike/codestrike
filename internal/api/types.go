package api

type HealthResponse struct {
	Status  string `json:"status" description:"Service status"`
	Version string `json:"version" description:"API version"`
}

type ReviewRequest struct {
	PrUrl      string `json:"pr_url"`
	FullContext bool   `json:"full_context,omitempty"`
}

type ReviewResponse struct {
	Status string `json:"status"`
	Owner  string `json:"owner"`
	Repo   string `json:"repo"`
	PR     int    `json:"pr"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
