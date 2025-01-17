package llmclient

import (
	"encoding/json"

	"github.com/kidusshun/ecom_bot/service/product"
)

type LlmClient interface{
	CallGemini(messageHistory []Message, tools []Tool) (*GeminiResponseBody, error)
	HandleFunctionCall(geminiresponse *GeminiResponseBody)(*ToolCallResponse, error)
}

type Tools interface{
	CompanyInfo(query string) (*ToolCallResponse, error)
	QueryProducts(query string) (*ToolCallResponse, error)
}

type GeminiRequestBody struct {
	SystemInstruction map[string]interface{} `json:"system_instruction,omitempty"`
	Contents          []Message              `json:"contents,omitempty"`
	ToolConfig        FunctionCallingConfig  `json:"tool_config,omitempty"`
	Tools             []Tool                 `json:"tools,omitempty"`
}

type RoleEnum int

const (
	USER RoleEnum = iota + 1
	SYSTEM
	MODEL
)

func (r RoleEnum) String() string {
	switch r {
    case 1:
        return "user"
    case 2:
        return "system"
    default:
        return "model"
    }
}

type ModeEnum int

const (
	AUTO ModeEnum = iota + 1
	ANY
	NONE
)

func (m ModeEnum) String() string {
	return []string{"AUTO", "ANY", "NONE"}[m-1]
}

type FunctionDeclaration struct {
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Parameters  Parameters `json:"parameters,omitempty"`
}

type Parameters struct {
	Type       string              `json:"type,omitempty"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"function_declarations,omitempty"`
}

type Message struct {
	Role  RoleEnum               `json:"role,omitempty"`
	Parts []Part `json:"parts,omitempty"`
}

type Mode struct {
	Mode ModeEnum `json:"mode,omitempty"`
}

type FunctionCallingConfig struct {
	FunctionCallingConfig Mode `json:"function_calling_config,omitempty"`
}

func (m Message) MarshalJSON() ([]byte, error) {
	type Alias Message // Create an alias to avoid infinite recursion in MarshalJSON
	return json.Marshal(&struct {
		Role string `json:"role"`
		Alias
	}{
		Role:  m.Role.String(), // Use the string representation of RoleEnum
		Alias: (Alias)(m),
	})
}

func (m Mode) MarshalJSON() ([]byte, error) {
	type Alias Mode
	return json.Marshal(&struct {
		Mode string `json:"mode"`
		Alias
	}{
		Mode:  m.Mode.String(),
		Alias: (Alias)(m),
	})
}

type GeminiResponseBody struct {
	Candidates    []Candidate   `json:"candidates"`
	UsageMetadata UsageMetadata `json:"usageMetadata"`
	ModelVersion  string        `json:"modelVersion"`
}

type Candidate struct {
	Content       Content        `json:"content"`
	FinishReason  string         `json:"finishReason"`
	AvgLogProbs   *float64       `json:"avgLogProbs,omitempty"`
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
}

type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type Content struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role"`
}

type ImageData struct {
	MimeType string `json:"mime_type"`
	Data string `json:"data"`
}

type Part struct {
	Text         string       `json:"text,omitempty"`
	InlineData *ImageData `json:"inline_data,omitempty"`
	FunctionCall *FunctionCall `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
}

type FunctionResponse struct {
	Name string 			   `json:"name"`
	Response Result `json:"response"`
}

type Result struct {
	Result string
}

type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

type ToolCallResponse struct {
	ModelResponse Message `json:"model_response"`
	Location string `json:"location"`
	Products []product.Product `json:"products"`
}