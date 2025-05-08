package googletranslate

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"saifutdinov/rats-langbot-tom/config"
	"strings"
)

func GoogleTranslate(text, fromLang, toLang string) string {
	endpoint := "https://translation.googleapis.com/language/translate/v2"

	// Если fromLang = "auto", Google просто не указывает source
	data := url.Values{}
	data.Set("q", text)
	data.Set("target", toLang)
	if fromLang != "auto" {
		data.Set("source", fromLang)
	}
	data.Set("format", "text")
	data.Set("key", config.GetValue("GOOGLE_API_KEY"))

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "Ошибка создания запроса к Google Translate"
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "Ошибка при обращении к Google Translate API"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			Translations []struct {
				TranslatedText string `json:"translatedText"`
			} `json:"translations"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "Ошибка разбора ответа от Google"
	}

	if len(result.Data.Translations) > 0 {
		return htmlUnescape(result.Data.Translations[0].TranslatedText)
	}

	return "Перевод не получен"
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
