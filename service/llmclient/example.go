package llmclient

import (
	"fmt"

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

func (client *llmclient) CallGemini(messageHistory []Message, tools []Tool, systemMessage string) (*GeminiResponseBody, error) {
	geminiRequest := GeminiRequestBody{
		SystemInstruction: map[string]interface{}{
			"parts": map[string]string{
				"text": systemMessage,
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

func (client *llmclient) HandleFunctionCall(userEntity *user.User, geminiresponse *GeminiResponseBody) *Message {
	if geminiresponse == nil || len(geminiresponse.Candidates) == 0 || geminiresponse.Candidates[0].Content.Parts == nil {
        return nil
    }

	nameFunctionMap := map[string]interface{}{
		"CreateEvents": client.tools.CreateEvents,
		"FetchEvents":   client.tools.FetchEvents,
	}

	functionCalls := geminiresponse.Candidates[0].Content.Parts

	parts := []Part{}

	for _, call := range functionCalls {
		if call.FunctionCall.Name != "" {
			functionName := call.FunctionCall.Name
			args := call.FunctionCall.Args
			if function, exists := nameFunctionMap[functionName]; exists {
				switch functionName {
				case "CreateEvents":
					description, ok := args["description"].(string)
					if !ok {
						parts = append(parts, Part{
							FunctionResponse: &FunctionResponse{
								Name: "CreateEvents",
								Response: Result{
									Result: "description argument missing",
								},
						}})
						continue
					}
					summary, ok := args["summary"].(string)
					if !ok {
						parts = append(parts, Part{
							FunctionResponse: &FunctionResponse{
								Name: "CreateEvents",
								Response: Result{
									Result: "summary argument missing",
								},
						}})
						continue
					}
					startTime, ok := args["startTime"].(string)
					if !ok {
						parts = append(parts, Part{
							FunctionResponse: &FunctionResponse{
								Name: "CreateEvents",
								Response: Result{
									Result: "startTime argument missing",
								},
						}})
						continue
					}
					endTime, ok := args["endTime"].(string)
					if !ok {
						parts = append(parts, Part{
							FunctionResponse: &FunctionResponse{
								Name: "CreateEvents",
								Response: Result{
									Result: "endTime argument missing",
								},
						}})
						continue
					}
					calendarId, ok := args["calendarId"].(string)
					if !ok {
						parts = append(parts, Part{
							FunctionResponse: &FunctionResponse{
								Name: "CreateEvents",
								Response: Result{
									Result: "calendarId argument missing",
								},
						}})
						continue
					}
					result, err := function.(func(string, string,string,string,string, *user.User) (*ToolCallResponse, error))(summary, description,startTime, endTime,calendarId, userEntity)
					if err != nil {
						parts = append(parts, Part{
							FunctionResponse: &FunctionResponse{
								Name: "CreateEvents",
								Response: Result{
									Result: fmt.Sprintf("error: %s", err),
								},
						}})
						continue
					}
					parts = append(parts, result.Part)
				case "FetchEvents":
					startTime, ok := args["startTime"].(string)
					if !ok {
						parts = append(parts, Part{
							FunctionResponse: &FunctionResponse{
								Name: "FetchEvents",
								Response: Result{
									Result: "startTime argument missing",
								},
						}})
						continue
					}
					endTime, ok := args["endTime"].(string)
					if !ok {
						parts = append(parts, Part{
							FunctionResponse: &FunctionResponse{
								Name: "FetchEvents",
								Response: Result{
									Result: "endTime argument missing",
								},
						}})
						continue
					}
					calendarId, ok := args["FetchEvents"].(string)
					if !ok {
						parts = append(parts, Part{
							FunctionResponse: &FunctionResponse{
								Name: "CreateEvents",
								Response: Result{
									Result: "calendarId argument missing",
								},
						}})
						continue
					}
					result,_ := function.(func(string, string, string) (*ToolCallResponse, error))(startTime, endTime, calendarId)

					if result != nil {
						parts = append(parts, result.Part)
					}
				default:
					parts = append(parts, Part{
						FunctionResponse: &FunctionResponse{
							Name: "FetchEvents",
							Response: Result{
								Result: "wrong function call",
							},
					}})
				}
				} else {
					parts = append(parts, Part{
						FunctionResponse: &FunctionResponse{
							Name: "FetchEvents",
							Response: Result{
								Result: "wrong function call",
							},
					}})
			}
		}
	}
	message := Message{
		Role: USER,
		Parts: parts,
	}
	return &message
}
