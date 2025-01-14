package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kidusshun/planLog/service/auth"
	"github.com/kidusshun/planLog/utils"
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
	router.Post("/auth/google/signup", h.googleSignupOrLogin)
	router.With(auth.CheckBearerToken).Get("/user/me", h.getUser)
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
		log.Println("error1", err)
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	user, err := verifyGoogleToken(body.AccessToken)

	if err != nil {
		log.Println("error2", err)
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	userStored, err := h.store.GetUserByEmail(user.Email)

	if userStored == nil {

		if err == sql.ErrNoRows {

			user, err := h.store.CreateUser(user.Name, user.Email,user.GoogleAccessToken, user.Picture)
			if err != nil {
				log.Println("error4", err)
				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}
			jwtToken, err := auth.GenerateJWT(user.Email)
			if err != nil {
				log.Println("error5", err)
				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}
			response := LoginResponse{
				Token: jwtToken,
				IsNewUser: true,
			}
			utils.WriteJSON(w, http.StatusOK, response)
			
		}else {
			log.Println("error3", err)
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

	}
	jwtToken, err := auth.GenerateJWT(user.Email)
	if err != nil {
		log.Println("error6", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := LoginResponse{
		Token: jwtToken,
		IsNewUser: false,	
	}
	utils.WriteJSON(w, http.StatusOK, response)

}

