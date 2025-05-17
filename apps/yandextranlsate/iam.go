package yandextranlsate

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"
)

type yandexIAM struct {
	Token     string    `json:"iamToken"`
	ExpiresAt time.Time `json:"expiresAt"`
}

const iamTokenFile = "iam-token.json"

func getValidIAMToken(apiKey string) string {
	token := loadIAMTokenFromFile()

	if token == nil || time.Now().After(token.ExpiresAt) {
		newToken := requestNewIAMToken(apiKey)
		if newToken != nil {
			saveIAMTokenToFile(newToken)
			return newToken.Token
		}
		return ""
	}

	return token.Token
}

func loadIAMTokenFromFile() *yandexIAM {
	data, err := os.ReadFile(iamTokenFile)
	if err != nil {
		return nil
	}

	var token yandexIAM
	if err := json.Unmarshal(data, &token); err != nil {
		return nil
	}

	return &token
}

func saveIAMTokenToFile(token *yandexIAM) {
	data, _ := json.MarshalIndent(token, "", "  ")
	_ = os.WriteFile(iamTokenFile, data, 0600)
}

func requestNewIAMToken(oauthToken string) *yandexIAM {
	url := "https://iam.api.cloud.yandex.net/iam/v1/tokens"
	payload := map[string]string{"yandexPassportOauthToken": oauthToken}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		IAMToken  string `json:"iamToken"`
		ExpiresAt string `json:"expiresAt"` // ISO 8601
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil
	}

	exp, err := time.Parse(time.RFC3339, result.ExpiresAt)
	if err != nil {
		return nil
	}

	return &yandexIAM{
		Token:     result.IAMToken,
		ExpiresAt: exp,
	}
}
