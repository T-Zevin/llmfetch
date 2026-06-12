package registry

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/xzw/llmfetch/internal/model"
)

//go:embed models.json
var bundledModels []byte

func LoadBundledModels() ([]model.Model, error) {
	var models []model.Model
	if err := json.Unmarshal(bundledModels, &models); err != nil {
		return nil, fmt.Errorf("decode bundled models: %w", err)
	}
	return models, nil
}
