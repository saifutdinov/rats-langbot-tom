package telegrambot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"saifutdinov/rats-langbot-tom/config"
	googletranslate "saifutdinov/rats-langbot-tom/translators/google-translate"
	yandextranslate "saifutdinov/rats-langbot-tom/translators/yandex-translate"
	"strings"
)

const telegramAPI = "https://api.telegram.org/bot"

type Update struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		ReplyToMessage *struct {
			Text string `json:"text"`
		} `json:"reply_to_message"`
	} `json:"message"`
}

func StartTelegramBot() {
	log.Println("Waiting for magic...")
	offset := 0
	for {
		updates := getUpdates(offset)
		for _, upd := range updates {
			offset = upd.UpdateID + 1
			handleCommand(upd)
		}
	}
}

func getUpdates(offset int) []Update {
	url := fmt.Sprintf("%s%s/getUpdates?offset=%d", telegramAPI, config.GetValue("TELEGRAM_TOKEN"), offset)
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Result
}

func sendMessage(chatID int64, text string) {
	body := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	b, _ := json.Marshal(body)
	http.Post(fmt.Sprintf("%s%s/sendMessage", telegramAPI, config.GetValue("TELEGRAM_TOKEN")), "application/json", bytes.NewBuffer(b))
}

func handleCommand(update Update) {
	if update.Message.Text == "/help" {
		sendMessage(update.Message.Chat.ID, `
			*Команды:*
			- /translate yandex — перевести сообщение
			- /translate ru-en google — с языками
			- /pool — сгенерировать вопрос
			`)
	}

	parts := strings.Fields(update.Message.Text)

	if len(parts) > 0 && parts[0] != "/translate" {
		return
	}

	if len(parts) < 2 {
		sendMessage(update.Message.Chat.ID, "Формат: /translate [ru-en] [yandex|google]")
		return
	}

	var fromLang, toLang, engine string

	if strings.Contains(parts[1], "-") {
		langs := strings.Split(parts[1], "-")
		if len(langs) != 2 {
			sendMessage(update.Message.Chat.ID, "Неверный формат языков. Пример: ru-en")
			return
		}
		fromLang, toLang = langs[0], langs[1]
		if len(parts) >= 3 {
			engine = parts[2]
		} else {
			engine = "yandex"
		}
	} else {
		fromLang = "auto"
		toLang = "en"
		engine = parts[1]
	}

	original := update.Message.ReplyToMessage.Text

	var translated string
	switch engine {
	case "yandex":
		translated = yandextranslate.YandexTranslateCustom(original, fromLang, toLang)
	case "google":
		translated = googletranslate.GoogleTranslate(original, fromLang, toLang)
	default:
		translated = "Неизвестный движок перевода: " + engine
	}

	sendMessage(update.Message.Chat.ID, translated)
}
