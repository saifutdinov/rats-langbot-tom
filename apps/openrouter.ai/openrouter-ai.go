package openrouterai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type (
	client struct {
		apiKey string
		apiUrl string
		model  string
	}
	llmResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
)

func NewChatClient(apiKey, apiUrl, model string) *client {
	return &client{apiKey, apiUrl, model}
}

func (c *client) RequestChat(prompt string) (string, error) {
	payload := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", c.apiUrl, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("HTTP-Referer", "https://rats-langbot-tom-tg.net")
	req.Header.Set("X-Title", "Rats - Tom Langbottom ")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	llm := new(llmResp)
	if err = json.Unmarshal(response, &llm); err != nil {
		return "", err
	}

	// fmt.Println(llm.Choices[0].Message.Content)

	return llm.Choices[0].Message.Content, nil
}
