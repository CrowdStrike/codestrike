package llm

import "fmt"

type LLMFamily string

const (
	FamilyAnthropic      LLMFamily = "anthropic"
	FamilyOpenAIPlatform LLMFamily = "openai_platform"
	FamilyOllama         LLMFamily = "ollama"
)

func ParseFamily(s string) (LLMFamily, error) {
	switch LLMFamily(s) {
	case FamilyAnthropic, FamilyOpenAIPlatform, FamilyOllama:
		return LLMFamily(s), nil
	default:
		return "", fmt.Errorf("unsupported model family: %s (valid: anthropic, openai_platform, ollama)", s)
	}
}

type LLMClientRegistry struct {
	clients map[LLMFamily]map[string]LLMClient
}

func NewLLMClientRegistry(clients map[LLMFamily]map[string]LLMClient) *LLMClientRegistry {
	if clients == nil {
		clients = make(map[LLMFamily]map[string]LLMClient)
	}
	return &LLMClientRegistry{
		clients: clients,
	}
}

func (r *LLMClientRegistry) Get(family LLMFamily, modelID string) (LLMClient, error) {
	familyClients, exists := r.clients[family]
	if !exists {
		return nil, fmt.Errorf("no clients registered for family: %s", family)
	}

	client, exists := familyClients[modelID]
	if !exists {
		return nil, fmt.Errorf("client not found for family %s and model %s", family, modelID)
	}

	return client, nil
}
