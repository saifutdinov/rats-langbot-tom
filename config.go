package main

import (
	"bufio"
	"os"
	"strings"
)

var config map[string]string

func LoadEnv() {
	config = make(map[string]string)
	file, err := os.Open(".env")
	if err != nil {
		panic("Не найден файл .env")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			config[parts[0]] = parts[1]
		}
	}
}
