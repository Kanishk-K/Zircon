package authutil

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/auth-service/models"
	"github.com/golang-jwt/jwt/v5"
)

type AuthClientMethods interface {
	SignJWT(user *models.ProfileData) (*JWTDetails, error)
	VerifyJWT(token string) error
	SecureRoute(w http.ResponseWriter, r *http.Request) error
}

type AuthClient struct {
	jwtPrivateKey []byte
}

func NewAuthClient(jwtPrivateKey []byte) AuthClientMethods {
	return &AuthClient{
		jwtPrivateKey: jwtPrivateKey,
	}
}

type JWTDetails struct {
	TokenString string `json:"token"`
	Email       string `json:"email"`
	Expiry      int64  `json:"expiry"`
}

func (ac *AuthClient) SignJWT(user *models.ProfileData) (*JWTDetails, error) {
	expiryTime := time.Now().Add(time.Hour * 24 * 30).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"exp":   expiryTime, // Default standard as described in RFC7519 (https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.4)
	})
	tokenString, err := token.SignedString(ac.jwtPrivateKey)
	if err != nil {
		return nil, err
	}
	return &JWTDetails{
		TokenString: tokenString,
		Email:       user.Email,
		Expiry:      expiryTime,
	}, nil
}

func (ac *AuthClient) VerifyJWT(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return ac.jwtPrivateKey, nil
	})
	if err != nil {
		log.Printf("Critial error in parsing JWT: %s", err.Error())
		return err
	}
	if !token.Valid {
		return fmt.Errorf("invalid JWT token provided")
	}
	return nil
}

func (ac *AuthClient) SecureRoute(w http.ResponseWriter, r *http.Request) error {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Missing authorization header (JWT)")
		return fmt.Errorf("missing authorization header (JWT)")
	}
	// Ensure to add Bearer to the token for good practice
	tokenString = tokenString[len("Bearer "):]

	err := ac.VerifyJWT(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Invalid Token Provided")
		return fmt.Errorf("invalid token provided")
	}

	return nil
}
