package llmclient

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kidusshun/planLog/service/auth"
	"github.com/kidusshun/planLog/service/user"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

type QueryStore struct {
}

func NewQueryStore() *QueryStore {
	return &QueryStore{
	}
}

func (s *QueryStore) CreateEvents(summary, description, startTime, endTime, calendarId string, userEntity *user.User) (*ToolCallResponse, error) {
	ctx := context.Background()

	oauthToken := &oauth2.Token{
		RefreshToken: userEntity.GoogleRefreshToken, 
	}
	client := auth.GoogleOAuthConfig.Client(ctx,oauthToken)

	srv, err := calendar.New(client)
	if err != nil {
		return nil, err
	}

	event := &calendar.Event{
		Summary:     summary,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: startTime,
			TimeZone: "Africa/Addis_Ababa",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime,
			TimeZone: "Africa/Addis_Ababa",
		},

	}
	event, err = srv.Events.Insert(calendarId, event).Do()
	if err != nil {
		log.Println("error creating event", err)
		return &ToolCallResponse{}, err
	}

	part := Part{
			FunctionResponse: &FunctionResponse{
				Name: "CreateCalendar",
				Response: Result{
					Result: fmt.Sprintf("Event created: %s\n", event.HtmlLink),
				},
			},
	}
	
	return &ToolCallResponse{
		Part: part,
	}, nil
}


func (s *QueryStore) FetchEvents(startTime, endTime, calendarId string, userEntity *user.User) (*ToolCallResponse, error)  {
	ctx := context.Background()
	start, err := time.Parse(time.RFC3339, startTime)
	
	if err != nil {
		log.Printf("Invalid start time format: %v", err)
		return nil, err
	}
	end, err := time.Parse(time.RFC3339, endTime)
	
	if err != nil {
		log.Printf("Invalid start time format: %v", err)
		return nil, err
	}

	oauthToken := &oauth2.Token{
		RefreshToken: userEntity.GoogleRefreshToken, 
	}
	client := auth.GoogleOAuthConfig.Client(ctx,oauthToken)

	srv, err := calendar.New(client)

	if err != nil {
		return nil, err
	}

	eventsCall := srv.Events.List(calendarId).
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		OrderBy("startTime").
		SingleEvents(true)

	events, err := eventsCall.Do()
	if err != nil {
		return nil, err
	}

	eventsJSON, err := json.Marshal(events.Items)
	if err != nil {
		return nil, err
	}

	eventsString := string(eventsJSON)
	part := Part{
				FunctionResponse: &FunctionResponse{
					Name: "CreateCalendar",
					Response: Result{
						Result: eventsString,
					},
				},
		}

	return &ToolCallResponse{
		Part: part,
	}, nil

}


func GetTools() []Tool {
	return []Tool{
		{
			FunctionDeclarations: []FunctionDeclaration{
				{
					Name:        "CreateEvents",
					Description: "a function that lets you create events on google calendar",
					Parameters: Parameters{
						Type: "object",
						Properties: map[string]Property{
							"summary": {
								Type: "string",
							},
							"description": {
								Type: "string",
							},
							"startTime": {
								Type: "string",
							},
							"endTime": {
								Type: "string",
							},
							"calendarId": {
								Type: "string",
							},
						},
						Required: []string{"summary", "description", "startTime", "endTime", "calendarId"},
					},
				},
				{
					Name:        "FetchEvents",
					Description: "a function to retrieve events within a given timeframe",
					Parameters: Parameters{
						Type: "object",
						Properties: map[string]Property{
							"startTime": {
								Type: "string",
							},
							"endTime": {
								Type: "string",
							},
							"calendarId": {
								Type: "string",
							},
						},
						Required: []string{"startTime", "endTime", "calendarId"},
					},
				},
			},
		},
	}
}