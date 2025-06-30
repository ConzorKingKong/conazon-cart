package token

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/conzorkingkong/conazon-cart/config"
	authtypes "github.com/conzorkingkong/conazon-users-and-auth/types"
	"github.com/golang-jwt/jwt/v5"
)

type MyJWT struct {
	Id int `json:"id"`
}


func CreateToken(id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512,
		jwt.MapClaims{
			"id":  id,
			"exp": time.Now().Add(time.Hour).Unix(),
		})

	tokenString, err := token.SignedString(config.SECRETKEY)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return config.SECRETKEY, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func ValidateAndReturnSession(w http.ResponseWriter, r *http.Request) (MyJWT, error) {

	cookie, err := r.Cookie("JWTToken")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			log.Printf("cookie not found")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusBadRequest, Message: "cookie not found", Data: ""})
		default:
			log.Printf("Cookie err: %s", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "server error", Data: ""})
		}
		return MyJWT{}, err
	}
	// auth check token

	err = VerifyToken(cookie.Value)
	if err != nil {
		log.Printf("Error verifying token: %s", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusUnauthorized, Message: "invalid token", Data: ""})
		return MyJWT{}, err
	}

	// if yes validate data
	tokenData := strings.Split(cookie.Value, ".")[1]

	value, err := base64.RawStdEncoding.DecodeString(tokenData)
	if err != nil {
		log.Printf("Error decoding token: %s", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(authtypes.Response{Status: http.StatusInternalServerError, Message: "internal service error", Data: ""})
		return MyJWT{}, err
	}

	TokenData := MyJWT{}
	json.Unmarshal(value, &TokenData)

	return TokenData, nil
}