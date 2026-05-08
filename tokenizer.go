package tokenizer

import (
	"fmt"
	"strings"

	"github.com/dlclark/regexp2"
)

// Tokenizer performs BPE tokenization
type Tokenizer struct {
	encoding              *Encoding
	cache                 *lruCache
	specialTokensRegex    *regexp2.Regexp
	inverseSpecialTokens  map[int][]byte
	hasSpecialTokens      bool
	specialTokenKeys      []string
	extendedSpecialTokens map[string]int
}

// Option configures a Tokenizer
type Option func(*Tokenizer)

// WithCacheSize sets the LRU cache size (default: 100000)
func WithCacheSize(size int) Option {
	return func(t *Tokenizer) {
		t.cache = newLRUCache(size)
	}
}

// WithSpecialTokens adds custom special tokens
func WithSpecialTokens(tokens map[string]int) Option {
	return func(t *Tokenizer) {
		t.extendedSpecialTokens = tokens
	}
}

// New creates a tokenizer for the given encoding name
func New(encodingName string, opts ...Option) (*Tokenizer, error) {
	enc, err := LoadEncoding(encodingName)
	if err != nil {
		return nil, err
	}
	return NewWithEncoding(enc, opts...)
}

// NewWithEncoding creates a tokenizer with a custom encoding
func NewWithEncoding(enc *Encoding, opts ...Option) (*Tokenizer, error) {
	t := &Tokenizer{
		encoding: enc,
		cache:    newLRUCache(100000), // Default cache size
	}

	// Apply options
	for _, opt := range opts {
		opt(t)
	}

	// Merge special tokens
	allSpecialTokens := make(map[string]int)
	for k, v := range enc.SpecialTokens {
		allSpecialTokens[k] = v
	}
	for k, v := range t.extendedSpecialTokens {
		allSpecialTokens[k] = v
	}

	// Build special tokens data
	t.specialTokenKeys = make([]string, 0, len(allSpecialTokens))
	for k := range allSpecialTokens {
		t.specialTokenKeys = append(t.specialTokenKeys, k)
	}
	t.hasSpecialTokens = len(t.specialTokenKeys) > 0

	// Pre-compile special tokens regex
	if t.hasSpecialTokens {
		escaped := make([]string, len(t.specialTokenKeys))
		for i, token := range t.specialTokenKeys {
			escaped[i] = regexp2.Escape(token)
		}
		pattern := strings.Join(escaped, "|")
		t.specialTokensRegex, _ = regexp2.Compile(pattern, regexp2.None)
	}

	// Pre-encode special tokens to bytes
	t.inverseSpecialTokens = make(map[int][]byte)
	for text, rank := range allSpecialTokens {
		t.inverseSpecialTokens[rank] = []byte(text)
	}

	return t, nil
}

// EncodeOptions for special token handling
type EncodeOptions struct {
	AllowedSpecial   []string // nil = none, empty slice = none, special value "all" = all
	DisallowedSpecial []string // nil = all, empty slice = none
}

// Encode encodes text to tokens with default options (no special tokens allowed)
func (t *Tokenizer) Encode(text string) ([]int, error) {
	return t.EncodeWithOptions(text, EncodeOptions{})
}

// EncodeWithOptions encodes text to tokens with custom special token handling
func (t *Tokenizer) EncodeWithOptions(text string, opts EncodeOptions) ([]int, error) {
	// Fast path: no special tokens to handle
	if !t.hasSpecialTokens || isAllAllowed(opts.AllowedSpecial) {
		return t.encodeOrdinary(text), nil
	}

	// Build allowed and disallowed sets
	allowedSet := make(map[string]bool)
	if isAllAllowed(opts.AllowedSpecial) {
		for _, k := range t.specialTokenKeys {
			allowedSet[k] = true
		}
	} else {
		for _, k := range opts.AllowedSpecial {
			allowedSet[k] = true
		}
	}

	disallowedSet := make(map[string]bool)
	if opts.DisallowedSpecial == nil {
		// Default: disallow all special tokens not in allowed set
		for _, k := range t.specialTokenKeys {
			if !allowedSet[k] {
				disallowedSet[k] = true
			}
		}
	} else {
		for _, k := range opts.DisallowedSpecial {
			disallowedSet[k] = true
		}
	}

	// Check for disallowed special tokens
	if len(disallowedSet) > 0 {
		disallowedKeys := make([]string, 0, len(disallowedSet))
		for k := range disallowedSet {
			disallowedKeys = append(disallowedKeys, k)
		}
		escaped := make([]string, len(disallowedKeys))
		for i, k := range disallowedKeys {
			escaped[i] = regexp2.Escape(k)
		}
		disallowedRegex, _ := regexp2.Compile(strings.Join(escaped, "|"), regexp2.None)
		if match, _ := disallowedRegex.FindStringMatch(text); match != nil {
			return nil, fmt.Errorf("text contains disallowed special token: %s", match.String())
		}
	}

	// Process text with special tokens
	result := []int{}
	start := 0

	for start < len(text) {
		// Find next allowed special token
		nextSpecialText := ""
		nextSpecialIdx := -1

		if t.specialTokensRegex != nil {
			match, _ := t.specialTokensRegex.FindStringMatch(text[start:])
			for match != nil {
				matchText := match.String()
				if allowedSet[matchText] {
					nextSpecialText = matchText
					nextSpecialIdx = start + match.Index
					break
				}
				match, _ = t.specialTokensRegex.FindNextMatch(match)
			}
		}

		end := len(text)
		if nextSpecialIdx != -1 {
			end = nextSpecialIdx
		}

		// Encode ordinary text before special token
		if end > start {
			ordinary := t.encodeOrdinary(text[start:end])
			result = append(result, ordinary...)
		}

		if nextSpecialIdx == -1 {
			break
		}

		// Add special token
		specialTokens := t.encoding.SpecialTokens
		for k, v := range t.extendedSpecialTokens {
			specialTokens[k] = v
		}
		result = append(result, specialTokens[nextSpecialText])
		start = nextSpecialIdx + len(nextSpecialText)
	}

	return result, nil
}

// encodeOrdinary encodes ordinary text (no special tokens)
func (t *Tokenizer) encodeOrdinary(text string) []int {
	if text == "" {
		return []int{}
	}

	// Quick single-token check for very short text
	if len(text) < 10 {
		if rank, ok := t.encoding.StringEncoder[text]; ok {
			return []int{rank}
		}
	}

	result := []int{}
	
	// Use regexp2 FindStringMatch for iteration
	match, _ := t.encoding.Pattern.FindStringMatch(text)
	for match != nil {
		piece := match.String()

		// Direct rank check (most common case)
		if rank, ok := t.encoding.StringEncoder[piece]; ok {
			result = append(result, rank)
			match, _ = t.encoding.Pattern.FindNextMatch(match)
			continue
		}

		// Cache check
		if cached, ok := t.cache.get(piece); ok {
			result = append(result, cached...)
			match, _ = t.encoding.Pattern.FindNextMatch(match)
			continue
		}

		// BPE encode
		bytes := []byte(piece)
		tokens := bytePairMerge(bytes, t.encoding.StringEncoder, t.encoding.binaryIndex)

		// Cache and add tokens
		t.cache.put(piece, tokens)
		result = append(result, tokens...)
		
		match, _ = t.encoding.Pattern.FindNextMatch(match)
	}

	return result
}

// Decode decodes tokens to text
func (t *Tokenizer) Decode(tokens []int) (string, error) {
	if len(tokens) == 0 {
		return "", nil
	}

	// Collect all bytes first
	var allBytes []byte
	
	for _, token := range tokens {
		value, ok := t.encoding.Decoder[token]
		if !ok {
			// Check inverse special tokens
			if specialBytes, ok := t.inverseSpecialTokens[token]; ok {
				value = specialBytes
			} else {
				continue // Skip unknown tokens
			}
		}
		
		allBytes = append(allBytes, value...)
	}

	// Convert all bytes to string at once
	return string(allBytes), nil
}

// Count counts tokens in text
func (t *Tokenizer) Count(text string) int {
	tokens, _ := t.Encode(text)
	return len(tokens)
}

// EncodingName returns the encoding name
func (t *Tokenizer) EncodingName() string {
	return t.encoding.Name
}

// isAllAllowed checks if the allowed special tokens list means "all"
func isAllAllowed(allowed []string) bool {
	if len(allowed) == 1 && allowed[0] == "all" {
		return true
	}
	return false
}
