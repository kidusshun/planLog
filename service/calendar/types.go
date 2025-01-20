package calendar

import (
	"github.com/kidusshun/planLog/service/llmclient"
	"github.com/kidusshun/planLog/service/whisper"
	"google.golang.org/api/calendar/v3"
)

type CalendarService interface {
	Chat(userEmail string, request llmclient.ChatRequest) (llmclient.ChatResponse, error)
	CreateCalendar(userEmail string) (*CreateCalendarResponse, error)
	Transcribe(audioData []byte, planOrLog string) (*whisper.WhisperResponseBody, error)
	GetCalendars(userEmail string) ([]string, error)
}

type CreateCalendarResponse struct {
	PlanCalendar *calendar.CalendarListEntry `json:"planCalendar"`
	LogCalendar  *calendar.CalendarListEntry `json:"logCalendar"`
}