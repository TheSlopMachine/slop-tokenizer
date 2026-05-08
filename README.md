# AI Tokenizer (Go)

> **⚠️ DISCLAIMER**: This Go implementation was AI-generated and rewritten from the original TypeScript version. While comprehensive tests pass, use with appropriate caution and verification for production use.

Fast, accurate BPE tokenizer for AI models, rewritten in Go from the TypeScript implementation.

## Features

- **Fast**: Optimized BPE algorithm with LRU caching
- **Accurate**: Matches tiktoken output for all supported encodings
- **Zero Dependencies**: Uses only Go standard library (plus regexp2 for PCRE support)
- **Multiple Encodings**: Supports cl100k_base, o200k_base, p50k_base, and Claude encodings
- **100+ Models**: Pre-configured token counting for OpenAI, Anthropic, Google, and more
- **Embedded Data**: All encoding data embedded in binary via go:embed

## Installation

```bash
go get github.com/TheSlopMachine/slop-tokenizer
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/TheSlopMachine/slop-tokenizer"
)

func main() {
    // Create a tokenizer with o200k_base encoding
    tok, err := tokenizer.New(tokenizer.O200K_BASE)
    if err != nil {
        panic(err)
    }

    // Encode text to tokens
    tokens, err := tok.Encode("Hello, world!")
    if err != nil {
        panic(err)
    }
    fmt.Println("Tokens:", tokens)

    // Decode tokens back to text
    text, err := tok.Decode(tokens)
    if err != nil {
        panic(err)
    }
    fmt.Println("Text:", text)

    // Count tokens
    count := tok.Count("Hello, world!")
    fmt.Println("Token count:", count)
}
```

## Supported Encodings

- `tokenizer.CL100K_BASE` - GPT-4, GPT-3.5-turbo
- `tokenizer.O200K_BASE` - GPT-4o, GPT-5
- `tokenizer.P50K_BASE` - GPT-3, Codex
- `tokenizer.CLAUDE` - Claude models

## Working with Models

```go
// Get model configuration
model, err := tokenizer.GetModel("openai/gpt-5")
if err != nil {
    panic(err)
}

fmt.Println("Model:", model.Name)
fmt.Println("Context window:", model.ContextWindow)
fmt.Println("Encoding:", model.Encoding)

// Get model and load its encoding
model, enc, err := tokenizer.GetModelEncoding("openai/gpt-5")
if err != nil {
    panic(err)
}

tok, err := tokenizer.NewWithEncoding(enc)
if err != nil {
    panic(err)
}
```

## Advanced Usage

### Custom Cache Size

```go
tok, err := tokenizer.New(
    tokenizer.O200K_BASE,
    tokenizer.WithCacheSize(50000), // Default is 100000
)
```

### Special Tokens

```go
// Add custom special tokens
tok, err := tokenizer.New(
    tokenizer.O200K_BASE,
    tokenizer.WithSpecialTokens(map[string]int{
        "<|custom|>": 200000,
    }),
)

// Encode with special tokens allowed
opts := tokenizer.EncodeOptions{
    AllowedSpecial: []string{"<|endoftext|>"},
}
tokens, err := tok.EncodeWithOptions(text, opts)
```

## API Reference

### Tokenizer

```go
// Create tokenizer
func New(encodingName string, opts ...Option) (*Tokenizer, error)
func NewWithEncoding(enc *Encoding, opts ...Option) (*Tokenizer, error)

// Encode/decode
func (t *Tokenizer) Encode(text string) ([]int, error)
func (t *Tokenizer) EncodeWithOptions(text string, opts EncodeOptions) ([]int, error)
func (t *Tokenizer) Decode(tokens []int) (string, error)
func (t *Tokenizer) Count(text string) int
func (t *Tokenizer) EncodingName() string
```

### Models

```go
func GetModel(id string) (*Model, error)
func ListModels() []*Model
func GetModelEncoding(modelID string) (*Model, *Encoding, error)
```

### Encodings

```go
func LoadEncoding(name string) (*Encoding, error)
```

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestRoundTrip
```

## Implementation Details

### BPE Algorithm

The tokenizer uses the Byte Pair Encoding (BPE) algorithm with the following optimizations:

- **Dual storage**: String-based encoder for UTF-8 tokens (fast map lookup) and binary encoder for non-UTF-8 tokens (binary search)
- **First-byte indexing**: Reduces binary search domain size
- **LRU cache**: Caches BPE merge results for repeated text patterns
- **UTF-8 validation**: Fast validation without try/catch overhead

### Regex Pattern Matching

Uses `regexp2` library for PCRE-compatible regex patterns, including support for lookaheads like `(?!\S)` which are not supported by Go's standard `regexp` package.

### Data Embedding

All encoding data (2-8MB per encoding) is embedded in the binary using `go:embed`, eliminating the need for external files.

## Differences from TypeScript Version

1. **Regex Library**: Uses `regexp2` instead of standard `regexp` for PCRE support
2. **Decode Optimization**: Simpler implementation that collects all bytes then converts (TypeScript version optimizes by handling strings directly)
3. **API Style**: Idiomatic Go with error returns and options pattern
4. **No AI SDK Integration**: Core tokenizer only (no message/tool counting)

## Performance

The Go implementation matches or exceeds the TypeScript version's performance characteristics:

- Fast initialization with embedded data
- Efficient BPE merging with caching
- Minimal allocations in hot paths

## License

MIT

## Credits

Rewritten in Go from the TypeScript implementation by Kyle Carberry (https://github.com/coder/ai-tokenizer)
