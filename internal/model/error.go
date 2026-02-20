package model

import "fmt"

// APIError represents an error response from the Zenodo API.
type APIError struct {
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Errors  []Detail `json:"errors,omitempty"`
}

// Detail is a single field-level error from the API.
type Detail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("API error %d: %s", e.Status, e.Message)
	for _, d := range e.Errors {
		msg += fmt.Sprintf("\n  - %s: %s", d.Field, d.Message)
	}
	return msg
}

// Hint returns a user-friendly suggestion based on the status code.
func (e *APIError) Hint() string {
	switch e.Status {
	case 400:
		if e.LikelyAuthError() {
			return "This may be caused by an invalid API token. Check your token with: zenodo config set token <TOKEN>"
		}
		return ""
	case 401:
		return "Check that your token is valid. Set it with: zenodo config set token <TOKEN>"
	case 403:
		return "Your token may lack required scopes (deposit:write, deposit:actions). Generate a new token at https://zenodo.org/account/settings/applications/"
	case 404:
		return "Record not found. Check the ID and ensure you have access."
	case 429:
		return "Rate limit exceeded. The CLI will automatically retry; if this persists, reduce request frequency."
	default:
		return ""
	}
}

// LikelyAuthError returns true if a 400 response looks like it was caused by
// an invalid token. Zenodo returns 400 (not 401) on some authenticated endpoints
// when the token is malformed.
func (e *APIError) LikelyAuthError() bool {
	if e.Status != 400 {
		return false
	}
	// Zenodo returns a "validation error" with empty field messages when the
	// token is invalid on /deposit/depositions endpoints.
	for _, d := range e.Errors {
		if d.Field != "" && d.Message == "" {
			return true
		}
	}
	return false
}
