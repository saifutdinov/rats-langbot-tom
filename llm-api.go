package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func generateQuestion() string {
	prompt := `
Придумай простой вопрос на английском языке с 1 правильным и 3 неправильными вариантами ответа. Отдай результат строго в формате JSON:
{
  "question": "Which animal barks?",
  "options": ["Dog", "Cat", "Mouse", "Bird"],
  "answer": "Dog"
}`

	payload := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", config["LLM_API_URL"], bytes.NewBuffer(body))
	if err != nil {
		return "Ошибка создания запроса к LLM"
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config["LLM_API_KEY"])

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "Ошибка обращения к LLM"
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	// Попытка распарсить ответ
	var gptResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &gptResp); err != nil || len(gptResp.Choices) == 0 {
		return "Ошибка парсинга ответа LLM"
	}

	// Теперь вытаскиваем JSON из content
	content := gptResp.Choices[0].Message.Content

	// В случае, если content — это JSON, пытаемся разобрать
	var quiz struct {
		Question string   `json:"question"`
		Options  []string `json:"options"`
		Answer   string   `json:"answer"`
	}
	if err := json.Unmarshal([]byte(content), &quiz); err != nil {
		return "Ошибка парсинга JSON из LLM-ответа:\n" + content
	}

	// Формируем Markdown
	msg := "*Вопрос:*\n" + quiz.Question + "\n\n"
	for _, opt := range quiz.Options {
		mark := "⬜️"
		if opt == quiz.Answer {
			mark = "✅"
		}
		msg += mark + " " + opt + "\n"
	}

	return msg
}
