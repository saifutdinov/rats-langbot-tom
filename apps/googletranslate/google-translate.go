package googletranslate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type (
	engine struct {
		apiKey    string
		projectId string
	}

	translateResult struct {
		Data struct {
			Translations []struct {
				TranslatedText string `json:"translatedText"`
			} `json:"translations"`
		} `json:"data"`
	}
)

func NewGoogleEngine(apiKey, projectId string) *engine {
	return &engine{apiKey, projectId}
}

func (e *engine) Translate(text, langFrom, langTo string) (string, error) {
	endpoint := "https://translation.googleapis.com/language/translate/v2"
	payload := map[string]any{
		"q":      text,
		"target": langTo,
		"format": "text",
	}
	if langFrom != "auto" {
		payload["source"] = langFrom
	}

	response, err := e.requestGoogle(endpoint, payload)
	if err != nil {
		return "", err
	}
	result := new(translateResult)
	if err := json.Unmarshal(response, &result); err != nil {
		return "", err
	}

	fmt.Println(string(response))
	if len(result.Data.Translations) == 0 {
		return "Перевод не найден!", nil
	}

	return htmlUnescape(result.Data.Translations[0].TranslatedText), nil
}

func (e *engine) requestGoogle(endpoint string, payload map[string]any) ([]byte, error) {
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", endpoint+"?key="+e.apiKey, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("x-goog-user-project", e.projectId)
	// req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Google возвращает HTML entities, например &quot; вместо " — убираем это
func htmlUnescape(s string) string {
	replacer := strings.NewReplacer(
		"&quot;", "\"",
		"&#39;", "'",
		"&lt;", "<",
		"&gt;", ">",
		"&amp;", "&",
	)
	return replacer.Replace(s)
}
