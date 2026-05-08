package tokenizer

import (
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name     string
		encoding string
		text     string
	}{
		{"empty", O200K_BASE, ""},
		{"simple word", O200K_BASE, "hello"},
		{"simple sentence", O200K_BASE, "Hello, world!"},
		{"unicode chinese", O200K_BASE, "你好"},
		{"unicode emoji", O200K_BASE, "🌍"},
		{"whitespace", O200K_BASE, "   "},
		{"numbers", O200K_BASE, "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok, err := New(tt.encoding)
			if err != nil {
				t.Fatalf("failed to create tokenizer: %v", err)
			}

			tokens, err := tok.Encode(tt.text)
			if err != nil {
				t.Fatalf("encode failed: %v", err)
			}

			// Test that we got some tokens (except for empty string)
			if tt.text != "" && len(tokens) == 0 {
				t.Errorf("expected non-empty tokens for %q", tt.text)
			}

			// Test round-trip
			decoded, err := tok.Decode(tokens)
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if decoded != tt.text {
				t.Errorf("round-trip failed: got %q, want %q", decoded, tt.text)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	texts := []string{
		"Hello, world!",
		"你好世界",
		"🌍🚀",
		"Mixed 日本語 text",
		"The quick brown fox jumps over the lazy dog.",
		"function test() { return 42; }",
		"",
		"a",
		"   \n\t  ",
	}

	encodings := []string{O200K_BASE, CL100K_BASE, P50K_BASE, CLAUDE}

	for _, encoding := range encodings {
		t.Run(encoding, func(t *testing.T) {
			tok, err := New(encoding)
			if err != nil {
				t.Fatalf("failed to create tokenizer: %v", err)
			}

			for _, text := range texts {
				t.Run(text, func(t *testing.T) {
					tokens, err := tok.Encode(text)
					if err != nil {
						t.Fatalf("encode failed: %v", err)
					}

					decoded, err := tok.Decode(tokens)
					if err != nil {
						t.Fatalf("decode failed: %v", err)
					}

					if decoded != text {
						t.Errorf("round-trip failed: got %q, want %q", decoded, text)
					}
				})
			}
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name     string
		encoding string
		text     string
	}{
		{"empty", O200K_BASE, ""},
		{"simple", O200K_BASE, "hello"},
		{"sentence", O200K_BASE, "Hello, world!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok, err := New(tt.encoding)
			if err != nil {
				t.Fatalf("failed to create tokenizer: %v", err)
			}

			got := tok.Count(tt.text)
			
			// Verify count matches encode length
			tokens, _ := tok.Encode(tt.text)
			if got != len(tokens) {
				t.Errorf("count %d doesn't match encode length %d", got, len(tokens))
			}
		})
	}
}

func TestEncodingName(t *testing.T) {
	tok, err := New(O200K_BASE)
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	if tok.EncodingName() != "o200k_base" {
		t.Errorf("got %q, want %q", tok.EncodingName(), "o200k_base")
	}
}

func TestAllEncodings(t *testing.T) {
	encodings := []string{CL100K_BASE, O200K_BASE, P50K_BASE, CLAUDE}

	for _, encoding := range encodings {
		t.Run(encoding, func(t *testing.T) {
			tok, err := New(encoding)
			if err != nil {
				t.Fatalf("failed to create tokenizer: %v", err)
			}

			// Test basic functionality
			text := "Hello, world!"
			tokens, err := tok.Encode(text)
			if err != nil {
				t.Fatalf("encode failed: %v", err)
			}

			if len(tokens) == 0 {
				t.Error("expected non-empty tokens")
			}

			decoded, err := tok.Decode(tokens)
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if decoded != text {
				t.Errorf("round-trip failed: got %q, want %q", decoded, text)
			}
		})
	}
}

func TestCacheEffectiveness(t *testing.T) {
	tok, err := New(O200K_BASE)
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	text := "repeated text for cache test"

	// Encode multiple times - should use cache
	first, _ := tok.Encode(text)
	second, _ := tok.Encode(text)
	third, _ := tok.Encode(text)

	if !slicesEqual(first, second) || !slicesEqual(second, third) {
		t.Error("cache should return consistent results")
	}
}

func TestLongText(t *testing.T) {
	tok, err := New(O200K_BASE)
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	// Generate long text
	longText := ""
	for i := 0; i < 1000; i++ {
		longText += "The quick brown fox jumps over the lazy dog. "
	}

	tokens, err := tok.Encode(longText)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	if len(tokens) == 0 {
		t.Error("expected non-empty tokens for long text")
	}

	decoded, err := tok.Decode(tokens)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded != longText {
		t.Error("round-trip failed for long text")
	}
}

func TestUnicodeEdgeCases(t *testing.T) {
	tok, err := New(O200K_BASE)
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	tests := []string{
		"Hello 世界! 🌍 Привет мир!",
		"مرحبا بالعالم",
		"こんにちは世界",
		"Mixed 日本語 and English",
		"Emoji: 😀😃😄😁",
	}

	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			tokens, err := tok.Encode(text)
			if err != nil {
				t.Fatalf("encode failed: %v", err)
			}

			decoded, err := tok.Decode(tokens)
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if decoded != text {
				t.Errorf("round-trip failed: got %q, want %q", decoded, text)
			}
		})
	}
}

func TestSpecialCharacters(t *testing.T) {
	tok, err := New(O200K_BASE)
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	tests := []string{
		"!@#$%^&*()_+-=[]{}|;':\",./<>?",
		"line1\nline2\r\nline3\rline4",
		"\t\t\ttabs",
		"null\x00byte",
	}

	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			tokens, err := tok.Encode(text)
			if err != nil {
				t.Fatalf("encode failed: %v", err)
			}

			decoded, err := tok.Decode(tokens)
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if decoded != text {
				t.Errorf("round-trip failed: got %q, want %q", decoded, text)
			}
		})
	}
}

func TestWithCacheSize(t *testing.T) {
	tok, err := New(O200K_BASE, WithCacheSize(10))
	if err != nil {
		t.Fatalf("failed to create tokenizer: %v", err)
	}

	text := "test"
	tokens, err := tok.Encode(text)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	if len(tokens) == 0 {
		t.Error("expected non-empty tokens")
	}
}

// Helper function to compare slices
func slicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
