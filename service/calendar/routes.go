package calendar

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kidusshun/planLog/service/auth"
	"github.com/kidusshun/planLog/service/user"
	"github.com/kidusshun/planLog/service/whisper"
	"github.com/kidusshun/planLog/utils"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

type Handler struct {
	store user.UserStore
}

func NewHandler(store user.UserStore) *Handler {
	return &Handler{
		store: store,
	}
}



func (h *Handler) RegisterRoutes(router chi.Router) {
	router.With(auth.CheckBearerToken).Post("/initialize_calendar",h.createPlanAndLogCalendar)
	router.With(auth.CheckBearerToken).Post("/Plan_or_log", h.planAndLog)
}

func (h *Handler) createPlanAndLogCalendar(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	
	userEmail := r.Context().Value("userEmail").(string)
	user, err := h.store.GetUserByEmail(userEmail)

	if err != nil {
		log.Println("error1",err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// create calendar
	log.Println("hereeeeee", user.GoogleRefreshToken)
	oauthToken := &oauth2.Token{
		RefreshToken: user.GoogleRefreshToken, 
	}
	client := auth.GoogleOAuthConfig.Client(ctx,oauthToken)

	srv, err := calendar.New(client)
	if err != nil {
		log.Println("error2",err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	planCal := &calendar.Calendar{
		Summary:     "Plans",
		Description: "This calendar contains all planned events and tasks.",
	}
	planCalendar, err := srv.Calendars.Insert(planCal).Do()
	if err != nil {
		log.Println("error3",err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	planListEntry := &calendar.CalendarListEntry{
		Id: planCalendar.Id,
	}
	insertedPlanEntry, err := srv.CalendarList.Insert(planListEntry).Do()
	if err != nil {
		log.Printf("Error inserting Plans calendar to list: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	planColorUpdate := &calendar.CalendarListEntry{
		ColorId: "8", // Graphite color
	}
	updatedPlanCal, err := srv.CalendarList.Patch(insertedPlanEntry.Id, planColorUpdate).Do()
	if err != nil {
		log.Printf("Error updating Plans calendar color: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	logCal := &calendar.Calendar{
		Summary:     "Logs",
		Description: "This calendar contains all logged events and tasks.",
	}

	logCalendar, err := srv.Calendars.Insert(logCal).Do()
	if err != nil {
		log.Println("error3",err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	logListEntry := &calendar.CalendarListEntry{
		Id: logCalendar.Id,
	}
	insertedLogEntry, err := srv.CalendarList.Insert(logListEntry).Do()
	if err != nil {
		log.Printf("Error inserting Logs calendar to list: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	logColorUpdate := &calendar.CalendarListEntry{
		ColorId: "8", // Graphite color
	}
	updatedLogCal, err := srv.CalendarList.Patch(insertedLogEntry.Id, logColorUpdate).Do()
	if err != nil {
		log.Printf("Error updating Logs calendar color: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	

	utils.WriteJSON(w, http.StatusOK, struct {
		PlanCalendar *calendar.CalendarListEntry `json:"planCalendar"`
		LogCalendar  *calendar.CalendarListEntry `json:"logCalendar"`
	}{
		PlanCalendar: updatedPlanCal,
		LogCalendar:  updatedLogCal,
	})
}

func (h *Handler) planAndLog(w http.ResponseWriter,r *http.Request) {
	audioData, fileName, err := parseAudioRequest(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	transcription, err := whisper.TranscribeAudio(audioData, fileName)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, transcription)

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