package main

import (
	"saifutdinov/rats-langbot-tom/config"
	telegrambot "saifutdinov/rats-langbot-tom/telegram-bot"
)

func main() {
	config.LoadEnv()
	telegrambot.StartTelegramBot()
}
