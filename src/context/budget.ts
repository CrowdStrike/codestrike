import type { Tokenizer } from '../tokenizer.js';

export interface SectionAllocation {
  systemPrompt: number;
  existingComments: number;
  memory: number;
  filePatches: number;
}

export class Budget {
  private tokenizer: Tokenizer;
  private contextWindow: number;
  private maxInputRatio: number;
  public reservedOutputTokens: number;

  constructor(tokenizer: Tokenizer, contextWindow: number, maxInputRatio: number, reservedOutput: number) {
    if (maxInputRatio <= 0 || maxInputRatio > 1.0) maxInputRatio = 0.75;
    this.tokenizer = tokenizer;
    this.contextWindow = contextWindow;
    this.maxInputRatio = maxInputRatio;
    this.reservedOutputTokens = reservedOutput;
  }

  availableInputTokens(): number {
    return Math.floor(this.contextWindow * this.maxInputRatio) - this.reservedOutputTokens;
  }

  countTokens(text: string): number {
    return this.tokenizer.countTokens(text);
  }

  allocate(
    systemPromptTokens: number,
    existingCommentsTokens: number,
    memoryTokens: number,
    _filePatchesTokens: number,
  ): SectionAllocation {
    const available = this.availableInputTokens();
    const alloc: SectionAllocation = { systemPrompt: 0, existingComments: 0, memory: 0, filePatches: 0 };

    alloc.systemPrompt = Math.min(systemPromptTokens, available);
    let remaining = available - alloc.systemPrompt;

    const commentsMax = Math.floor(available * 0.15);
    alloc.existingComments = Math.min(existingCommentsTokens, Math.min(commentsMax, remaining));
    remaining -= alloc.existingComments;

    const memoryMax = Math.floor(available * 0.1);
    alloc.memory = Math.min(memoryTokens, Math.min(memoryMax, remaining));
    remaining -= alloc.memory;

    alloc.filePatches = Math.min(_filePatchesTokens, remaining);

    return alloc;
  }

  fitsInSingleCall(totalTokens: number): boolean {
    return totalTokens <= this.availableInputTokens();
  }
}
