package whisper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/kidusshun/planLog/config"
)




func TranscribeAudio(audioData []byte, fileName string) (*WhisperResponseBody, error) {
	url := "https://api.groq.com/openai/v1/audio/transcriptions"

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	_ = writer.WriteField("model", "whisper-large-v3-turbo")
	_ = writer.WriteField("response_format", "verbose_json")

	filePart, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}

	_, err = filePart.Write(audioData)
	if err != nil {
		return nil, err
	}

	writer.Close()

	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.GeminiEnvs.GroqAPIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		return nil, errors.New(string(bodyBytes))
	}

	whisperResponse := parseWhisperResponse(res.Body)
	return whisperResponse, nil
}


func parseWhisperResponse(responseBody io.ReadCloser) *WhisperResponseBody {

	var whisperResponse WhisperResponseBody

	decoder := json.NewDecoder(responseBody)
	err := decoder.Decode(&whisperResponse)

	if err != nil {
		log.Fatal("error decoding json: ", err)
	}

	return &whisperResponse

}