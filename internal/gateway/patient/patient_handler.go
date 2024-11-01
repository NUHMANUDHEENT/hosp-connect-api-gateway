package patient

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/NUHMANUDHEENT/hosp-connect-pb/proto/patient"
	"github.com/gorilla/websocket"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/di"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/utils"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/middleware"
	"github.com/sirupsen/logrus"
)

// PatientSignUp handles the patient signup via gRPC
func (p *PatientServerClient) PatientSignUp(w http.ResponseWriter, req *http.Request) {
	p.Logger.Info("Received patient signup request")

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
		p.Logger.WithFields(logrus.Fields{
			"function": "PatientSignUp",
			"error":    err.Error(),
		}).Error("Failed to decode request body")
		utils.JSONStandardResponse(w, "error", "Invalid request format", "", http.StatusBadRequest, req)
		return
	}

	// Call the gRPC service
	resp, err := p.PatientClient.SignUp(context.Background(), &patient.SignUpRequest{
		Email:    reqBody.Email,
		Password: reqBody.Password,
		Name:     reqBody.Name,
		Phone:    reqBody.Phone,
		Age:      reqBody.Age,
		Gender:   reqBody.Gender,
	})
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "PatientSignUp",
			"email":    reqBody.Email,
			"error":    err.Error(),
		}).Error("gRPC error during patient signup")
		utils.JSONStandardResponse(w, "error", "GRPC error", "", http.StatusInternalServerError, req)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function":  "PatientSignUp",
		"patientId": resp.Message,
		"status":    resp.Status,
	}).Info("Patient signed up successfully")

	// Return a success response
	utils.JSONResponse(w, resp, int(resp.StatusCode), req)
	p.Logger.Info("Signup response sent")
}

// SignUpVerify verifies the patient signup token
func (p *PatientServerClient) SignUpVerify(w http.ResponseWriter, req *http.Request) {
	p.Logger.Info("Received signup verification request")

	token := req.URL.Query().Get("token")
	if token == "" {
		p.Logger.Warn("Invalid token provided for signup verification")
		utils.JSONResponse(w, "Invalid token", http.StatusBadRequest, req)
		return
	}

	resp, err := p.PatientClient.SignUpVerify(context.Background(), &patient.SignUpVerifyRequest{
		Token: token,
	})
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "SignUpVerify",
			"token":    token,
			"error":    err.Error(),
		}).Error("gRPC error during signup verification")
		utils.JSONResponse(w, "Failed to verify token", http.StatusInternalServerError, req)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function": "SignUpVerify",
		"status":   resp.Status,
	}).Info("Signup verification successful")
	utils.JSONResponse(w, resp, 200, req)
}

// PatientSignIn handles the patient sign-in via gRPC
func (p *PatientServerClient) PatientSignIn(w http.ResponseWriter, req *http.Request) {
	p.Logger.Info("Received patient sign-in request")

	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "PatientSignIn",
			"error":    err.Error(),
		}).Error("Failed to decode request body")
		utils.JSONStandardResponse(w, "error", "Invalid request format", "", http.StatusBadRequest, req)
		return
	}

	// Call the gRPC service
	resp, err := p.PatientClient.SignIn(context.Background(), &patient.SignInRequest{
		Email:    reqBody.Email,
		Password: reqBody.Password,
	})
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "PatientSignIn",
			"email":    reqBody.Email,
			"error":    err.Error(),
		}).Error("gRPC error during patient sign-in")
		utils.JSONStandardResponse(w, "error", "User GRPC error", "", http.StatusInternalServerError, req)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function":  "PatientSignIn",
		"patientId": resp.PatientId,
		"status":    resp.Status,
	}).Info("Patient signed in successfully")

	if resp.Status == "success" {
		// Create JWT token
		jwtToken, err := middleware.CreateJWTToken(resp.PatientId, "patient")
		if err != nil {
			p.Logger.WithFields(logrus.Fields{
				"function":  "PatientSignIn",
				"patientId": resp.PatientId,
				"error":     err.Error(),
			}).Error("Failed to create JWT token")
			utils.JSONStandardResponse(w, "error", "Failed to create JWT token", "", http.StatusInternalServerError, req)
			return
		}

		// Set JWT token in cookie
		middleware.SetJWTToken(w, jwtToken, "patient")
		p.Logger.Info("JWT token set in cookie for patient")
	}

	// Return a success response with JWT token
	utils.JSONResponse(w, resp, int(resp.StatusCode), req)
	p.Logger.Info("Sign-in response sent")
}

// PatientLogout handles patient logout
func (p *PatientServerClient) PatientLogout(w http.ResponseWriter, req *http.Request) {
	p.Logger.Info("Received patient logout request")

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
	p.Logger.Info("Patient logged out successfully")
}

// GetPatientProfile retrieves the patient profile
func (p *PatientServerClient) GetPatientProfile(w http.ResponseWriter, req *http.Request) {
	p.Logger.Info("Received request to get patient profile")

	claims, err := middleware.ExtractClaimsFromCookie(req, "patient")
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "GetPatientProfile",
			"error":    err.Error(),
		}).Error("Unauthorized access")
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, req)
		return
	}

	resp, err := p.PatientClient.GetProfile(context.Background(), &patient.GetProfileRequest{
		PatientId: claims.UserId,
	})
	if err != nil || resp.Status != "success" {
		p.Logger.WithFields(logrus.Fields{
			"function": "GetPatientProfile",
			"error":    err.Error(),
		}).Error("Failed to get profile")
		utils.JSONResponse(w, "Failed to get profile", http.StatusInternalServerError, req)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function":  "GetPatientProfile",
		"patientId": claims.UserId,
	}).Info("Successfully retrieved patient profile")
	utils.JSONResponse(w, resp, http.StatusOK, req)
}

func (p *PatientServerClient) UpdatePatientProfile(w http.ResponseWriter, req *http.Request) {
	p.Logger.Info("Received request to update patient profile")

	var reqBody patient.Patient
	claims, err := middleware.ExtractClaimsFromCookie(req, "patient")
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "UpdatePatientProfile",
			"error":    err.Error(),
		}).Error("Unauthorized access")
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, req)
		return
	}

	err = json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "UpdatePatientProfile",
			"error":    err.Error(),
		}).Error("Failed to decode request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := p.PatientClient.UpdateProfile(context.Background(), &patient.UpdateProfileRequest{
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
		p.Logger.WithFields(logrus.Fields{
			"function":  "UpdatePatientProfile",
			"patientId": claims.UserId,
			"error":     err.Error(),
		}).Error("gRPC error while updating profile")
		utils.JSONResponse(w, err.Error(), http.StatusInternalServerError, req)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function":  "UpdatePatientProfile",
		"patientId": claims.UserId,
	}).Info("Patient profile updated successfully")

	utils.JSONResponse(w, resp, http.StatusOK, req)
}

func (p *PatientServerClient) AddPrescriptionForPatient(w http.ResponseWriter, r *http.Request) {
	p.Logger.Info("Received request to add prescription for patient")

	var reqBody struct {
		PatientId    string `json:"patientid"`
		Prescription []struct {
			Medication string `json:"medication"`
			Dosage     string `json:"dosage"`
			Frequency  string `json:"frequency"`
		} `json:"prescription"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "AddPrescriptionForPatient",
			"error":    err.Error(),
		}).Error("Failed to bind JSON")
		utils.JSONStandardResponse(w, "fail", "failed to bind json", "", http.StatusBadRequest, r)
		return
	}

	claims, err := middleware.ExtractClaimsFromCookie(r, "doctor")
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "AddPrescriptionForPatient",
			"error":    "Unauthorized",
		}).Error("Unauthorized access")
		utils.JSONStandardResponse(w, "fail", "Unauthorized", "", http.StatusUnauthorized, r)
		return
	}

	doctorId := claims.UserId
	if doctorId == "" {
		p.Logger.WithFields(logrus.Fields{
			"function": "AddPrescriptionForPatient",
			"error":    "Unauthorized - missing doctor ID",
		}).Error("Unauthorized access")
		utils.JSONStandardResponse(w, "fail", "Unauthorized", "", http.StatusUnauthorized, r)
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

	resp, err := p.PatientClient.AddPrescription(context.Background(), &patient.AddPrescriptionRequest{
		PatientId:    reqBody.PatientId,
		DoctorId:     doctorId,
		Prescription: prescription,
	})
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function":  "AddPrescriptionForPatient",
			"patientId": reqBody.PatientId,
			"error":     err.Error(),
		}).Error("gRPC error while adding prescription")
		utils.JSONResponse(w, err.Error(), http.StatusInternalServerError, r)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function":  "AddPrescriptionForPatient",
		"patientId": reqBody.PatientId,
	}).Info("Prescription added successfully")

	utils.JSONResponse(w, resp, http.StatusOK, r)
}

func (p *PatientServerClient) GetPrescriptions(w http.ResponseWriter, r *http.Request) {
	p.Logger.Info("Received request to get prescriptions")

	val := r.URL.Query()
	query := val.Get("query")
	p.Logger.WithFields(logrus.Fields{
		"function": "GetPrescriptions",
		"query":    query,
	}).Info("Fetching prescriptions with query")

	claims, err := middleware.ExtractClaimsFromCookie(r, "patient")
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "GetPrescriptions",
			"error":    err.Error(),
		}).Error("Unauthorized access")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	resp, err := p.PatientClient.GetPrescription(context.Background(), &patient.GetPrescriptionRequest{
		PatientId: claims.UserId,
		Query:     query,
	})
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function":  "GetPrescriptions",
			"patientId": claims.UserId,
			"error":     err.Error(),
		}).Error("gRPC error while getting prescriptions")
		utils.JSONResponse(w, resp, http.StatusInternalServerError, r)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function":  "GetPrescriptions",
		"patientId": claims.UserId,
	}).Info("Fetched prescriptions successfully")

	utils.JSONResponse(w, resp, http.StatusOK, r)
}

func (p *PatientServerClient) GetPrescriptionsForDoctors(w http.ResponseWriter, r *http.Request) {
	p.Logger.Info("Received request to get prescriptions for doctor")

	val := r.URL.Query()
	query := val.Get("query")
	p.Logger.WithFields(logrus.Fields{
		"function": "GetPrescriptionsForDoctors",
		"query":    query,
	}).Info("Fetching prescriptions for doctor with query")

	var reqBody struct {
		Patient_Id string `json:"patientid"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "GetPrescriptionsForDoctors",
			"error":    err.Error(),
		}).Error("Invalid request format")
		utils.JSONStandardResponse(w, "error", "Invalid request format", "", http.StatusBadRequest, r)
		return
	}

	resp, err := p.PatientClient.GetPrescription(context.Background(), &patient.GetPrescriptionRequest{
		PatientId: reqBody.Patient_Id,
		Query:     query,
	})
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function":  "GetPrescriptionsForDoctors",
			"patientId": reqBody.Patient_Id,
			"error":     err.Error(),
		}).Error("gRPC error while getting prescriptions for doctor")
		utils.JSONResponse(w, resp, http.StatusInternalServerError, r)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function":  "GetPrescriptionsForDoctors",
		"patientId": reqBody.Patient_Id,
	}).Info("Fetched prescriptions for doctor successfully")

	utils.JSONResponse(w, resp, http.StatusOK, r)
}

func (d *PatientServerClient) VideoCallRender(w http.ResponseWriter, r *http.Request) {
	d.Logger.Info("Serving video call HTML page")
	videocallhtml := filepath.Join("..", "templates", "video_call_jitsi.html")
	http.ServeFile(w, r, videocallhtml)
	d.Logger.Info("Video call HTML page served successfully")
}

func (p *PatientServerClient) PatientChatHandler(w http.ResponseWriter, r *http.Request) {
	p.Logger.Info("Patient chat connection request received")
	conn, err := di.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "PatientChatHandler",
			"error":    err.Error(),
		}).Error("Error upgrading patient connection")
		return
	}
	di.PatientConnections[conn] = true
	defer conn.Close()

	for {
		var message di.Message
		err := conn.ReadJSON(&message)
		if err != nil {
			p.Logger.WithFields(logrus.Fields{
				"function": "PatientChatHandler",
				"error":    err.Error(),
			}).Error("Error reading message from patient")
			delete(di.PatientConnections, conn)
			break
		}
		p.Logger.WithFields(logrus.Fields{
			"function": "PatientChatHandler",
			"message":  message,
		}).Info("Received message from patient")

		sendMessageToCustomerCare(message) // Implement this function to handle sending messages
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
