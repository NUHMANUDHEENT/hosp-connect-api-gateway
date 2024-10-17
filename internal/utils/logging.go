package utils

import (
	"log"
	"net/http"
)

// logRequestResponse is a utility function to log request details and response information
func logRequestResponse(statusCode int, message interface{}, requestDetails *http.Request) {
	log.Printf("[REQUEST] %s %s from %s | [RESPONSE] Status: %d, Message: %v",
		requestDetails.Method, requestDetails.URL.Path, requestDetails.RemoteAddr, statusCode, message)
}
