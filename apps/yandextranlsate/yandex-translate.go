package yandextranlsate

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type (
	engine struct {
		apiKey   string
		folderId string
	}
	TranslateResult struct {
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}

	detectLangResult struct {
		LanguageCode string `json:"languageCode"`
	}
)

func NewYandexEngine(apiKey, folderId string) *engine {
	return &engine{apiKey, folderId}
}

func (e *engine) Translate(text, langFrom, langTo string) (translatedText string, err error) {
	endpoint := "https://translate.api.cloud.yandex.net/translate/v2/translate"

	if langFrom == "auto" {
		langFrom, err = e.DetectLang(text)
		if err != nil {
			return "", err
		}
	}

	payload := map[string]any{
		"sourceLanguageCode": langFrom,
		"targetLanguageCode": langTo,
		"texts":              []string{text},
		"folderId":           e.folderId,
	}

	response, err := e.requestYandex(endpoint, payload)
	if err != nil {
		return "", err
	}

	result := new(TranslateResult)
	if err := json.Unmarshal(response, &result); err != nil {
		return "", err
	}
	if len(result.Translations) > 0 {
		translatedText = result.Translations[0].Text
	}

	return
}

func (e *engine) DetectLang(text string) (string, error) {
	endpoint := "https://translate.api.cloud.yandex.net/translate/v2/detect"
	payload := map[string]any{
		"text": text,
		"languageCodeHints": []string{
			"ru", "en", "es", "fr", "it",
		},
		"folderId": e.folderId,
	}
	response, err := e.requestYandex(endpoint, payload)
	if err != nil {
		return "", err
	}

	result := new(detectLangResult)
	if err := json.Unmarshal(response, &result); err != nil {
		return "", err
	}
	return result.LanguageCode, nil
}

func (e *engine) requestYandex(endpoint string, payload map[string]any) ([]byte, error) {
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+getValidIAMToken(e.apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
