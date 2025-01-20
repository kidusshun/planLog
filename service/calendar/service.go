package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/kidusshun/planLog/service/auth"
	"github.com/kidusshun/planLog/service/llmclient"
	"github.com/kidusshun/planLog/service/user"
	"github.com/kidusshun/planLog/service/whisper"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)


type Service struct {
	client llmclient.LlmClient
	store user.UserStore
}

func NewService(client llmclient.LlmClient, store user.UserStore) *Service {
	return &Service{
		client: client,
		store: store,
	}
}

func (chat *Service) Chat(userEmail string, request llmclient.ChatRequest) (llmclient.ChatResponse, error) {
	userEntity, err := chat.store.GetUserByEmail(userEmail)
	if err != nil {
		return llmclient.ChatResponse{}, err
	}
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

	log.Println("RESPONSE", response.Candidates[0].Content.Parts[0].FunctionCall)

	if err != nil {
		return llmclient.ChatResponse{}, err
	}
	chatHistory = append(chatHistory, llmclient.Message{
		Role: llmclient.MODEL,
		Parts: response.Candidates[0].Content.Parts,
	})

	var chatResponse llmclient.ChatResponse

	for response.Candidates[0].Content.Parts[0].FunctionCall != nil {
		call_result, err := chat.client.HandleFunctionCall(userEntity, response)
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


func (service *Service) CreateCalendar(userEmail string) (*CreateCalendarResponse, error) {
	ctx := context.Background()
	user, err := service.store.GetUserByEmail(userEmail)

	if err != nil {
		return nil, err
	}

	oauthToken := &oauth2.Token{
		RefreshToken: user.GoogleRefreshToken, 
	}
	client := auth.GoogleOAuthConfig.Client(ctx,oauthToken)

	srv, err := calendar.New(client)
	if err != nil {
		return nil, err
	}

	planCal := &calendar.Calendar{
		Summary:     "Plans",
		Description: "This calendar contains all planned events and tasks.",
	}
	planCalendar, err := srv.Calendars.Insert(planCal).Do()
	
	if err != nil {
		return nil, err
	}

	planListEntry := &calendar.CalendarListEntry{
		Id: planCalendar.Id,
	}
	
	insertedPlanEntry, err := srv.CalendarList.Insert(planListEntry).Do()
	if err != nil {
		return nil, err
	}

	planColorUpdate := &calendar.CalendarListEntry{
		ColorId: "16", // Graphite color
	}
	updatedPlanCal, err := srv.CalendarList.Patch(insertedPlanEntry.Id, planColorUpdate).Do()
	if err != nil {
		return nil, err
	}

	logCal := &calendar.Calendar{
		Summary:     "Logs",
		Description: "This calendar contains all logged events and tasks.",
	}

	logCalendar, err := srv.Calendars.Insert(logCal).Do()
	if err != nil {
		return nil, err
	}

	logListEntry := &calendar.CalendarListEntry{
		Id: logCalendar.Id,
	}
	insertedLogEntry, err := srv.CalendarList.Insert(logListEntry).Do()
	if err != nil {
		return nil, err
	}

	logColorUpdate := &calendar.CalendarListEntry{
		ColorId: "19", // Graphite color
	}
	updatedLogCal, err := srv.CalendarList.Patch(insertedLogEntry.Id, logColorUpdate).Do()
	if err != nil {
		return nil, err
	}

	return &CreateCalendarResponse{
		PlanCalendar: updatedPlanCal,
		LogCalendar:  updatedLogCal,
	}, nil
}
func (service *Service) Transcribe(audioData []byte, fileName string) (*whisper.WhisperResponseBody, error) {
	transcription, err := whisper.TranscribeAudio(audioData, fileName)
	if err != nil {
		return nil, err
	}

	return transcription, nil
}

func (service *Service) GetCalendars(userEmail string) ([]string, error) {
	return nil, nil
}