package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// Entry is the schema of records
type Entry struct {
	Email        string
	FirstName    string
	LastName     string
	Area         string
	Group        string
	Function     string
	Gender       string
	Local        string
	District     string
	Status       string
	PreferredDay string
	Available    string //Deprecated property; to-be deleted
}

func main() {
	// Doing the seeding out of habit ;)
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/internal/register", RegistrationHandler)
	http.HandleFunc("/internal/retrieve", RetrievalHandler)
	http.HandleFunc("/confirmation", confirmation)
	http.HandleFunc("/failed", failed)

	// Health and miscellaneous APIs
	http.HandleFunc("/status", StatusHandler)
	http.HandleFunc("/ping", PingHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
