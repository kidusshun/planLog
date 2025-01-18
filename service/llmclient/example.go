package llmclient

import (
	"errors"
)


type llmclient struct {
	tools Tools
}

func NewLlmClient(tools Tools) *llmclient {
	return &llmclient{
		tools: tools,
	}
}

func (client *llmclient) CallGemini(messageHistory []Message, tools []Tool) (*GeminiResponseBody, error) {

	geminiRequest := GeminiRequestBody{
		SystemInstruction: map[string]interface{}{
			"parts": map[string]string{
				"text": SystemInstruction,
			},
		},
		Contents: messageHistory,
		Tools:    tools,
		ToolConfig: FunctionCallingConfig{
			FunctionCallingConfig: Mode{
				Mode: AUTO,
			},
		},
	}
	res, err := GeminiClient(geminiRequest)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (client *llmclient) HandleFunctionCall(geminiresponse *GeminiResponseBody) (*ToolCallResponse, error) {

	nameFunctionMap := map[string]interface{}{
		"CreateEvents": client.tools.CreateEvents,
		"FetchEvents":   client.tools.FetchEvents,
	}

	functionCalls := geminiresponse.Candidates[0].Content.Parts

	for _, call := range functionCalls {
		if call.FunctionCall.Name != "" {
			functionName := call.FunctionCall.Name
			args := call.FunctionCall.Args
			if function, exists := nameFunctionMap[functionName]; exists {
				switch functionName {
				case "QueryProducts":
					query, ok := args["query"].(string)
					if !ok {
						return &ToolCallResponse{}, errors.New("invalid argument for queryProducts")
					}
					result, err := function.(func(string) (*ToolCallResponse, error))(query)
					if err != nil {
						return &ToolCallResponse{}, err
					}
					return result, nil
				case "CompanyInfo":
					query, ok := args["query"].(string)
					if !ok {
						return &ToolCallResponse{}, errors.New("invalid argument for companyInfo")
					}
					result,err := function.(func(string) (*ToolCallResponse, error))(query)
					if err != nil {
						return &ToolCallResponse{},err
					}
					return result, nil
				case "TrackOrder":
					orderID, ok := args["orderID"].(string)
					if !ok {
						return &ToolCallResponse{}, errors.New("invalid argument for trackOrder")
					}
					result, err := function.(func(string) (*ToolCallResponse, error))(orderID)
					if err != nil {
						return &ToolCallResponse{}, err
					}
					return result, nil
				default:
					return &ToolCallResponse{},errors.New("function not found")
				}
				} else {
				return &ToolCallResponse{},errors.New("function not found")
			}
		}
	}
	return &ToolCallResponse{},errors.New("function not found")
}
