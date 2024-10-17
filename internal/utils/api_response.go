package utils

import (
	"encoding/json"
	"net/http"
)

func JSONResponse(w http.ResponseWriter, message interface{}, statusCode int, requestDetails *http.Request) {
	logRequestResponse(statusCode, message, requestDetails)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(message)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

type StandardResponse struct {
	Status     string `json:"status"`
	Error      string `json:"error"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

// JSONStandardResponse sends a standardized JSON response
func JSONStandardResponse(w http.ResponseWriter, status string, errorStr string, message string, statusCode int, req *http.Request) {

	logRequestResponse(statusCode, message, req)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := StandardResponse{
		Status:     status,
		Error:      errorStr,
		Message:    message,
		StatusCode: statusCode,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
