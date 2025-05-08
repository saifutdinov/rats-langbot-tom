package yandextranslate

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"saifutdinov/rats-langbot-tom/config"
)

func YandexTranslateCustom(text, fromLang, toLang string) string {
	payload := map[string]interface{}{
		"sourceLanguageCode": fromLang,
		"targetLanguageCode": toLang,
		"texts":              []string{text},
		"folderId":           config.GetValue("YANDEX_FOLDER_ID"),
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://translate.api.cloud.yandex.net/translate/v2/translate", bytes.NewBuffer(body))
	if err != nil {
		return "Ошибка при создании запроса к Яндекс API"
	}

	req.Header.Set("Content-Type", "application/json")

	iamToken := getValidIAMToken()

	req.Header.Set("Authorization", "Bearer "+iamToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "Ошибка при отправке запроса в Яндекс"
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "Ошибка парсинга ответа от Яндекса"
	}

	if len(result.Translations) > 0 {
		return result.Translations[0].Text
	}

	return "Перевод не найден"
}
