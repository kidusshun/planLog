package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kidusshun/planLog/service/auth"
	"github.com/kidusshun/planLog/service/llmclient"
	"github.com/kidusshun/planLog/service/user"
	"github.com/kidusshun/planLog/utils"
	"golang.org/x/oauth2"
)

type Handler struct {
	service CalendarService
	store user.UserStore
}

func NewHandler(store user.UserStore, service CalendarService) *Handler {
	return &Handler{
		service: service,
		store: store,
	}
}



func (h *Handler) RegisterRoutes(router chi.Router) {
	router.With(auth.CheckBearerToken).Post("/initialize_calendar",h.createPlanAndLogCalendar)
	router.With(auth.CheckBearerToken).Post("/plan_or_log", h.planAndLog)
	router.With(auth.CheckBearerToken).Post("/analyze_calendar", h.auditCalendar)

}

func (h *Handler) createPlanAndLogCalendar(w http.ResponseWriter, r *http.Request) {
	
	userEmail := r.Context().Value("userEmail").(string)
	userEntity, err := h.store.GetUserByEmail(userEmail)
	if err != nil {
		log.Println("err1",err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	response, err := h.service.CreateCalendar(userEmail)
	
	if err != nil {
		log.Println("err2",err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	err = h.store.AddCalendarIDToUser(userEntity.ID, response.PlanCalendar.Id, response.LogCalendar.Id)
	
	if err != nil {
		log.Println("err2",err)
		utils.WriteError(w, http.StatusInternalServerError, err)
	}


	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) planAndLog(w http.ResponseWriter,r *http.Request) {
	userEmail := r.Context().Value("userEmail").(string)
	audioData, fileName, err := parseAudioRequest(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	response, err := h.service.Transcribe(audioData, fileName)
	
	if err != nil {
		log.Println("transcription err", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}	
	log.Println("transcription", response.Text)
	// message := r.FormValue("message")
	chatResponse, err := h.service.Chat(userEmail, llmclient.ChatRequest{
		Message: response.Text,
	})

	if err != nil {
		log.Print()
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, chatResponse)
}

func (h *Handler) auditCalendar(w http.ResponseWriter, r *http.Request) {
	userEmail := r.Context().Value("userEmail").(string)
	// Parse request body
	var reqBody TimeRangeRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("Error parsing request body: %v", err)
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	defer r.Body.Close()

	response, err := h.service.AnalyzeEvents(userEmail, reqBody.Start, reqBody.End)
	if err != nil {
		log.Printf("Error analyzing events: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func parseAudioRequest(r *http.Request) ([]byte, string, error) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return nil, "", err
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	// Read file into a byte slice
	audioData, err := io.ReadAll(file)
	if err != nil {
		return nil, "", err
	}
	return audioData, fileHeader.Filename, nil
}


func RefreshOAuthToken(ctx context.Context, userEmail string, userStore user.UserStore, config *oauth2.Config, oldRefreshToken string) (*http.Client, error) {
	log.Println("CAAAAAAAALLLLLED")
	// Create a token source using the refresh token
	tokenSource := config.TokenSource(ctx, &oauth2.Token{RefreshToken: oldRefreshToken})

	// Get a new token
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Println("this",err)
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Check if Google returned a new refresh token
	if newToken.RefreshToken != "" && newToken.RefreshToken != oldRefreshToken {
		log.Println("New refresh token received, updating database...")

		// Update the refresh token in the database
		err := userStore.UpdateUserRefreshToken(userEmail, newToken.RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("failed to update refresh token in DB: %w", err)
		}
	}
	log.Println("neeeeeeew",newToken.RefreshToken)

	client := config.Client(ctx, newToken)
	return client, nil
}