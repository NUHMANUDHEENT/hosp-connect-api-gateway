package patient

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/NUHMANUDHEENT/hosp-connect-pb/proto/patient"
	"github.com/gorilla/websocket"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/di"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/utils"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/middleware"
)

// PatientSignUp handles the patient signup via gRPC
func (p *PatientServerClient) PatientSignUp(w http.ResponseWriter, req *http.Request) {
	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Phone    int32  `json:"phone"`
		Age      int32  `json:"age"`
		Gender   string `json:"gender"`
	}
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		utils.JSONStandardResponse(w, "error", "Invalid request format", "", http.StatusBadRequest, req)
		return
	}

	// Call the gRPC service
	resp, err := p.SignUp(context.Background(), &patient.SignUpRequest{
		Email:    reqBody.Email,
		Password: reqBody.Password,
		Name:     reqBody.Name,
		Phone:    reqBody.Phone,
		Age:      reqBody.Age,
		Gender:   reqBody.Gender,
	})
	if err != nil {
		utils.JSONStandardResponse(w, "error", "GRPC error", "", http.StatusInternalServerError, req)
		return
	}

	// Return a success response
	utils.JSONResponse(w, resp, int(resp.StatusCode), req)
}
func (p *PatientServerClient) SignUpverify(w http.ResponseWriter, req *http.Request) {
	token := req.URL.Query().Get("token")
	if token == "" {
		utils.JSONResponse(w, "Invalid token", http.StatusBadRequest, req)
		return
	}
	resp, err := p.SignUpVerify(context.Background(), &patient.SignUpVerifyRequest{
		Token: token,
	})
	if err != nil {
		utils.JSONResponse(w, resp, 500, req)
	}
	utils.JSONResponse(w, resp, 200, req)

}

// PatientSignIn handles the patient sign-in via gRPC
func (p *PatientServerClient) PatientSignIn(w http.ResponseWriter, req *http.Request) {
	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		utils.JSONStandardResponse(w, "error", "Invalid request format", "", http.StatusBadRequest, req)
		return
	}

	// Call the gRPC service
	resp, err := p.SignIn(context.Background(), &patient.SignInRequest{
		Email:    reqBody.Email,
		Password: reqBody.Password,
	})
	if err != nil {
		utils.JSONStandardResponse(w, "error", "User GRPC error", "", http.StatusInternalServerError, req)
		return
	}
	fmt.Println("patient_id", resp.PatientId)
	if resp.Status == "success" {
		// Create JWT token
		jwtToken, err := middleware.CreateJWTToken(resp.PatientId, "patient")
		if err != nil {
			utils.JSONStandardResponse(w, "error", "Failed to create JWT token", "", http.StatusInternalServerError, req)
			return
		}

		middleware.SetJWTToken(w, jwtToken, "patient")
	}

	// Return a success response with JWT token
	utils.JSONResponse(w, resp, int(resp.StatusCode), req)
}

func (p *PatientServerClient) PatientLogout(w http.ResponseWriter, req *http.Request) {
	http.SetCookie(w, &http.Cookie{
		MaxAge:   -1,
		Name:     "patienttoken",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	utils.JSONStandardResponse(w, "success", "", "Patient logged out successfully", http.StatusOK, req)
}
func (p *PatientServerClient) GetPatientProfile(w http.ResponseWriter, req *http.Request) {
	claims, err := middleware.ExtractClaimsFromCookie(req, "patient")
	if err != nil {
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, req)
		return
	}

	resp, err := p.GetProfile(context.Background(), &patient.GetProfileRequest{
		PatientId: claims.UserId,
	})
	if err != nil || resp.Status != "success" {
		utils.JSONResponse(w, "Failed to get profile", http.StatusInternalServerError, req)
		return
	}

	utils.JSONResponse(w, resp, http.StatusOK, req)
}

func (p *PatientServerClient) UpdatePatientProfile(w http.ResponseWriter, req *http.Request) {
	var reqBody patient.Patient
	claims, err := middleware.ExtractClaimsFromCookie(req, "patient")
	if err != nil {
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, req)
		return
	}

	err = json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := p.UpdateProfile(context.Background(), &patient.UpdateProfileRequest{
		Patient: &patient.Patient{
			PatientId: claims.UserId,
			Name:      reqBody.Name,
			Email:     reqBody.Email,
			Phone:     reqBody.Phone,
			Age:       int32(reqBody.Age),
			Gender:    reqBody.Gender,
		},
	})
	if err != nil {
		utils.JSONResponse(w, err.Error(), http.StatusInternalServerError, req)
		return
	}

	utils.JSONResponse(w, resp, http.StatusOK, req)
}
func (d *PatientServerClient) AddPrescriptionForPatient(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		PatientId    string `json:"patientid"`
		Prescription []struct {
			Medication string `json:"medication"`
			Dosage     string `json:"dosage"`
			Frequency  string `json:"frequency"`
		} `json:"prescription"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		utils.JSONStandardResponse(w, "fail", "failed to bind json", "", 400, r)
		return
	}
	claims, err := middleware.ExtractClaimsFromCookie(r, "doctor")
	if err != nil {
		utils.JSONStandardResponse(w, "fail", "Unauthorized", "", 401, r)
		return
	}
	doctorId := claims.UserId
	if doctorId == "" {
		utils.JSONStandardResponse(w, "fail", "Unauthorized", "", 401, r)
		return
	}
	var prescription []*patient.Prescription
	for _, v := range reqBody.Prescription {
		prescription = append(prescription, &patient.Prescription{
			Medication: v.Medication,
			Dosage:     v.Dosage,
			Frequency:  v.Frequency,
		})
	}
	resp, err := d.AddPrescription(context.Background(), &patient.AddPrescriptionRequest{
		PatientId:    reqBody.PatientId,
		DoctorId:     doctorId,
		Prescription: prescription,
	})
	if err != nil {
		utils.JSONResponse(w, err.Error(), http.StatusInternalServerError, r)
		return
	}
	utils.JSONResponse(w, resp, http.StatusOK, r)

}
func (p *PatientServerClient) GetPrescriptions(w http.ResponseWriter, r *http.Request) {
	val := r.URL.Query()
	query := val.Get("query")
	fmt.Println("q", query)
	claims, err := middleware.ExtractClaimsFromCookie(r, "patient")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := p.GetPrescription(context.Background(), &patient.GetPrescriptionRequest{
		PatientId: claims.UserId,
		Query:     query,
	})
	if err != nil {
		utils.JSONResponse(w, resp, http.StatusInternalServerError, r)
		return
	}
	utils.JSONResponse(w, resp, http.StatusOK, r)

}
func (d *PatientServerClient) VideoCallRender(w http.ResponseWriter, r *http.Request) {
	videocallhtml := filepath.Join("..", "templates", "video_call_jitsi.html")
	http.ServeFile(w, r, videocallhtml)
}

func (p *PatientServerClient) PatientChatHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := di.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading patient connection:", err)
		return
	}
	di.P1atientConnections[conn] = true
	defer conn.Close()

	for {
		var message di.Message
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println("Error reading message from patient:", err)
			delete(di.P1atientConnections, conn)
			break
		}
		log.Printf("Received message from patient: %s", message)

		sendMessageToCustomerCare(message)
	}
}

func sendMessageToCustomerCare(msg di.Message) {
	messageJSON, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshalling message to JSON:", err)
		return
	}
	for conn := range di.CustomerConnections {
		err := conn.WriteMessage(websocket.TextMessage, messageJSON)
		if err != nil {
			log.Println("Error sending message to customer care:", err)
			conn.Close()
			delete(di.CustomerConnections, conn)
		}
	}
}
func (p *PatientServerClient) PatientChatRender(w http.ResponseWriter, r *http.Request) {
	paymentPagePath := filepath.Join("..", "templates", "user_chat.html")
	http.ServeFile(w, r, paymentPagePath)
}
