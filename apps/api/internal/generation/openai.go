package generation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const openAIModel = "gpt-4.1-mini"

type GenerationProvider interface {
	GenerateSpec(ctx context.Context, projectName, prompt string) (Spec, error)
}

type SpecGenerator = GenerationProvider

type OpenAI struct {
	apiKey string
	client *http.Client
}

func NewOpenAI(apiKey string) *OpenAI {
	return &OpenAI{
		apiKey: strings.TrimSpace(apiKey),
		client: &http.Client{Timeout: 90 * time.Second},
	}
}

func (o *OpenAI) GenerateSpec(ctx context.Context, projectName, prompt string) (Spec, error) {
	if o.apiKey == "" {
		return Spec{}, fmt.Errorf("openai api key is not configured")
	}

	body, err := json.Marshal(chatRequest{
		Model: openAIModel,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: fmt.Sprintf("Project name: %s\n\nUser prompt:\n%s", projectName, prompt)},
		},
		ResponseFormat: responseFormat{
			Type: "json_schema",
			JSONSchema: jsonSchemaFormat{
				Name:   "forgeflow_ui",
				Strict: true,
				Schema: json.RawMessage(uiSchema),
			},
		},
	})
	if err != nil {
		return Spec{}, fmt.Errorf("encode generation request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Spec{}, fmt.Errorf("create generation request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return Spec{}, fmt.Errorf("reach generation model: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Spec{}, fmt.Errorf("read generation response: %w", err)
	}

	var parsed chatResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return Spec{}, fmt.Errorf("decode generation response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := "generation model request failed"
		if parsed.Error != nil && parsed.Error.Message != "" {
			message = publicModelMessage(parsed.Error.Message)
		}
		return Spec{}, &ModelError{Status: resp.StatusCode, Message: message}
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return Spec{}, fmt.Errorf("generation model returned an empty specification")
	}

	var spec Spec
	if err := json.Unmarshal([]byte(parsed.Choices[0].Message.Content), &spec); err != nil {
		return Spec{}, fmt.Errorf("decode ui specification: %w", err)
	}
	return spec, nil
}

type ModelError struct {
	Status  int
	Message string
}

func (e *ModelError) Error() string {
	return e.Message
}

func publicModelMessage(message string) string {
	lower := strings.ToLower(message)
	if strings.Contains(message, "sk-") || strings.Contains(lower, "api key") {
		return "OpenAI rejected the API key"
	}
	return clip(message, 300)
}

type chatRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	ResponseFormat responseFormat `json:"response_format"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type       string           `json:"type"`
	JSONSchema jsonSchemaFormat `json:"json_schema"`
}

type jsonSchemaFormat struct {
	Name   string          `json:"name"`
	Strict bool            `json:"strict"`
	Schema json.RawMessage `json:"schema"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

const systemPrompt = `You design interactive web UI prototypes for ForgeFlow.
Return a UI specification that matches the schema.
Use 1 to 3 pages. Every page route is unique, lowercase, and starts with /.
The first page route must be /.
Write specific copy for the user's product. Do not use lorem ipsum.
Navbar items use title for the label and route for an existing page route.
Pricing items use title for the plan name, price for the price, and description for features separated by " | ".
Testimonial items use title for the author and description for the quote.
FAQ items use title for the question and description for the answer.
Feature items use title and description.
Hero, cta, and pricing buttons use a route that already exists in pages, or an empty string.
Leave unused strings empty and unused item arrays empty.
Include a signup section when the user asks for an account, form, or signup flow.`

const uiSchema = `{
  "type": "object",
  "additionalProperties": false,
  "required": ["project", "pages"],
  "properties": {
    "project": {
      "type": "object",
      "additionalProperties": false,
      "required": ["name"],
      "properties": {
        "name": { "type": "string" }
      }
    },
    "pages": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["name", "route", "sections"],
        "properties": {
          "name": { "type": "string" },
          "route": { "type": "string" },
          "sections": {
            "type": "array",
            "items": {
              "type": "object",
              "additionalProperties": false,
              "required": ["type", "title", "subtitle", "primaryLabel", "primaryRoute", "secondaryLabel", "secondaryRoute", "items"],
              "properties": {
                "type": {
                  "type": "string",
                  "enum": ["navbar", "hero", "features", "pricing", "testimonials", "signup", "cta", "faq", "stats", "footer"]
                },
                "title": { "type": "string" },
                "subtitle": { "type": "string" },
                "primaryLabel": { "type": "string" },
                "primaryRoute": { "type": "string" },
                "secondaryLabel": { "type": "string" },
                "secondaryRoute": { "type": "string" },
                "items": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "additionalProperties": false,
                    "required": ["title", "description", "price", "route"],
                    "properties": {
                      "title": { "type": "string" },
                      "description": { "type": "string" },
                      "price": { "type": "string" },
                      "route": { "type": "string" }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`
