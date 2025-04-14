package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"git.mi6e4ka.dev/prod-2025/internal/config"
)

type LLMService struct {
	BaseURL string
	Model   string
}

func NewLLMService(config *config.Config) (*LLMService, error) {
	if config == nil {
		return nil, errors.New("invalid configuration")
	}
	// check is ollama running
	resp, err := http.Get(config.Ollama.BaseURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("ollama not running")
	}
	return &LLMService{BaseURL: config.Ollama.BaseURL, Model: config.Ollama.Model}, nil
}

type OllamaApi struct {
	Model   string                 `json:"model"`
	System  string                 `json:"system,omitempty"`
	Prompt  string                 `json:"prompt,omitempty"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
	// Format  OllamaApiFormat        `json:"format,omitempty"`
}
type OllamaApiFormat struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required"`
}
type OllamaApiResponse struct {
	Response string `json:"response"`
}

const systemPrompt = "Ты – эксперт в рекламе. Тебе присылают заголовки объявлений и название компании, и твоя задача – сгенерировать краткое (2-3 предложения) продающее описание на том же языке, на котором написан заголовок. Если заголовок на английском – пиши на английском, и так далее. Напоминаю, описание должно быть продающим и вызывать эмоции и пользователя. Если возможно сделать - упоминай название компании в ответе. В описание ни в коем случае не может быть переносов"

func (s *LLMService) GenerateDescription(adTitle string, companyName string) (string, error) {
	body := OllamaApi{
		Model:   s.Model,
		System:  systemPrompt,
		Prompt:  fmt.Sprintf("Заголовок: \"%s\"\nКомпания: \"%s\"", adTitle, companyName),
		Stream:  false,
		Options: map[string]interface{}{"temperature": 0.4, "num_predict": 200},
	}
	bodyJson, _ := json.Marshal(body)
	res, err := http.Post(s.BaseURL+"/api/generate", "application/json", bytes.NewBuffer(bodyJson))
	if err != nil || res.StatusCode != http.StatusOK {
		if err == nil {
			err = fmt.Errorf("ollama api error: %s", http.StatusText(res.StatusCode))
		}
		return "", err
	}
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)
	var jsonBody OllamaApiResponse
	json.Unmarshal(resBody, &jsonBody)
	return jsonBody.Response, nil
}
