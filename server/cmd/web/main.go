package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

type App struct {
	// abstract class ? methods
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("ERROR: CANNOT LOAD ENVs %v", err)
	}

	app := &App{}
	server := &http.Server{
		Addr:    ":8080",
		Handler: app.Router(),
	}

	log.Print("SERVER LISTENING")
	server.ListenAndServe()
}
