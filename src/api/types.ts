export interface HealthResponse {
  status: string;
  version: string;
}

export interface ReviewRequest {
  pr_url: string;
  full_context?: boolean;
}

export interface ReviewResponse {
  status: string;
  owner: string;
  repo: string;
  pr: number;
}

export interface ErrorResponse {
  error: string;
}
