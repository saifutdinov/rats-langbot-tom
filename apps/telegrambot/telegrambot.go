package telegrambot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var telegramAPI string

type (
	Update struct {
		UpdateID int      `json:"update_id"`
		Message  *Message `json:"message,omitempty"`
	}

	Message struct {
		MessageID int    `json:"message_id"`
		From      *User  `json:"from,omitempty"`
		Chat      Chat   `json:"chat"`
		Date      int    `json:"date"`
		Text      string `json:"text,omitempty"`

		ReplyToMessage    *Message    `json:"reply_to_message,omitempty"`
		NewChatMembers    []User      `json:"new_chat_members,omitempty"`
		LeftChatMember    *User       `json:"left_chat_member,omitempty"`
		NewChatTitle      string      `json:"new_chat_title,omitempty"`
		NewChatPhoto      []PhotoSize `json:"new_chat_photo,omitempty"`
		DeleteChatPhoto   bool        `json:"delete_chat_photo,omitempty"`
		GroupChatCreated  bool        `json:"group_chat_created,omitempty"`
		SupergroupCreated bool        `json:"supergroup_chat_created,omitempty"`
	}

	User struct {
		ID           int64  `json:"id"`
		IsBot        bool   `json:"is_bot"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name,omitempty"`
		Username     string `json:"username,omitempty"`
		LanguageCode string `json:"language_code,omitempty"`
	}

	Chat struct {
		ID    int64  `json:"id"`
		Type  string `json:"type"`
		Title string `json:"title,omitempty"`
	}

	PhotoSize struct {
		FileID   string `json:"file_id"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		FileSize int    `json:"file_size,omitempty"`
	}

	Config struct {
		APIToken    string
		HandlerFunc func(update Update) (string, *Poll, error)
	}

	Poll struct {
		ChatId          int64    `json:"chat_id"`
		Question        string   `json:"question"`
		Options         []string `json:"options"`
		IsAnonymous     string   `json:"is_anonymous"`
		Type            string   `json:"type"`
		CorrectOptionId string   `json:"correct_option_id"`
	}
)

func handlerFunc(update Update) (string, *Poll, error) {
	return update.Message.Text, nil, nil
}

func EscapeMarkdownV2(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

func Listen(config *Config) {
	fmt.Println("[lang-bot]: Start listening...")
	offset := 0

	telegramAPI = fmt.Sprintf("https://api.telegram.org/bot%s", config.APIToken)
	for {
		updates := getUpdates(offset)
		for _, update := range updates {
			offset = update.UpdateID + 1
			var handler func(update Update) (string, *Poll, error)
			if config.HandlerFunc != nil {
				handler = config.HandlerFunc
			} else {
				handler = handlerFunc
			}
			message, poll, err := handler(update)
			if err != nil {
				sendMessage(update.Message.Chat.ID, "Произошла ошибка и мы ее уже исправляем!")
				log.Printf("[lang-bot]: ERROR: %s\n", err.Error())
			}

			if message != "" {
				sendMessage(update.Message.Chat.ID, message)
			}

			if poll != nil {
				sendPoll(update.Message.Chat.ID, poll)
			}
		}
	}
}

// get updates(messages) for telegram bot
func getUpdates(offset int) []Update {
	url := fmt.Sprintf("%s/getUpdates?offset=%d", telegramAPI, offset)
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

// send message to telegram chat
func sendMessage(chatID int64, text string) {
	body := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	b, _ := json.Marshal(body)

	http.Post(fmt.Sprintf("%s/sendMessage", telegramAPI), "application/json", bytes.NewBuffer(b))
}

func sendPoll(chatID int64, poll *Poll) {
	poll.ChatId = chatID
	b, _ := json.Marshal(poll)
	http.Post(fmt.Sprintf("%s/sendPoll", telegramAPI), "application/json", bytes.NewBuffer(b))
}
