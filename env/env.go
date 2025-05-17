package env

import (
	"errors"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type (
	Env struct {
		TgBotAPIToken string `env:"TGBOT_API_TOKEN"`

		YandexAPIKey   string `env:"YANDEX_API_KEY"`
		YandexFolderId string `env:"YANDEX_FOLDER_ID"`

		GoogleAPIKey    string `env:"GOOGLE_API_KEY"`
		GoogleProjectId string `env:"GOOGLE_PROJECT_ID"`

		ChatBotApiKey string `env:"CHATBOT_API_KEY"`
		ChatBotApiUrl string `env:"CHATBOT_API_URL"`
		ChtaBotModel  string `env:"CHATBOT_MODEL"`
	}
)

func LoadEnv(path ...string) *Env {

	log.SetPrefix("[env]: ")

	fileparsed, err := parsefile(path...)
	if err != nil {
		panic(err)
	}

	env := new(Env)
	v := reflect.ValueOf(env).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)
		tag := structField.Tag.Get("env")
		if tag == "" {
			continue
		}
		val, ok := fileparsed[tag]
		if !ok {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(val)
		case reflect.Int, reflect.Int64:
			intVal, err := strconv.Atoi(val)
			if err != nil {
				panic(err)
			}
			field.SetInt(int64(intVal))
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(val)
			if err != nil {
				panic(err)
			}
			field.SetBool(boolVal)
		}
	}
	log.SetPrefix("")
	return env
}

// parsing .env file. In root dir by default.
func parsefile(path ...string) (map[string]string, error) {
	filePath := ".env"
	if len(path) == 1 && len(path[0]) > 0 {
		filePath = path[0]
	}
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	rows := strings.Split(string(fileBytes), "\n")
	if len(rows) == 0 {
		return nil, errors.New("no values in .env file")
	}

	parsedfile := make(map[string]string)
	for _, row := range rows {
		if row == "" || strings.HasPrefix(row, "#") {
			continue
		}

		keyvalue := strings.Split(row, "=")
		if len(keyvalue) != 2 {
			continue
		}

		key := strings.TrimSpace(keyvalue[0])
		value := strings.TrimSpace(keyvalue[1])
		parsedfile[key] = value
	}
	return parsedfile, nil
}
