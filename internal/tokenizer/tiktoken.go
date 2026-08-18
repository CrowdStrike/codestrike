package tokenizer

import (
	"sync"

	tiktoken "github.com/pkoukk/tiktoken-go"
)

var supportedEncodings = map[string]string{
	"o200k_base":  "gpt-4o, gpt-4o-mini",
	"cl100k_base": "gpt-4, gpt-3.5-turbo, claude (approximate)",
	"p50k_base":   "text-davinci-003, codex",
}

func SupportedEncodings() map[string]string {
	return supportedEncodings
}

type TiktokenTokenizer struct {
	encoding *tiktoken.Tiktoken
}

var (
	cache   = make(map[string]*tiktoken.Tiktoken)
	cacheMu sync.Mutex
)

func New() *TiktokenTokenizer {
	return NewForModel("o200k_base")
}

func NewForModel(model string) *TiktokenTokenizer {
	if model == "" {
		model = "o200k_base"
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	if enc, ok := cache[model]; ok {
		return &TiktokenTokenizer{encoding: enc}
	}

	// Try as encoding name first (e.g., "cl100k_base")
	enc, err := tiktoken.GetEncoding(model)
	if err != nil {
		// Fall back to model name lookup (e.g., "gpt-4o")
		enc, err = tiktoken.EncodingForModel(model)
		if err != nil {
			enc, _ = tiktoken.GetEncoding("cl100k_base")
		}
	}

	cache[model] = enc
	return &TiktokenTokenizer{encoding: enc}
}

func (t *TiktokenTokenizer) CountTokens(text string) int {
	if t.encoding == nil {
		return len(text) / 4
	}
	return len(t.encoding.Encode(text, nil, nil))
}
