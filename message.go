package main

import (
	"regexp"
)

var rxEmail = regexp.MustCompile(".+@.+\\..+")

// Message will check inputs for errors
// type Message struct {
// 	Email string
// 	// Content string
// 	Errors map[string]string
// }

// Message will check inputs for errors
type Message struct {
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
	Errors       map[string]string
}

// Validate user input
func (msg *Message) Validate() bool {
	msg.Errors = make(map[string]string)

	match := rxEmail.Match([]byte(msg.Email))
	if match == false {
		msg.Errors["Email"] = "Please enter a valid email address"
	}

	// if strings.TrimSpace(msg.Content) == "" {
	// 	msg.Errors["Content"] = "Please enter a message"
	// }

	return len(msg.Errors) == 0
}
