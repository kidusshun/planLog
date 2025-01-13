package user

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kidusshun/planLog/service/auth"
	"github.com/kidusshun/planLog/utils"
	"golang.org/x/oauth2"
)

type Handler struct {
	store UserStore
}

func NewHandler(store UserStore) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Post("/auth/google", h.googleSignupOrLogin)
	router.Get("/auth/google/callback", h.googleCallbackHandler)
	router.Post("/auth/google/signup", h.googleSignupOrLogin)
	router.With(auth.CheckBearerToken).Get("/user/me", h.getUser)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	url := auth.GoogleOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
func (h *Handler) googleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != "state-token" {
		http.Error(w, "State token mismatch", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	token, err := auth.GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := auth.GoogleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// write the user to database
	user := User{
		Name:           userInfo["name"].(string),
		Email:          userInfo["email"].(string),
		ProfilePicture: userInfo["picture"].(string),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	res, err := h.store.GetUserByEmail(user.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	if res != nil {
		// Send token back to the frontend
		sessionToken, err := auth.GenerateJWT(res.Email)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
		}
		http.Redirect(w, r, fmt.Sprintf("http://localhost:3000?token=%s", sessionToken), http.StatusSeeOther)
		return
	}

	_, err = h.store.CreateUser(user.Name, user.Email, user.ProfilePicture)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	sessionToken, err := auth.GenerateJWT(user.Email)
	if err != nil {
		http.Error(w, "Failed to generate session token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("http://localhost:3000?token=%s", sessionToken), http.StatusSeeOther)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	userEmail := r.Context().Value("userEmail").(string)
	log.Println("emaaaaaaaaaaaaaaail", userEmail)
	user, err := h.store.GetUserByEmail(userEmail)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) googleSignupOrLogin(w http.ResponseWriter, r *http.Request) {
	var body LoginPayload
	err := json.NewDecoder(r.Body).Decode(&body)
	fmt.Println("request", body.AccessToken)
	if err != nil {

		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	user, err := verifyGoogleToken(body.AccessToken)
	fmt.Println("error", err)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	userStored, err := h.store.GetUserByEmail(user.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if userStored == nil {

		user, err := h.store.CreateUser(user.Name, user.Email, user.Picture)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		jwtToken, err := auth.GenerateJWT(user.Email)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		response := LoginResponse{
			Token: jwtToken,
		}
		utils.WriteJSON(w, http.StatusOK, response)

	}
	jwtToken, err := auth.GenerateJWT(user.Email)

	response := LoginResponse{
		Token: jwtToken,
	}
	utils.WriteJSON(w, http.StatusOK, response)

}

