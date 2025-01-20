package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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

	timeRightNow := fmt.Sprintf(`
	The time and date in ISO 8601 format right now is %s
	This is the reference time to be used when you are to create and fetch events by the user
	`, time.Now().Format(time.RFC3339))

	planCalendarId, logCalendarId, err := chat.store.GetCalendarIDByUserID(userEntity.ID)

	if err != nil {
		return llmclient.ChatResponse{}, err
	}

	userCalendars := fmt.Sprintf(`
	The plan calendar id is %s
	The log calendar id is %s

	This are the id's to be used when creating and fetching events.
	`, planCalendarId, logCalendarId)

	systemMessage := llmclient.SystemInstruction + timeRightNow + userCalendars

	response, err := chat.client.CallGemini(chatHistory, tools, systemMessage)



	if err != nil {
		return llmclient.ChatResponse{}, err
	}
	chatHistory = append(chatHistory, llmclient.Message{
		Role: llmclient.MODEL,
		Parts: response.Candidates[0].Content.Parts,
	})

	var chatResponse llmclient.ChatResponse

	for response.Candidates[0].Content.Parts[0].FunctionCall != nil {
		call_result := chat.client.HandleFunctionCall(userEntity, response)

		
		messageString, err := json.Marshal(call_result)
		if err != nil {
			log.Println("can't marshal")
		}
		log.Println("RESPONSE TO FUNCTION CALL",string(messageString))
		chatHistory = append(chatHistory, *call_result)

		response, err = chat.client.CallGemini(chatHistory, tools, systemMessage)

		if err != nil {
			log.Println(err)
			return llmclient.ChatResponse{}, err
		}

		chatHistory = append(chatHistory, llmclient.Message{
			Role:  llmclient.MODEL,
			Parts: response.Candidates[0].Content.Parts,
		})
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