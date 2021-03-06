package main

import (
	"regexp"
	"strings"
)

var rxEmail = regexp.MustCompile(".+@.+\\..+")
var rxDigit = regexp.MustCompile(`\d`)

// Message will check inputs for errors
// type Message struct {
// 	Email string
// 	// Content string
// 	Errors map[string]string
// }

// Inputs will be validated for errors
type Inputs struct {
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
func (msg *Inputs) Validate() bool {
	msg.Errors = make(map[string]string)

	match := rxEmail.Match([]byte(msg.Email))
	matchArea := rxDigit.FindString(string(msg.Area))
	matchGroup := rxDigit.FindString(string(msg.Group))

	if match == false {
		msg.Errors["Email"] = "Please enter a valid email address"
	}

	if strings.TrimSpace(msg.FirstName) == "" {
		msg.Errors["FirstName"] = "Please enter your first name"
	}

	if strings.TrimSpace(msg.LastName) == "" {
		msg.Errors["LastName"] = "Please enter your family name"
	}

	// if strings.TrimSpace(msg.Area) == "" {
	if matchArea == "" {
		msg.Errors["Area"] = "Please enter your area (ex: 1)"
	}

	// if strings.TrimSpace(msg.Group) == "" {
	if matchGroup == "" {
		msg.Errors["Group"] = "Please enter your group (ex: 1)"
	}

	if strings.TrimSpace(msg.Gender) == "" {
		msg.Errors["Gender"] = "Please indicate your gender"
	}

	if strings.TrimSpace(msg.Local) == "" {
		msg.Errors["Local"] = "Please indicate your local"
	}

	if strings.TrimSpace(msg.District) == "" {
		msg.Errors["District"] = "Please indicate your district"
	}

	if strings.TrimSpace(msg.Status) == "" {
		msg.Errors["Status"] = "Please indicate your status"
	}

	if strings.TrimSpace(msg.PreferredDay) == "" {
		msg.Errors["PreferredDay"] = "Please pick your preferred day & time"
	}

	if strings.TrimSpace(msg.PreferredDay) == "overbooked" {
		msg.Errors["PreferredDay"] = "It's overbooked. Please pick another day & time"
	}

	return len(msg.Errors) == 0
}
