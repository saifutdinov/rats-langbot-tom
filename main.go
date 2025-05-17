package main

import (
	"encoding/json"
	"fmt"
	openrouterai "saifutdinov/rats-langbot-tom/apps/openrouter.ai"
	"saifutdinov/rats-langbot-tom/apps/telegrambot"
	"saifutdinov/rats-langbot-tom/apps/yandextranlsate"
	"saifutdinov/rats-langbot-tom/env"
	"strconv"
	"strings"
)

const (
	translateCommand string = "/translate"
	poolCommand      string = "/pool"
	explainCommand   string = "/explain"
	helpCommand      string = "/help"
)

const (
	yandexEngine string = "yandex"
	// googleEngine string = "google"
)

type (
	BotCommand struct {
		Command          string
		LangFrom, LangTo string
		Engine           string
	}
	quiz struct {
		Question  string   `json:"question"`
		Variables []string `json:"variables"`
		Answer    int      `json:"answer"`
	}
	explanation struct {
		Phrase  string `json:"phrase"`
		Meaning string `json:"meaning"`
		Example string `json:"example"`
	}
)

func main() {
	//load .env file
	config := env.LoadEnv()

	yandex := yandextranlsate.NewYandexEngine(config.YandexAPIKey, config.YandexFolderId)

	// disabled - payment required(. I'm broke(
	// google := googletranslate.NewGoogleEngine(config.GoogleAPIKey, config.GoogleProjectId)

	chatbot := openrouterai.NewChatClient(config.ChatBotApiKey, config.ChatBotApiUrl, config.ChtaBotModel)

	handleCommand := func(update telegrambot.Update) (string, *telegrambot.Poll, error) {
		if update.Message == nil {
			return "", nil, nil
		}
		if update.Message.Text[0] != '/' {
			return "", nil, nil
		}
		message := update.Message
		text := message.Text

		// fmt.Println(text)

		command, err := extract(text)
		if err != nil {
			return "", nil, err
		}

		// fmt.Println(command)

		reply := message.ReplyToMessage

		fmt.Println("read request:", command)

		switch command.Command {
		case translateCommand:
			{
				if reply == nil {
					return "Нужно запросить перевод ответом на сообщение!", nil, nil
				}
				switch command.Engine {
				case yandexEngine:
					{
						tranlatedText, err := yandex.Translate(reply.Text, command.LangFrom, command.LangTo)
						if err != nil {
							return "", nil, err
						}

						return tranlatedText, nil, nil
					}
					// case googleEngine:
					// 	{
					// 		return google.Translate(update.Message.ReplyToMessage.Text, command.LangFrom, command.LangTo)
					// 	}
				}
			}
		case poolCommand:
			{
				if reply == nil {
					return "Нужно запросить перевод ответом на сообщение!", nil, nil
				}

				command.LangFrom, err = yandex.DetectLang(reply.Text)
				if err != nil {
					return "", nil, err
				}

				command.LangTo = getMirroredLanguage(command.LangFrom)
				prompt := PoolPrompt(reply.Text, command.LangTo)

				response, err := chatbot.RequestChat(prompt)
				if err != nil {
					return "", nil, err
				}

				q := new(quiz)
				if err = json.Unmarshal([]byte(response), &q); err != nil {
					return "", nil, err
				}

				poll := MakePoolMarkdown(q)
				return "", poll, nil
			}
		case explainCommand:
			{
				if reply == nil {
					return "Нужно запросить перевод ответом на сообщение!", nil, nil
				}
				command.LangFrom, err = yandex.DetectLang(reply.Text)
				if err != nil {
					return "", nil, err
				}
				command.LangTo = getMirroredLanguage(command.LangFrom)

				//
				prompt := ExplainPrompt(reply.Text, command.LangFrom, command.LangTo)

				response, err := chatbot.RequestChat(prompt)
				if err != nil {
					return "", nil, err
				}

				e := new(explanation)
				if err = json.Unmarshal([]byte(response), &e); err != nil {
					return "", nil, err
				}

				explain := MakeExplainMarkdown(e)
				return explain, nil, nil
			}
		case helpCommand:
			{
				return "*Список команд:*\n" +
					"/translate — автоопределение → русский \\(Яндекс\\)\n" +
					"/translate ru-en — перевод с русского на английский\n" +
					"/pool — генерация вопроса\n" +
					// "/explain — объяснение смысла\n" +
					"/help — показать это сообщение", nil, nil
			}
		}
		return "Не понял, что вы пытаетесь сделать?:(", nil, nil
	}

	tgconfig := &telegrambot.Config{
		APIToken:    config.TgBotAPIToken,
		HandlerFunc: handleCommand,
	}
	// run tg bot listener
	telegrambot.Listen(tgconfig)
}

// sometimes command comes with bot tag /command@bot_name
func extract(commandtext string) (*BotCommand, error) {
	commandParams := strings.Fields(commandtext)

	// 1. What bot command we extracting
	var command string
	commandParts := strings.Split(commandParams[0], "@")
	command = commandParts[0]

	// 2. Languages
	var langFrom, langTo, engine string

	commandlen := len(commandParams)

	if commandlen >= 2 {
		if strings.Contains(commandParams[1], "-") {
			langs := strings.Split(commandParams[1], "-")
			langFrom, langTo = langs[0], langs[1]
			if commandlen >= 3 {
				engine = commandParams[2]
			} else {
				engine = yandexEngine
			}
		} else {
			langFrom = "auto"
			langTo = "ru"
			engine = commandParams[1]
		}
	}

	return &BotCommand{
		Command:  command,
		LangFrom: langFrom,
		LangTo:   langTo,
		Engine:   engine,
	}, nil

}

// TOTAL CUSTOM
func getMirroredLanguage(langFrom string) string {
	if langFrom == "ru" {
		return "en"
	}
	return "ru"
}

// POOL
func PoolPrompt(text, langTo string) string {
	return fmt.Sprintf(`Напиши на "%s" языке понятный текст проверяющего знания языка "%s" вопроса и дай 4 варианта ответа. Контекст должен строиться вокруг фразы "%s".
формат строго в JSON
{
"question":"",
"variables":["вариант1","вариант2","вариант3","вариант4"],
"answer": 1
}`, langTo, langTo, text)
}

func MakePoolMarkdown(q *quiz) *telegrambot.Poll {
	return &telegrambot.Poll{
		Question:        q.Question,
		Options:         q.Variables,
		IsAnonymous:     "false",
		Type:            "quiz",
		CorrectOptionId: strconv.Itoa(q.Answer),
	}
}

// EXPLAIN
func ExplainPrompt(text, langFrom, langTo string) string {
	return fmt.Sprintf(`Поясни фразу на "%s" языке: "%s". Объясни смысл, переведи на "%s" и приведи короткий пример. Ответ строго в JSON:
{
  "phrase": "сама фраза на нужном языке",
  "meaning": "перевод и смысл",
  "example": "пример использования"
}`, langFrom, text, langTo)
}

func MakeExplainMarkdown(e *explanation) string {
	return fmt.Sprintf("*Фраза:* %s\n*Значение:* %s\n_Пример:_\n%s",
		telegrambot.EscapeMarkdownV2(e.Phrase),
		telegrambot.EscapeMarkdownV2(e.Meaning),
		telegrambot.EscapeMarkdownV2(e.Example),
	)
}
