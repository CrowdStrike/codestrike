package tokenizer

type Tokenizer interface {
	CountTokens(text string) int
}

type ModelLimits struct {
	ContextWindow   int
	MaxOutputTokens int
}

var defaultModelLimits = map[string]ModelLimits{
	"gpt-4o":                    {ContextWindow: 128000, MaxOutputTokens: 16384},
	"gpt-4o-mini":               {ContextWindow: 128000, MaxOutputTokens: 16384},
	"gpt-4-turbo":               {ContextWindow: 128000, MaxOutputTokens: 4096},
	"claude-sonnet-4-20250514":  {ContextWindow: 200000, MaxOutputTokens: 8192},
	"claude-haiku-4-5-20251001": {ContextWindow: 200000, MaxOutputTokens: 8192},
	"anthropic.claude-3-5-sonnet-20241022-v2:0": {ContextWindow: 200000, MaxOutputTokens: 8192},
	"anthropic.claude-3-haiku-20240307-v1:0":    {ContextWindow: 200000, MaxOutputTokens: 4096},
	"llama3":                                    {ContextWindow: 8192, MaxOutputTokens: 2048},
	"llama3:70b":                                {ContextWindow: 8192, MaxOutputTokens: 2048},
	"codellama":                                 {ContextWindow: 16384, MaxOutputTokens: 4096},
}

func GetModelLimits(modelID string) ModelLimits {
	if limits, ok := defaultModelLimits[modelID]; ok {
		return limits
	}
	return ModelLimits{ContextWindow: 8192, MaxOutputTokens: 4096}
}
