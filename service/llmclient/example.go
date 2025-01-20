package llmclient

import (
	"errors"
	"log"

	"github.com/kidusshun/planLog/service/user"
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

func (client *llmclient) HandleFunctionCall(userEntity *user.User, geminiresponse *GeminiResponseBody) (*ToolCallResponse, error) {
	nameFunctionMap := map[string]interface{}{
		"CreateEvent": client.tools.CreateEvents,
		"FetchEvents":   client.tools.FetchEvents,
	}

	functionCalls := geminiresponse.Candidates[0].Content.Parts
	log.Println("CALLS", functionCalls)

	for _, call := range functionCalls {
		if call.FunctionCall.Name != "" {
			functionName := call.FunctionCall.Name
			args := call.FunctionCall.Args
			log.Println("ARGS", args)
			log.Println("FUNCTION", functionName)
			if function, exists := nameFunctionMap[functionName]; exists {
				switch functionName {
				case "CreateEvent":
					description, ok := args["description"].(string)
					if !ok {
						return &ToolCallResponse{}, errors.New("invalid argument for queryProducts")
					}
					summary, ok := args["summary"].(string)
					if !ok {
						return &ToolCallResponse{}, errors.New("invalid argument for queryProducts")
					}
					startTime, ok := args["start_time"].(string)
					if !ok {
						return &ToolCallResponse{}, errors.New("invalid argument for queryProducts")
					}
					endTime, ok := args["end_time"].(string)
					if !ok {
						return &ToolCallResponse{}, errors.New("invalid argument for queryProducts")
					}
					result, err := function.(func(string, string,string,string, *user.User) (*ToolCallResponse, error))(summary, description,startTime, endTime, userEntity)
					if err != nil {
						log.Println("creating event error", err)
						return &ToolCallResponse{}, err
					}
					return result, nil
				case "FetchEvents":
					query, ok := args["query"].(string)
					if !ok {
						return &ToolCallResponse{}, errors.New("invalid argument for companyInfo")
					}
					result,err := function.(func(string) (*ToolCallResponse, error))(query)
					if err != nil {
						return &ToolCallResponse{},err
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
