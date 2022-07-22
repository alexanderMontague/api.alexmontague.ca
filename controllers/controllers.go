package controllers

import (
	"api.alexmontague.ca/graphql"
	"api.alexmontague.ca/helpers"
	"encoding/json"
	"fmt"
	GQL "github.com/graphql-go/graphql"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"net/smtp"
)

// BaseURL Controller
// Route : '/'
// Type  : 'GET'
func BaseURL(w http.ResponseWriter, r *http.Request) {
	// set headers
	w.Header().Set("content-type", "application/json")

	json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 404, Message: "Invalid Route"})
}

// EmailService Controller
// Route : '/email
// Type  : 'POST'
func EmailService(w http.ResponseWriter, r *http.Request) {
	// Read body
	var responseEmail helpers.Email
	json.NewDecoder(r.Body).Decode(&responseEmail)

	// set headers
	w.Header().Set("content-type", "application/json")

	// Configuration
	from := "me@alexmontague.ca"
	password := os.Getenv("ZOHO_EMAIL_SECRET")
	to := []string{"business@alexmontague.ca"}
	smtpHost := "smtp.zoho.com"
	smtpPort := "587"

	messageSubject := fmt.Sprintf("[alexmontague.ca] - %s", responseEmail.Subject)
	messageBody := fmt.Sprintf("Sent By: %s\n\nSender Email: %s\n\n%s\n", responseEmail.Sender, responseEmail.FromEmail, responseEmail.Message)

	msg := []byte("To: " + to[0] + "\r\n" +
		"Subject: " + messageSubject + "\r\n" +
		"\r\n" +
		messageBody + "\r\n")

	// Create authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send actual message
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 401, Message: "Something went wrong sending the email."})
		fmt.Println(err)
		return
	}

	json.NewEncoder(w).Encode(helpers.Response{Error: false, Code: 200, Message: "Email Received"})
}

// ResumeJSON Controller
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

// GraphQL Controller
// Route : '/graphql
// Type  : 'POST'
func GraphQL(w http.ResponseWriter, r *http.Request) {
	// Set headers
	w.Header().Set("content-type", "application/json")

	// Read request query
	var request helpers.GQLQuery
	json.NewDecoder(r.Body).Decode(&request)

	fmt.Println("QUERY -", request.Query)

	params := GQL.Params{
		Schema:         graphql.GetSchema(),
		OperationName:  request.OperationName,
		RequestString:  request.Query,
		VariableValues: request.Variables,
	}
	resolvedValue := GQL.Do(params)
	if len(resolvedValue.Errors) > 0 {
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 400, Message: fmt.Sprintf("%s", resolvedValue.Errors)})
		return
	}

	json.NewEncoder(w).Encode(resolvedValue)
}

// CorsAnywhere Controller
// Route : '/cors
// Type  : 'GET'
// Description : Proxies requests appended to the route to bypass CORS issues
func CorsAnywhere(w http.ResponseWriter, r *http.Request) {
	// set headers
	w.Header().Set("content-type", "application/json")

	// grab url to proxy
	url := r.URL.Query().Get("url")

	if url == "" {
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 400, Message: "A URL was not specified"})
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 400, Message: "Something went wrong with the request"})
		fmt.Println(err.Error())
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 400, Message: "Something went wrong parsing the request"})
		fmt.Println(err.Error())
	}

	// return proxied request body
	var result map[string]interface{}
	json.Unmarshal([]byte(body), &result)
	json.NewEncoder(w).Encode(result)
}