package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/conzorkingkong/conazon-cart/config"
	"github.com/conzorkingkong/conazon-cart/controllers"
	authcontrollers "github.com/conzorkingkong/conazon-users-and-auth/controllers"
	"github.com/joho/godotenv"
)

var PORT, PORTExists = "", false
var JwtSecret, jwtSecretExists = "", false
var DatabaseURLEnv, DatabaseURLExists = "", false

func main() {

	godotenv.Load()

	PORT, PORTExists = os.LookupEnv("PORT")
	JwtSecret, jwtSecretExists = os.LookupEnv("JWTSECRET")
	DatabaseURLEnv, DatabaseURLExists = os.LookupEnv("DATABASEURL")

	config.SECRETKEY = []byte(JwtSecret)
	config.DatabaseURLEnv = DatabaseURLEnv

	if !jwtSecretExists || !DatabaseURLExists {
		log.Fatal("Required environment variable not set")
	}

	if !PORTExists {
		PORT = "8082"
	}

	http.HandleFunc("/", authcontrollers.Root)

	http.HandleFunc("/cart/", controllers.CartHandler)
	http.HandleFunc("/cart/{id}", controllers.CartId)
	http.HandleFunc("/cart/user/{id}", controllers.UserId)

	http.HandleFunc("/healthz", authcontrollers.Healthz)

	fmt.Println("server starting on port", PORT)
	http.ListenAndServe(":"+PORT, nil)
}
