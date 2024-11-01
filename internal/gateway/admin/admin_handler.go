package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"

	pb "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/admin"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/di"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/utils"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/middleware"
)

const role = "admin"

// AdminSignIn handles the admin sign-in via gRPC
func (a *AdminServerClient) AdminSignIn(w http.ResponseWriter, r *http.Request) {
	a.Logger.Info("AdminSignIn: Starting sign-in process")

	// Parse the JSON request body
	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.JSONResponse(w, "Unable to read request body", http.StatusBadRequest, r)
		return
	}

	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		utils.JSONResponse(w, "Invalid request format", http.StatusBadRequest, r)
		return
	}

	// Call the Admin gRPC SignIn method
	resp, err := a.AdminClient.SignIn(context.Background(), &pb.SignInRequest{
		Email:    reqBody.Email,
		Password: reqBody.Password,
	})
	if err != nil {
		a.Logger.WithFields(logrus.Fields{
			"function": "AdminSignIn",
			"error":    err.Error(),
			"email":    reqBody.Email,
		}).Error("AdminSignIn: gRPC error during sign-in")
		utils.JSONResponse(w, "GRPC error", http.StatusInternalServerError, r)
		return
	}

	if resp.Status == "success" {
		jwtToken, err := middleware.CreateJWTToken(reqBody.Email, role)
		if err != nil {
			utils.JSONResponse(w, "Failed to create JWT token", http.StatusInternalServerError, r)
			return
		}
		middleware.SetJWTToken(w, jwtToken, "admin")
		a.Logger.WithFields(logrus.Fields{
			"function": "AdminSignIn",
			"email":    reqBody.Email,
		}).Info("AdminSignIn: JWT token created successfully")
	}

	a.Logger.WithFields(logrus.Fields{
		"function": "AdminSignIn",
		"email":    reqBody.Email,
	}).Info("AdminSignIn: Sign-in successful")
	utils.JSONResponse(w, resp, http.StatusOK, r)
}

// AdminLogout handles admin logout
func (a *AdminServerClient) AdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		MaxAge:   -1,
		Name:     "admintoken",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	a.Logger.WithFields(logrus.Fields{
		"function": "AdminLogout",
	}).Info("AdminLogout: Admin logged out successfully")
	utils.JSONStandardResponse(w, "success", "", "Admin logged out successfully", http.StatusOK, r)
}

// DoctorRegister handles the doctor registration via gRPC
func (a *AdminServerClient) DoctorRegister(w http.ResponseWriter, req *http.Request) {
	a.Logger.Info("DoctorRegister: Starting registration process")

	var reqBody struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		Name             string `json:"name"`
		Phone            int    `json:"phone"`
		SpecializationId int    `json:"specializationId"`
	}
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		// Skip logging for decoding errors as it's too verbose
		utils.JSONResponse(w, "failed to decode request", http.StatusBadGateway, req)
		return
	}

	resp, err := a.AdminClient.AddDoctor(context.Background(), &pb.AddDoctorRequest{
		Email:            reqBody.Email,
		Password:         reqBody.Password,
		Name:             reqBody.Name,
		SpecializationId: int32(reqBody.SpecializationId),
		Phone:            int32(reqBody.Phone),
	})
	if err != nil {
		a.Logger.WithFields(logrus.Fields{
			"function": "DoctorRegister",
			"error":    err.Error(),
			"email":    reqBody.Email,
		}).Error("DoctorRegister: Failed to register doctor")
		utils.JSONResponse(w, resp, http.StatusBadGateway, req)
		return
	}

	a.Logger.WithFields(logrus.Fields{
		"function": "DoctorRegister",
		"email":    reqBody.Email,
	}).Info("DoctorRegister: Doctor registered successfully")
	utils.JSONResponse(w, resp, int(resp.StatusCode), req)
}

// PatientCreate handles the creation of a new patient
func (a *AdminServerClient) PatientCreate(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Phone    int    `json:"phone"`
		Age      int    `json:"age"`
		Gender   string `json:"gender"`
	}

	// Decode the request body
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		utils.JSONResponse(w, "failed to decode request", http.StatusBadRequest, r)
		return
	}

	// Call the Patient gRPC Create method
	resp, err := a.AdminClient.AddPatient(context.Background(), &pb.AddPatientRequest{
		Name:     reqBody.Name,
		Email:    reqBody.Email,
		Phone:    int32(reqBody.Phone),
		Password: reqBody.Password,
	})
	if err != nil {
		a.Logger.WithFields(logrus.Fields{
			"function": "PatientCreate",
			"email":    reqBody.Email,
			"error":    err.Error(),
		}).Error("PatientCreate: gRPC error during patient creation")
		utils.JSONResponse(w, "GRPC error", http.StatusInternalServerError, r)
		return
	}

	a.Logger.WithFields(logrus.Fields{
		"function": "PatientCreate",
		"email":    reqBody.Email,
	}).Info("PatientCreate: Patient created successfully")
	utils.JSONResponse(w, resp, int(resp.StatusCode), r)
}

// PatientDelete handles the deletion of a patient
func (a *AdminServerClient) PatientDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	patientId := vars["ID"]

	// Call the Patient gRPC Delete method
	resp, err := a.AdminClient.DeletePatient(context.Background(), &pb.DeletePatientRequest{
		PatientId: patientId,
	})
	if err != nil {
		a.Logger.WithFields(logrus.Fields{
			"function":  "PatientDelete",
			"patientId": patientId,
			"error":     err.Error(),
		}).Error("PatientDelete: gRPC error during patient deletion")
		utils.JSONResponse(w, "GRPC error", http.StatusInternalServerError, r)
		return
	}

	a.Logger.WithFields(logrus.Fields{
		"function":  "PatientDelete",
		"patientId": patientId,
	}).Info("PatientDelete: Patient deleted successfully")
	utils.JSONResponse(w, resp, int(resp.StatusCode), r)
}

// DoctorDelete handles the deletion of a doctor
func (a *AdminServerClient) DoctorDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	doctorID := vars["ID"]

	if doctorID == "" {
		utils.JSONResponse(w, "Doctor ID is missing", http.StatusBadRequest, r)
		return
	}

	// Call the Doctor gRPC Delete method
	resp, err := a.AdminClient.DeleteDoctor(context.Background(), &pb.DeleteDoctorRequest{
		DoctorId: doctorID,
	})
	if err != nil {
		a.Logger.WithFields(logrus.Fields{
			"function": "DoctorDelete",
			"doctorID": doctorID,
			"error":    err.Error(),
		}).Error("DoctorDelete: gRPC error during doctor deletion")
		utils.JSONResponse(w, "GRPC error", http.StatusInternalServerError, r)
		return
	}

	a.Logger.WithFields(logrus.Fields{
		"function": "DoctorDelete",
		"doctorID": doctorID,
	}).Info("DoctorDelete: Doctor deleted successfully")
	utils.JSONResponse(w, resp, int(resp.StatusCode), r)
}

// PatientBlock handles the blocking of a patient
func (a *AdminServerClient) PatientBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	patientId := vars["ID"]

	var reqBody struct {
		Reason string `json:"reason"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		utils.JSONResponse(w, "failed to decode request", http.StatusBadRequest, r)
		return
	}

	// Call the Patient gRPC Block method
	resp, err := a.AdminClient.BlockPatient(context.Background(), &pb.BlockPatientRequest{
		PatientId: patientId,
		Reason:    reqBody.Reason,
	})
	if err != nil {
		a.Logger.WithFields(logrus.Fields{
			"function":  "PatientBlock",
			"patientId": patientId,
			"error":     err.Error(),
		}).Error("PatientBlock: gRPC error during patient blocking")
		utils.JSONResponse(w, "GRPC error", http.StatusInternalServerError, r)
		return
	}

	a.Logger.WithFields(logrus.Fields{
		"function":  "PatientBlock",
		"patientId": patientId,
		"reason":    reqBody.Reason,
	}).Info("PatientBlock: Patient blocked successfully")
	utils.JSONResponse(w, resp, int(resp.StatusCode), r)
}

// ListDoctorsHandler handles the request for listing all doctors
func (a *AdminServerClient) ListDoctorsHandler(w http.ResponseWriter, req *http.Request) {
	resp, err := a.AdminClient.ListDoctors(context.Background(), &pb.Empty{})
	if err != nil || resp.Status != "success" {
		utils.JSONResponse(w, "Failed to list doctors", http.StatusInternalServerError, req)
		return
	}
	fmt.Println("Doctors response:", resp)
	utils.JSONResponse(w, resp, http.StatusOK, req)
}

// ListPatientsHandler handles the request for listing all patients
func (a *AdminServerClient) ListPatientsHandler(w http.ResponseWriter, req *http.Request) {
	resp, err := a.AdminClient.ListPatients(context.Background(), &pb.Empty{})
	if err != nil || resp.Status != "success" {
		utils.JSONResponse(w, "Failed to list patients", http.StatusInternalServerError, req)
		return
	}
	utils.JSONResponse(w, resp, http.StatusOK, req)
}

// CustomerCareChatHandler handles WebSocket connections for customer care chat
func (a *AdminServerClient) CustomerCareChatHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := di.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading customer connection:", err)
		return
	}
	di.CustomerConnections[conn] = true // Mark the connection as active
	defer func() {
		conn.Close()
		delete(di.CustomerConnections, conn) // Clean up on disconnect
	}()

	for {
		var message di.Message
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println("Error reading message from customer:", err)
			break // Exit the loop on error
		}
		log.Printf("Received message from customer care: %s", message)

		// Route the message to patients
		sendMessageToPatients(message) // Function to send messages to patients
	}
}

// sendMessageToPatients broadcasts messages to all patient connections
func sendMessageToPatients(msg di.Message) {
	messageJSON, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshalling message to JSON:", err)
		return
	}
	for conn := range di.PatientConnections {
		err := conn.WriteMessage(websocket.TextMessage, messageJSON)
		if err != nil {
			log.Println("Error sending message to patient:", err)
			conn.Close()                         // Close the connection on error
			delete(di.PatientConnections, conn) // Clean up inactive connections
		}
	}
}

// AdminChatRender renders the customer care chat HTML page
func (p *AdminServerClient) AdminChatRender(w http.ResponseWriter, r *http.Request) {
	chatPagePath := filepath.Join("..", "templates", "customer_care_chat.html")
	http.ServeFile(w, r, chatPagePath) // Serve the chat HTML file
}
