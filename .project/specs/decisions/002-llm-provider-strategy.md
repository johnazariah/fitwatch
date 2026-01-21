# ADR-002: LLM Provider Strategy

## Status
Proposed

## Context
The application needs LLM capabilities for:
- Generating workout summaries
- Analyzing training patterns
- Providing coaching suggestions
- Conversational interface for workout data

We need to decide which LLM provider(s) to support.

## Options Considered

### Option A: OpenAI Only
Use OpenAI's GPT-4 API exclusively.

**Pros:**
- Best-in-class quality
- Simple to implement
- Good documentation

**Cons:**
- Privacy concerns (workout data sent to OpenAI)
- Costs money per request
- Single vendor dependency

### Option B: Anthropic Claude Only
Use Claude API exclusively.

**Pros:**
- Strong reasoning capabilities
- Good at structured analysis
- 100K context window useful for multi-workout analysis

**Cons:**
- Same privacy/cost concerns as OpenAI
- Slightly less ecosystem tooling

### Option C: Local LLM Only (Ollama/llama.cpp)
Run models locally (Llama 3, Mistral, etc.).

**Pros:**
- Complete privacy - data never leaves machine
- No per-request costs
- Works offline

**Cons:**
- Requires significant hardware (8GB+ VRAM ideal)
- Lower quality than frontier models
- More complex setup for users

### Option D: Pluggable Provider (Recommended)
Abstract LLM calls behind an interface, support multiple providers.

**Pros:**
- User choice based on their priorities
- Privacy-conscious users can use local
- Quality-focused users can use cloud APIs
- Future-proof as new models emerge

**Cons:**
- More code to maintain
- Need to test with multiple providers
- Prompts may need tuning per model

## Decision
**Option D: Pluggable provider with Ollama as default**

Implement a provider abstraction layer supporting:
1. **Ollama** (default) - local, private, free
2. **OpenAI** - optional, for users who want best quality
3. **Anthropic** - optional, alternative cloud provider
4. **OpenAI-compatible APIs** - for users running other backends

## Rationale
1. Aligns with self-hosted philosophy (ADR-001)
2. Local LLMs are now good enough for summarization tasks
3. Power users can upgrade to cloud APIs if desired
4. Abstract interface is relatively simple to build

## Implementation Notes

```python
# Abstract interface
class LLMProvider(Protocol):
    async def complete(self, prompt: str, system: str = None) -> str: ...
    async def chat(self, messages: list[Message]) -> str: ...

# Implementations
class OllamaProvider(LLMProvider): ...
class OpenAIProvider(LLMProvider): ...
class AnthropicProvider(LLMProvider): ...

# Configuration
llm:
  provider: ollama  # or openai, anthropic
  model: llama3.2   # provider-specific model name
  base_url: http://localhost:11434  # for ollama/custom
  api_key: ${OPENAI_API_KEY}  # for cloud providers
```

## Prompt Engineering Notes
- Keep prompts model-agnostic where possible
- Test critical prompts (workout summary) on all supported models
- May need model-specific system prompts for best results
- Provide example outputs in prompts for consistency

## Hardware Recommendations (for local)
- Minimum: 8GB RAM, CPU-only (slow but works)
- Recommended: 16GB RAM + 8GB VRAM GPU
- Ideal: 32GB RAM + 16GB+ VRAM

## Consequences
- Documentation must cover multiple provider setups
- CI/CD should test with at least Ollama
- Quality may vary by provider - set user expectations

## Review Date
Revisit as local model quality improves (every 6 months).
