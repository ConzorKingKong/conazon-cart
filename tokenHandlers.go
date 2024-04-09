package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func verifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return SECRETKEY, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func validateAndReturnSession(w http.ResponseWriter, r *http.Request) (MyJWT, error) {

	cookie, err := r.Cookie("JWTToken")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			log.Printf("cookie not found")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusBadRequest, Message: "cookie not found"})
		default:
			log.Printf("Cookie err: %s", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusInternalServerError, Message: "server error"})
		}
		return MyJWT{}, err
	}
	// auth check token

	err = verifyToken(cookie.Value)
	if err != nil {
		log.Printf("Error verifying token: %s", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusUnauthorized, Message: "invalid token"})
		return MyJWT{}, err
	}

	// if yes validate data
	tokenData := strings.Split(cookie.Value, ".")[1]

	value, err := base64.RawStdEncoding.DecodeString(tokenData)
	if err != nil {
		log.Printf("Error decoding token: %s", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GenericResponse{Status: http.StatusInternalServerError, Message: "internal service error"})
		return MyJWT{}, err
	}

	TokenData := MyJWT{}
	json.Unmarshal(value, &TokenData)

	return TokenData, nil
}
