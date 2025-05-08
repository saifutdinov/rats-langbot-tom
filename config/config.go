package config

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

var mx *sync.RWMutex
var configurations map[string]string

func LoadEnv() {
	mx = new(sync.RWMutex)
	configurations = make(map[string]string)

	file, err := os.Open(".env")
	if err != nil {
		panic("Не найден файл .env")
	}
	defer file.Close()

	mx.Lock()
	defer mx.Unlock()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			configurations[parts[0]] = parts[1]
		}
	}
}

func GetValue(envkey string) string {
	mx.RLock()
	defer mx.RUnlock()
	v, ok := configurations[envkey]
	if !ok {
		return ""
	}
	return v
}
