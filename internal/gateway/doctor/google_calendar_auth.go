package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

var (
	googleOauthConfig *oauth2.Config
	oauthStateString  = "random"
)

type GoogleUserInfo struct {
	Email string `json:"email"`
}

// Initialize OAuth config for Google Calendar
func init() {
	fmt.Println(os.Getenv("CLIENT_ID"))
	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/api/v1/doctor/auth/callback",
		Scopes:       []string{calendar.CalendarScope, "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
}

// Doctor clicks login
func (d *DoctorServerClient) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Callback from Google after doctor login
func (d *DoctorServerClient) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Get the user's email from Google User Info API
	email, err := getGoogleUserEmail(token)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Store the token with the email instead of doctor_id
	err = d.DoctorStoreAccessToken(email, token)
	if err != nil {
		http.Error(w, "Failed to store token", http.StatusInternalServerError)
		return
	}

	// Success! You can redirect to the next page or dashboard
	fmt.Fprintf(w, "Google Calendar authenticated successfully for email: %s", email)
}

// Helper function to retrieve the user's email using the OAuth2 token
func getGoogleUserEmail(token *oauth2.Token) (string, error) {
	client := googleOauthConfig.Client(context.Background(), token)

	// Request user info from Google's userinfo endpoint
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse the response
	userInfo, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Unmarshal the response to extract the email
	var googleUser GoogleUserInfo
	err = json.Unmarshal(userInfo, &googleUser)
	if err != nil {
		return "", err
	}

	return googleUser.Email, nil
}
