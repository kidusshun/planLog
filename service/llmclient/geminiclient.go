package llmclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/kidusshun/ecom_bot/config"
)

func GeminiClient(userRequest GeminiRequestBody) (*GeminiResponseBody, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-pro-002:generateContent?key=%s", config.GeminiEnvs.GeminiAPIKey)
	jsonRequest, err := json.Marshal(userRequest)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonRequest))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		return nil, errors.New(string(bodyBytes))
	}

	geminiResponse := parseGeminiResponse(res.Body)
	return geminiResponse, nil
}

func parseGeminiResponse(responseBody io.ReadCloser) *GeminiResponseBody {

	var geminiResponse GeminiResponseBody

	decoder := json.NewDecoder(responseBody)
	err := decoder.Decode(&geminiResponse)

	if err != nil {
		log.Fatal("error decoding json: ", err)
	}

	return &geminiResponse

}
