package tokenizer

import "testing"

func TestGetModel(t *testing.T) {
	tests := []struct {
		name      string
		modelID   string
		wantError bool
	}{
		{"valid openai model", "openai/gpt-5", false},
		{"valid anthropic model", "anthropic/claude-sonnet-4.5", false},
		{"valid google model", "google/gemini-2.5-pro", false},
		{"invalid model", "invalid/model", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := GetModel(tt.modelID)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if model == nil {
					t.Fatal("expected model, got nil")
				}
				if model.ID != tt.modelID {
					t.Errorf("got ID %q, want %q", model.ID, tt.modelID)
				}
				if model.Name == "" {
					t.Error("expected non-empty name")
				}
				if model.Encoding == "" {
					t.Error("expected non-empty encoding")
				}
			}
		})
	}
}

func TestListModels(t *testing.T) {
	models := ListModels()
	if len(models) == 0 {
		t.Error("expected non-empty models list")
	}

	// Check that all models have required fields
	for _, model := range models {
		if model.ID == "" {
			t.Error("model missing ID")
		}
		if model.Name == "" {
			t.Error("model missing name")
		}
		if model.Encoding == "" {
			t.Error("model missing encoding")
		}
	}
}

func TestGetModelEncoding(t *testing.T) {
	model, enc, err := GetModelEncoding("openai/gpt-5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if model == nil {
		t.Fatal("expected model, got nil")
	}

	if enc == nil {
		t.Fatal("expected encoding, got nil")
	}

	if enc.Name != model.Encoding {
		t.Errorf("encoding name %q doesn't match model encoding %q", enc.Name, model.Encoding)
	}
}

func TestModelTokenConfig(t *testing.T) {
	model, err := GetModel("openai/gpt-5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that token config has reasonable values
	if model.Tokens.BaseOverhead < 0 {
		t.Error("baseOverhead should be non-negative")
	}
	if model.Tokens.PerMessage < 0 {
		t.Error("perMessage should be non-negative")
	}
	if model.Tokens.ContentMultiplier <= 0 {
		t.Error("contentMultiplier should be positive")
	}
}

func TestModelPricing(t *testing.T) {
	model, err := GetModel("openai/gpt-5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that pricing has reasonable values
	if model.Pricing.Input < 0 {
		t.Error("input pricing should be non-negative")
	}
	if model.Pricing.Output < 0 {
		t.Error("output pricing should be non-negative")
	}
}
