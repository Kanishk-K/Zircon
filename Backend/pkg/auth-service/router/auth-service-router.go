package router

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/auth-service/services"
)

type AuthRouter struct {
	service services.AuthServiceMethods
}

func NewAuthServiceRouter(as services.AuthServiceMethods) *AuthRouter {
	return &AuthRouter{
		service: as,
	}
}

func (asr *AuthRouter) RegisterRoutes() {
	http.HandleFunc("/login", asr.HandleLogin)
	http.HandleFunc("/callback", asr.HandleCallback)
}

func (asr *AuthRouter) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Handle redirect
		oauthState := asr.service.GenerateStateOAuthCookie(w)
		u := asr.service.GenerateAuthCodeURL(oauthState)
		http.Redirect(w, r, u, http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
	}
}

func (asr *AuthRouter) HandleCallback(w http.ResponseWriter, r *http.Request) {
	profile, err := asr.service.GetUserDataFromGoogle(r.FormValue("code"))
	if err != nil {
		log.Printf("Failed to get user data: %s", err.Error())
		http.Error(w, "Failed to authenticate user, please try again.", http.StatusInternalServerError)
		return
	}
	if profile.OrganizationDomain != "umn.edu" {
		http.Error(w, "You're not registered with a valid \"umn.edu\" email.", http.StatusUnauthorized)
	}
	tokenDetails, err := asr.service.GenerateJWTFromProfile(profile)
	if err != nil {
		log.Printf("Failed to get encode JWT: %s", err.Error())
		http.Error(w, "Failed to encode JWT, please try again.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenDetails)
}
