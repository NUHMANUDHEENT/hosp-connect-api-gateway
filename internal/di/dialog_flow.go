package di

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2/google"
)

var cachedToken string
var tokenExpiryTime time.Time

func GetAccessToken() (string, error) {
	currentTime := time.Now()

	if cachedToken != "" && tokenExpiryTime.After(currentTime) {
		return cachedToken, nil
	}
	fmt.Println("hiiiii")
	// Load the Service Account credentials file
	data, err := os.ReadFile("/home/nuhmanudheen-t/Broto/2ndProject/HospitalConnect/docto-sheduler-1ebb4fc933f8.json")
	if err != nil {
		return "", fmt.Errorf("failed to read service account file: %v", err)
	}

	// Parse the credentials file
	creds, err := google.CredentialsFromJSON(context.Background(), data, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", fmt.Errorf("failed to parse credentials: %v", err)
	}

	// Get the token
	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}

	cachedToken = token.AccessToken
	tokenExpiryTime = time.Now().Add(time.Hour)

	return cachedToken, nil
}
func HelpDeskRender(w http.ResponseWriter, r *http.Request) {
	paymentPagePath := filepath.Join("..", "templates", "help_desk.html")
	http.ServeFile(w, r, paymentPagePath)
}

// Function to handle the chatbot request
func HelpDeskHandler(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Message string `json:"message"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		fmt.Println("err", err)
		return
	}
	log.Println("request message : ", reqBody)
	token, err := GetAccessToken()
	if err != nil {
		http.Error(w, "Error getting access token", http.StatusInternalServerError)
		fmt.Println("errr", err)
		return
	}

	// Build Dialogflow API request
	url := "https://dialogflow.googleapis.com/v2/projects/docto-sheduler/agent/sessions/12345:detectIntent"
	dialogflowRequest := map[string]interface{}{
		"queryInput": map[string]interface{}{
			"text": map[string]string{
				"text":         reqBody.Message,
				"languageCode": "en-US",
			},
		},
	}

	// Make the Dialogflow request
	client := &http.Client{}
	reqBodyJson, _ := json.Marshal(dialogflowRequest)
	req, _ := http.NewRequest("POST", url, io.NopCloser(bytes.NewReader(reqBodyJson)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		http.Error(w, "Error communicating with Dialogflow", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	var dialogflowResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&dialogflowResp)


	fulfillmentText := dialogflowResp["queryResult"].(map[string]interface{})["fulfillmentText"].(string)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"reply": fulfillmentText,
	})
}
