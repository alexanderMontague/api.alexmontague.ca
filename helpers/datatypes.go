package datatypes

// Response : format of all responses coming back from server
type Response struct {
	Error   bool   `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}
