package tokenizer

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed data/models.json
var modelsJSON []byte

// Model represents an AI model configuration
type Model struct {
	ID            string        `json:"-"`
	Name          string        `json:"name"`
	Encoding      string        `json:"encoding"`
	ContextWindow int           `json:"contextWindow"`
	MaxTokens     int           `json:"maxTokens"`
	Tokens        TokenConfig   `json:"tokens"`
	Pricing       PricingConfig `json:"pricing"`
}

// TokenConfig contains token counting parameters for a model
type TokenConfig struct {
	ContentMultiplier float64 `json:"contentMultiplier"`
	BaseOverhead      int     `json:"baseOverhead"`
	PerMessage        int     `json:"perMessage"`
	ToolsExist        int     `json:"toolsExist"`
	PerTool           int     `json:"perTool"`
	PerDesc           int     `json:"perDesc"`
	PerFirstProp      int     `json:"perFirstProp"`
	PerAdditionalProp int     `json:"perAdditionalProp"`
	PerPropDesc       int     `json:"perPropDesc"`
	PerEnum           int     `json:"perEnum"`
	PerNestedObject   int     `json:"perNestedObject"`
	PerArrayOfObjects int     `json:"perArrayOfObjects"`
}

// PricingConfig contains pricing information for a model
type PricingConfig struct {
	Input           float64 `json:"input"`
	Output          float64 `json:"output"`
	InputCacheRead  float64 `json:"input_cache_read,omitempty"`
	InputCacheWrite float64 `json:"input_cache_write,omitempty"`
}

var modelsCache map[string]*Model

func init() {
	// Parse models on package init
	var models map[string]*Model
	if err := json.Unmarshal(modelsJSON, &models); err != nil {
		panic(fmt.Sprintf("failed to parse models.json: %v", err))
	}

	// Set ID field from map key
	for id, model := range models {
		model.ID = id
	}

	modelsCache = models
}

// GetModel retrieves a model by ID
func GetModel(id string) (*Model, error) {
	model, ok := modelsCache[id]
	if !ok {
		return nil, fmt.Errorf("model not found: %s", id)
	}
	return model, nil
}

// ListModels returns all available models
func ListModels() []*Model {
	models := make([]*Model, 0, len(modelsCache))
	for _, model := range modelsCache {
		models = append(models, model)
	}
	return models
}

// GetModelEncoding is a convenience function to get a model and load its encoding
func GetModelEncoding(modelID string) (*Model, *Encoding, error) {
	model, err := GetModel(modelID)
	if err != nil {
		return nil, nil, err
	}

	enc, err := LoadEncoding(model.Encoding)
	if err != nil {
		return nil, nil, err
	}

	return model, enc, nil
}
