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
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
	"gopkg.in/gomail.v2"
)

type templateTags struct {
	Name  string
	ZQR   string
	ZLink string
	ZMeet string
	ZPass string
	ZDate string
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

func retrieveSecrets() (secretsQuery []secrets) {
	ctx := context.Background()

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	q := datastore.NewQuery("Secrets").
		Filter("ID =", 1).
		Limit(1)

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

func createMessage(name string) (message string) {
	secrets := retrieveSecrets()
	zqr := secrets[0].QR
	zlink := secrets[0].Link
	zmeet := secrets[0].Meet
	zpass := secrets[0].Pass
	zdate := secrets[0].Date
	var tags = templateTags{name, zqr, zlink, zmeet, zpass, zdate}
	emailBody := template.New("emailtemplate.html")

	emailBody, err := emailBody.ParseFiles("emailtemplate.html")
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

// RegistrationHandler adds new record in the database
func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("form.html"))

	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}

	details := Entry{
		Email:     r.FormValue("email"),
		FirstName: r.FormValue("fname"),
		LastName:  r.FormValue("lname"),
		Area:      r.FormValue("area"),
		Group:     r.FormValue("group"),
		Function:  r.FormValue("function"),
		Gender:    r.FormValue("gender"),
		Local:     r.FormValue("local"),
		District:  r.FormValue("district"),
		Available: r.FormValue("available"),
	}

	recorded := record(details)

	if recorded {
		createdMessage := createMessage(details.FirstName)
		sentEmail := sendEmail(details.Email, createdMessage)
		if sentEmail {
			tmpl.Execute(w, struct{ Success bool }{true})
		} else {
			tmpl.Execute(w, struct{ EmailFailed bool }{true})
		}
	} else {
		tmpl.Execute(w, struct{ RecordFailed bool }{true})
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
	// tmpl := template.Must(template.ParseFiles("records.html"))
	// records := retrieveRecords()
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
