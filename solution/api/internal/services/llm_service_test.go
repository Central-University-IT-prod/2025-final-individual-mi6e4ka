package services

import (
	"testing"

	"git.mi6e4ka.dev/prod-2025/internal/config"
)

var llmConfig = config.Config{
	Ollama: config.OllamaConfig{BaseURL: "http://localhost:11434", Model: "qwen2.5:0.5b"},
}

func TestNewLLMService(t *testing.T) {
	_, err := NewLLMService(&llmConfig)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestGenerateDescription(t *testing.T) {
	llm, err := NewLLMService(&llmConfig)
	if err != nil {
		t.Error(err)
		return
	}
	testCases := []struct {
		adTitle     string
		companyName string
	}{
		{"Курсы Java", "Шахов Production"},
		{"Устройся в Avito Golang разработчиком!", "Балюкпобеда.рф"},
		{"Вступай в клуб хейторов Flutter", "ИП Михаил Кузенецов"},
	}
	for _, tc := range testCases {
		t.Logf("Заголовок: %s, Компания: %s", tc.adTitle, tc.companyName)
		res, err := llm.GenerateDescription(tc.adTitle, tc.companyName)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(res)
	}
}
