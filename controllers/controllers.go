package controllers

import (
	"am.ca-server/helpers"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mailgun/mailgun-go"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// BaseURL Route
// Route : '/'
// Type  : 'GET'
func BaseURL(w http.ResponseWriter, r *http.Request) {
	// set headers
	w.Header().Set("content-type", "application/json")

	json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 404, Message: "Invalid Route"})
}

// EmailService Route
// Route : '/email
// Type  : 'POST'
func EmailService(w http.ResponseWriter, r *http.Request) {
	// Read body
	var responseEmail helpers.Email
	json.NewDecoder(r.Body).Decode(&responseEmail)

	// set headers
	w.Header().Set("content-type", "application/json")

	// Create an instance of the Mailgun Client
	mg := mailgun.NewMailgun(os.Getenv("MAILGUN_DOMAIN"), os.Getenv("MAILGUN_API_KEY"))

	// Create and format mailgun email
	messageSubject := fmt.Sprintf("[alexmontague.ca] - %s", responseEmail.Subject)
	messageBody := fmt.Sprintf("Sent By: %s\n\nSender Email: %s\n\n%s\n", responseEmail.Sender, responseEmail.FromEmail, responseEmail.Message)
	message := mg.NewMessage("info@bookbuy.ca", messageSubject, messageBody, responseEmail.ToEmail)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message
	resp, id, err := mg.Send(ctx, message)
	if err != nil {
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 401, Message: "Something went wrong. Make sure you have all parameters!"})
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("ID: %s Resp: %s\n", id, resp)

	json.NewEncoder(w).Encode(helpers.Response{Error: false, Code: 200, Message: "Email Received"})
}

// ResumeJSON Route
// Route : '/resume
// Type  : 'GET'
func ResumeJSON(w http.ResponseWriter, r *http.Request) {
	// set headers
	w.Header().Set("content-type", "application/json")

	// open resume file
	absPath, _ := filepath.Abs("assets/resume.json")
	resumeFile, err := os.Open(absPath)

	if err != nil {
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 401, Message: "Something went wrong parsing the resume file. Sorry!"})
		fmt.Println(err.Error())
		return
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer resumeFile.Close()

	// read file in
	byteValue, _ := ioutil.ReadAll(resumeFile)

	// parse and return file as JSON
	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)
	json.NewEncoder(w).Encode(result)
}
