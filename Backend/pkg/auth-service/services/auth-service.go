package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/auth-service/models"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/shared/authutil"
	"golang.org/x/oauth2"
)

type AuthServiceMethods interface {
	GenerateStateOAuthCookie(w http.ResponseWriter) string
	GenerateAuthCodeURL(oauthState string) string
	GetUserDataFromGoogle(codeValue string) (*models.ProfileData, error)
	GenerateJWTFromProfile(profile *models.ProfileData) (*authutil.JWTDetails, error)
}

type AuthService struct {
	googleOauthConfig *oauth2.Config
	jwtClient         authutil.AuthClientMethods
}

func NewAuthService(googleOauthConfig *oauth2.Config, jwtClient authutil.AuthClientMethods) AuthServiceMethods {
	return &AuthService{
		googleOauthConfig: googleOauthConfig,
		jwtClient:         jwtClient,
	}
}

func (as *AuthService) GenerateStateOAuthCookie(w http.ResponseWriter) string {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	return state
}

func (as *AuthService) GenerateAuthCodeURL(oauthState string) string {
	return as.googleOauthConfig.AuthCodeURL(oauthState, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func (as *AuthService) GetUserDataFromGoogle(codeValue string) (*models.ProfileData, error) {
	token, err := as.googleOauthConfig.Exchange(context.Background(), codeValue)
	if err != nil {
		return nil, fmt.Errorf("code exchange went wrong: %s", err.Error())
	}

	client := as.googleOauthConfig.Client(context.Background(), token)

	response, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed getting info for user: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from API: %s", err.Error())
	}

	responseData := models.ProfileData{}
	if err = json.Unmarshal(contents, &responseData); err != nil {
		return nil, fmt.Errorf("failed to parse response from response: %s", err.Error())
	}

	return &responseData, nil
}

func (as *AuthService) GenerateJWTFromProfile(profile *models.ProfileData) (*authutil.JWTDetails, error) {
	return as.jwtClient.SignJWT(profile)
}
