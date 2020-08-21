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

// GQLQuery : format of a GQL request
type GQLQuery struct {
	OperationName string                 `json:"operationName"`
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
}

// Book : GQL Book Type
type Book struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	ISBN     string `json:"isbn"`
	AuthorID int64  `json:"authorID"`
	Author   Author `json:"author"`
}

// Author : GQL Author Type
type Author struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DateOfBirth string `json:"dateOfBirth"`
	Books       []Book `json:"books"`
}
