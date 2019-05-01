package helpers

// Response : format of all responses coming back from server
type Response struct {
	Error   bool   `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	// Data    string `json:"data"`
}

// Email : format of email requests coming in from website form
type Email struct {
	Sender    string `json:"sender"`
	ToEmail   string `json:"toEmail"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
	FromEmail string `json:"fromEmail"`
}
