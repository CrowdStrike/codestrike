package context

import "github.com/CrowdStrike/codestrike/internal/tokenizer"

type Budget struct {
	tokenizer            tokenizer.Tokenizer
	contextWindow        int
	maxInputRatio        float64
	reservedOutputTokens int
}

func NewBudget(tok tokenizer.Tokenizer, contextWindow int, maxInputRatio float64, reservedOutput int) *Budget {
	if maxInputRatio <= 0 || maxInputRatio > 1.0 {
		maxInputRatio = 0.75
	}
	return &Budget{
		tokenizer:            tok,
		contextWindow:        contextWindow,
		maxInputRatio:        maxInputRatio,
		reservedOutputTokens: reservedOutput,
	}
}

func (b *Budget) AvailableInputTokens() int {
	return int(float64(b.contextWindow)*b.maxInputRatio) - b.reservedOutputTokens
}

func (b *Budget) CountTokens(text string) int {
	return b.tokenizer.CountTokens(text)
}

type SectionAllocation struct {
	SystemPrompt     int
	PRMetadata       int
	ExistingComments int
	Memory           int
	FilePatches      int
}

func (b *Budget) Allocate(systemPromptTokens, prMetadataTokens, existingCommentsTokens, memoryTokens, filePatchesTokens int) SectionAllocation {
	available := b.AvailableInputTokens()

	alloc := SectionAllocation{}

	// Priority 1: system prompt (always included)
	alloc.SystemPrompt = min(systemPromptTokens, available)
	remaining := available - alloc.SystemPrompt

	// Priority 2: PR metadata (up to 10% of total available)
	metadataMax := int(float64(available) * 0.10)
	alloc.PRMetadata = min(prMetadataTokens, min(metadataMax, remaining))
	remaining -= alloc.PRMetadata

	// Priority 3: existing comments (up to 15% of total available)
	commentsMax := int(float64(available) * 0.15)
	alloc.ExistingComments = min(existingCommentsTokens, min(commentsMax, remaining))
	remaining -= alloc.ExistingComments

	// Priority 4: memory (up to 10% of total available)
	memoryMax := int(float64(available) * 0.10)
	alloc.Memory = min(memoryTokens, min(memoryMax, remaining))
	remaining -= alloc.Memory

	// Priority 5: file patches (all remaining)
	alloc.FilePatches = min(filePatchesTokens, remaining)

	return alloc
}

func (b *Budget) FitsInSingleCall(totalTokens int) bool {
	return totalTokens <= b.AvailableInputTokens()
}
