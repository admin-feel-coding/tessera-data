package httpx

import "net/http"

type errorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

// WriteError writes a canonical JSON error response.
func WriteError(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	if details == nil {
		details = map[string]any{}
	}
	WriteJSON(w, status, errorResponse{
		Error: errorBody{Code: code, Message: message, Details: details},
	})
}

// WriteNotImplemented writes a 501 NOT_IMPLEMENTED error.
func WriteNotImplemented(w http.ResponseWriter) {
	WriteError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "This endpoint is not yet implemented.", nil)
}
