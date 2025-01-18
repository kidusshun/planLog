package calendar

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/kidusshun/planLog/service/llmclient"
)


type Service struct {
	client llmclient.LlmClient
}

func (chat *Service) Chat(request llmclient.ChatRequest) (llmclient.ChatResponse, error) {
	tools := llmclient.GetTools()
	chatHistory := []llmclient.Message{
		{
			Role: llmclient.USER,
			Parts: []llmclient.Part{
				{
					Text: request.Message,
				},
			},
		},
	}

	response, err := chat.client.CallGemini(chatHistory, tools)

	if err != nil {
		return llmclient.ChatResponse{}, err
	}
	chatHistory = append(chatHistory, llmclient.Message{
		Role: llmclient.MODEL,
		Parts: response.Candidates[0].Content.Parts,
	})

	var chatResponse llmclient.ChatResponse

	for response.Candidates[0].Content.Parts[0].FunctionCall != nil {
		call_result, err := chat.client.HandleFunctionCall(response)
		if err != nil {
			return llmclient.ChatResponse{}, nil
		}
		chatHistory = append(chatHistory, call_result.ModelResponse)

		response, err = chat.client.CallGemini(chatHistory, tools)

		chatHistory = append(chatHistory, llmclient.Message{
			Role:  llmclient.MODEL,
			Parts: response.Candidates[0].Content.Parts,
		})
		if err != nil {
			return llmclient.ChatResponse{}, err
		}
		str, err := json.Marshal(response)
		if err != nil {
			log.Print(err)
			return llmclient.ChatResponse{}, err
		}
		fmt.Println(string(str))
	}

	chatResponse.Content = response.Candidates[0].Content.Parts[0].Text
	chatResponse.Role = "model"

	return chatResponse, nil
}