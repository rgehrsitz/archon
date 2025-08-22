package errors

type Envelope struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func Wrap(code, message string, details interface{}) Envelope {
	return Envelope{Code: code, Message: message, Details: details}
}
