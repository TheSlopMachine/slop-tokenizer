package tokenizer

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/dlclark/regexp2"
)

//go:embed data/encodings/cl100k_base.json
var cl100kBaseJSON []byte

//go:embed data/encodings/o200k_base.json
var o200kBaseJSON []byte

//go:embed data/encodings/p50k_base.json
var p50kBaseJSON []byte

//go:embed data/encodings/claude.json
var claudeJSON []byte

// Encoding names
const (
	CL100K_BASE = "cl100k_base"
	O200K_BASE  = "o200k_base"
	P50K_BASE   = "p50k_base"
	CLAUDE      = "claude"
)

// encodingData maps encoding names to embedded JSON
var encodingData = map[string][]byte{
	CL100K_BASE: cl100kBaseJSON,
	O200K_BASE:  o200kBaseJSON,
	P50K_BASE:   p50kBaseJSON,
	CLAUDE:      claudeJSON,
}

// Encoding represents a tokenizer encoding
type Encoding struct {
	Name          string
	Pattern       *regexp2.Regexp
	SpecialTokens map[string]int
	StringEncoder map[string]int
	BinaryEncoder []BinaryToken
	Decoder       map[int][]byte

	// Internal optimization: first-byte index for binary search
	binaryIndex [256][]BinaryToken
}

// BinaryToken represents a non-UTF-8 token
type BinaryToken struct {
	Bytes []byte
	Rank  int
}

// encodingJSON is the JSON structure for parsing
type encodingJSON struct {
	Name          string            `json:"name"`
	PatStr        string            `json:"pat_str"`
	SpecialTokens map[string]int    `json:"special_tokens"`
	BPERanks      string            `json:"bpe_ranks"`
}

// LoadEncoding loads an encoding by name from embedded JSON
func LoadEncoding(name string) (*Encoding, error) {
	data, ok := encodingData[name]
	if !ok {
		return nil, fmt.Errorf("unknown encoding: %s", name)
	}

	var raw encodingJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse encoding: %w", err)
	}

	// Parse compressed BPE ranks
	ranks := parseBPERanks(raw.BPERanks)

	// Build encoding structure
	enc := &Encoding{
		Name:          name, // Use the name parameter, not from JSON
		SpecialTokens: raw.SpecialTokens,
		StringEncoder: make(map[string]int),
		BinaryEncoder: make([]BinaryToken, 0),
		Decoder:       make(map[int][]byte),
	}

	// Compile pattern regex using regexp2 for PCRE support (handles lookaheads)
	var err error
	enc.Pattern, err = regexp2.Compile(raw.PatStr, regexp2.None)
	if err != nil {
		return nil, fmt.Errorf("failed to compile pattern: %w", err)
	}

	// Separate string vs binary tokens
	for token, rank := range ranks {
		bytes, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			continue
		}

		if str, ok := tryBytesToString(bytes); ok {
			enc.StringEncoder[str] = rank
			enc.Decoder[rank] = []byte(str)
		} else {
			enc.BinaryEncoder = append(enc.BinaryEncoder, BinaryToken{
				Bytes: bytes,
				Rank:  rank,
			})
			enc.Decoder[rank] = bytes
		}
	}

	// Sort binary encoder for binary search
	sort.Slice(enc.BinaryEncoder, func(i, j int) bool {
		return compareBytesLess(enc.BinaryEncoder[i].Bytes, enc.BinaryEncoder[j].Bytes)
	})

	// Build first-byte index
	enc.buildBinaryIndex()

	return enc, nil
}

// parseBPERanks parses the compressed BPE ranks format
func parseBPERanks(compressed string) map[string]int {
	ranks := make(map[string]int)
	lines := []string{}
	
	// Split by newlines
	start := 0
	for i := 0; i < len(compressed); i++ {
		if compressed[i] == '\n' {
			if i > start {
				lines = append(lines, compressed[start:i])
			}
			start = i + 1
		}
	}
	if start < len(compressed) {
		lines = append(lines, compressed[start:])
	}

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		// Split by spaces
		parts := []string{}
		partStart := 0
		for i := 0; i <= len(line); i++ {
			if i == len(line) || line[i] == ' ' {
				if i > partStart {
					parts = append(parts, line[partStart:i])
				}
				partStart = i + 1
			}
		}

		if len(parts) < 2 {
			continue
		}

		// First part is empty, second is offset
		var offset int
		fmt.Sscanf(parts[1], "%d", &offset)

		// Remaining parts are tokens
		for i, token := range parts[2:] {
			ranks[token] = offset + i
		}
	}

	return ranks
}

// compareBytesLess compares two byte slices lexicographically
func compareBytesLess(a, b []byte) bool {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}

	return len(a) < len(b)
}

// buildBinaryIndex builds the first-byte index for faster binary search
func (e *Encoding) buildBinaryIndex() {
	for i := range e.binaryIndex {
		e.binaryIndex[i] = nil
	}

	for _, token := range e.BinaryEncoder {
		if len(token.Bytes) == 0 {
			continue
		}
		firstByte := token.Bytes[0]
		e.binaryIndex[firstByte] = append(e.binaryIndex[firstByte], token)
	}
}

// binarySearch searches for a byte slice in a sorted slice of BinaryTokens
func binarySearch(tokens []BinaryToken, key []byte) int {
	low := 0
	high := len(tokens) - 1

	for low <= high {
		mid := (low + high) / 2
		midKey := tokens[mid].Bytes

		cmp := 0
		minLen := len(midKey)
		if len(key) < minLen {
			minLen = len(key)
		}

		for i := 0; i < minLen; i++ {
			if midKey[i] != key[i] {
				cmp = int(midKey[i]) - int(key[i])
				break
			}
		}

		if cmp == 0 {
			cmp = len(midKey) - len(key)
		}

		if cmp == 0 {
			return mid
		}

		if cmp < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return -1
}
