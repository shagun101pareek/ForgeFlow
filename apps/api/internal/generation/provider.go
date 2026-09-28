package generation

import (
	"fmt"
	"strings"
)

func NewProvider(name, openAIKey string) (GenerationProvider, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "mock":
		return Mock{}, nil
	case "openai":
		return NewOpenAI(openAIKey), nil
	default:
		return nil, fmt.Errorf("unknown generation provider %q", name)
	}
}
