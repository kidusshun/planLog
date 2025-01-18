package llmclient

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kidusshun/planLog/service/auth"
	"github.com/kidusshun/planLog/service/user"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

type QueryStore struct {
}

func NewQueryStore(db *sql.DB) *QueryStore {
	return &QueryStore{
	}
}

func (s *QueryStore) CreateEvents(summary, description string, user user.User) (*ToolCallResponse, error) {
	ctx := context.Background()

	oauthToken := &oauth2.Token{
		RefreshToken: user.GoogleRefreshToken, 
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
			DateTime: "2021-09-01T09:00:00-07:00",
			TimeZone: "Africa/Addis_Ababa",
		},
		End: &calendar.EventDateTime{
			DateTime: "2021-09-01T17:00:00-07:00",
			TimeZone: "Africa/Addis_Ababa",
		},

	}
	event, err = srv.Events.Insert("primary", event).Do()
	if err != nil {
		return &ToolCallResponse{}, err
	}

	response := Message{
		Role: USER,
		Parts: []Part{
			{
				FunctionResponse: &FunctionResponse{
					Name: "CreateCalendar",
					Response: Result{
						Result: fmt.Sprintf("Event created: %s\n", event.HtmlLink),
					},
				},
			},
		},
	}

	return &ToolCallResponse{
		ModelResponse:response,
	}, nil
}


func (s *QueryStore) FetchEvents(query string) (*ToolCallResponse, error)  {
	return nil, nil
}


func GetTools() []Tool {
	return []Tool{
		{
			FunctionDeclarations: []FunctionDeclaration{
				{
					Name:        "QueryProducts",
					Description: "a function to get products that matches the query passed",
					Parameters: Parameters{
						Type: "object",
						Properties: map[string]Property{
							"query": {
								Type: "string",
							},
						},
						Required: []string{"query"},
					},
				},
				{
					Name:        "CompanyInfo",
					Description: "a function to ask questions about a company's identity and general info",
					Parameters: Parameters{
						Type: "object",
						Properties: map[string]Property{
							"query": {
								Type: "string",
							},
						},
						Required: []string{"query"},
					},
				},
				{
					Name:        "TrackOrder",
					Description: "a function to get the location of a user's order ",
					Parameters: Parameters{
						Type: "object",
						Properties: map[string]Property{
							"orderID": {
								Type: "string",
							},
						},
						Required: []string{"orderID"},
					},
				},
			},
		},
	}
}