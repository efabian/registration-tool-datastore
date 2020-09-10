package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
	"gopkg.in/gomail.v2"
)

type templateTags struct {
	FName    string
	LName    string
	District string
	Time     string
	ZQR      string
	ZLink    string
	ZMeet    string
	ZPass    string
	ZDate    string
}

type secrets struct {
	ID      int    `datastore:"ID"`
	API     string `datastore:"api"`
	Key     string `datastore:"key"`
	SMTP    string `datastore:"smtp"`
	Sender  string `datastore:"sender"`
	BCC     string `datastore:"bcc"`
	BCCnick string `datastore:"bccnick"`
	Subject string `datastore:"subject"`
	Link    string `datastore:"zlink"`
	Meet    string `datastore:"zmeet"`
	Pass    string `datastore:"zpass"`
	Date    string `datastore:"zdate"`
	QR      string `datastore:"zqr"`
}

const projectID string = "hk-thai-kadiwa"

// TimeSalt for the Ping Handler
const TimeSalt string = "LMjKASwwzUvFQwtr8jmFrjKXeBQQ3LzC"

func render(w http.ResponseWriter, filename string, data interface{}) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		log.Println(err)
		http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Println(err)
		http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
	}
}

func retrieveSecrets() (secretsQuery []secrets) {
	ctx := context.Background()

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	q := datastore.NewQuery("Secrets").
		Filter("ID <", 7).
		Limit(6)

	if _, err := client.GetAll(ctx, q, &secretsQuery); err != nil {
		log.Fatalf("Failed to retrieve secrets: %v", err)
	}
	return secretsQuery
}

func record(details Entry) (ok bool) {
	ctx := context.Background()

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	kind := "Registration"
	key := details.Email
	recordKey := datastore.NameKey(kind, key, nil)
	record := details

	if _, err := client.Put(ctx, recordKey, &record); err == nil {
		ok = true
	} else {
		log.Fatalf("Failed to save entry: %v", err)
		ok = false
	}
	return ok
}

func createMessage(fname string, lname string, district string, day string) (message string) {
	secrets := retrieveSecrets()
	var i int
	if day == "mon" {
		i = 0
	}
	if day == "tue" {
		i = 1
	}
	if day == "wed" {
		i = 2
	}
	if day == "thu" {
		i = 3
	}
	if day == "fri" {
		i = 4
	}
	if day == "sat" {
		i = 5
	}

	zqr := secrets[i].QR
	zlink := secrets[i].Link
	zmeet := secrets[i].Meet
	zpass := secrets[i].Pass
	zdate := secrets[i].Date

	var time string
	if district == "hk" {
		time = "9:50PM"
	} else {
		time = "8:50PM"
	}

	var tags = templateTags{fname, lname, strings.ToUpper(district), time, zqr, zlink, zmeet, zpass, zdate}
	emailBody := template.New("emailtemplate.html")

	emailBody, err := emailBody.ParseFiles("templates/emailtemplate.html")
	if err != nil {
		log.Println(err)
	}

	//Declare template as buffer of bytes
	var tpl bytes.Buffer
	if err := emailBody.Execute(&tpl, tags); err != nil {
		log.Println(err)
	}

	return tpl.String()
}

func sendEmail(email string, message string) (ok bool) {
	secrets := retrieveSecrets()
	smtpServ := secrets[0].SMTP
	smtpPort := 587

	sesAPI := secrets[0].API
	sesKey := secrets[0].Key

	sender := secrets[0].Sender
	bcc := secrets[0].BCC
	bccNickname := secrets[0].BCCnick
	subject := secrets[0].Subject

	mailParam := gomail.NewMessage()
	mailParam.SetHeader("From", sender)
	mailParam.SetHeader("To", email)
	mailParam.SetAddressHeader("Bcc", bcc, bccNickname)
	mailParam.SetHeader("Subject", subject)
	mailParam.SetBody("text/html", message)

	send := gomail.NewDialer(smtpServ, smtpPort, sesAPI, sesKey)

	if err := send.DialAndSend(mailParam); err != nil {
		ok = false
	} else {
		ok = true
	}
	return ok
}

func confirmation(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/confirmation.html", nil)
}

func registrationfailure(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/registrationfailure.html", nil)
}

func sendingfailure(w http.ResponseWriter, r *http.Request) {
	render(w, "templates/sendingfailure.html", nil)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		render(w, "templates/form.html", nil)
	}
}

func checkSize(day string) (size int) {
	ctx := context.Background()

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	q := datastore.NewQuery("Registration").Filter("PreferredDay =", day)

	if size, err = client.Count(ctx, q); err != nil {
		log.Fatalf("Failed to retrieve records: %v", err)
	}
	log.Println(size)
	return size
}

// RegistrationHandler adds new record in the database
func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		getHandler(w, r)
		return
	}

	msg := &Inputs{
		Email:        strings.ToLower(r.FormValue("email")),
		FirstName:    r.PostFormValue("fname"),
		LastName:     r.PostFormValue("lname"),
		Area:         r.PostFormValue("area"),
		Group:        r.PostFormValue("group"),
		Function:     r.PostFormValue("function"),
		Gender:       r.PostFormValue("gender"),
		Local:        r.PostFormValue("local"),
		District:     r.PostFormValue("district"),
		Status:       r.PostFormValue("status"),
		PreferredDay: r.PostFormValue("prefday"),
	}

	size := checkSize(msg.PreferredDay)

	if size > 12 {
		msg.PreferredDay = "overbooked"
	}

	if msg.Validate() == false {
		render(w, "templates/form.html", msg)
		return
	}

	validatedInputs := Entry{
		Email:        strings.ToLower(r.FormValue("email")),
		FirstName:    r.FormValue("fname"),
		LastName:     r.FormValue("lname"),
		Area:         r.FormValue("area"),
		Group:        r.FormValue("group"),
		Function:     r.FormValue("function"),
		Gender:       r.FormValue("gender"),
		Local:        r.FormValue("local"),
		District:     r.FormValue("district"),
		Status:       r.FormValue("status"),
		PreferredDay: r.FormValue("prefday"),
	}

	recorded := record(validatedInputs)

	if recorded {
		createdMessage := createMessage(validatedInputs.FirstName, validatedInputs.LastName, validatedInputs.District, validatedInputs.PreferredDay)
		sentEmail := sendEmail(validatedInputs.Email, createdMessage)

		if sentEmail {
			http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/sendingfailure", http.StatusSeeOther)
		}
	} else {
		http.Redirect(w, r, "/recordfailure", http.StatusSeeOther)
	}
}

func retrieveRecords() (entriesQuery []Entry) {
	ctx := context.Background()

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	q := datastore.NewQuery("Registration")

	if _, err := client.GetAll(ctx, q, &entriesQuery); err != nil {
		log.Fatalf("Failed to retrieve records: %v", err)
	}
	return entriesQuery
}

// RetrievalHandler retrieves the records from the database
func RetrievalHandler(w http.ResponseWriter, r *http.Request) {
	// To-Do: Add a security
	// tmpl := template.Must(template.ParseFiles("templates/records.html"))
	// records := retrieveRecords()
	// log.Println(records[0].Email)
	// if records.PreferredDay = "wed" {
	// 	records.
	// }
	// err := tmpl.Execute(w, records)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
}

// Tester for a new handler
func Tester(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	if u == nil {
		url, _ := user.LoginURL(ctx, "http://localhost:8080/internal/retrieve")
		fmt.Fprintf(w, `<a href="%s">Sign in or register</a>`, url)
		return
	}
	url, _ := user.LogoutURL(ctx, "/internal/retrieve")
	fmt.Fprintf(w, `Welcome, %s! (<a href="%s">sign out</a>)`, u, url)
}

// StatusHandler provides basic health check
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Alive and well :)")
}

// PingHandler provides basic Health check and timestamp
func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain ")
	now := strconv.FormatInt(time.Now().Unix(), 10)
	w.Write([]byte(now + fmt.Sprintf("%x", md5.Sum([]byte(now+TimeSalt)))))
}
