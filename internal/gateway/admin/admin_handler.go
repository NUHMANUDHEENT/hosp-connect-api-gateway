package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	pb "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/admin"
	"github.com/gorilla/mux"

	// "github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/admin"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/utils"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/middleware"
)

const role = "admin"

// AdminSignIn handles the admin signin via gRPC
func (a *AdminServerClient) AdminSignIn(w http.ResponseWriter, r *http.Request) {
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
	resp, err := a.SignIn(context.Background(), &pb.SignInRequest{
		Email:    reqBody.Email,
		Password: reqBody.Password,
	})
	if err != nil {
		utils.JSONResponse(w, "GRPC error", http.StatusInternalServerError, r)
		return
	}
	if resp.Status == "success" {
		jwtToken, err := middleware.CreateJWTToken(reqBody.Email, role)
		if err != nil {
			utils.JSONResponse(w, "Failed to create JWT token", http.StatusInternalServerError, r)
			return
		}
		middleware.SetJWTToken(w, jwtToken, role)
	}

	utils.JSONResponse(w, resp, http.StatusOK, r)
}
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

	utils.JSONStandardResponse(w, "success", "", "Admin logged out successfully", http.StatusOK, r)
}

func (a *AdminServerClient) DoctorRegister(w http.ResponseWriter, req *http.Request) {
	var reqBody struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		Name             string `json:"name"`
		Phone            int    `json:"phone"`
		SpecializationId int    `json:"specializationId"`
	}
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		utils.JSONResponse(w, "failed to decode request", http.StatusBadGateway, req)
		return
	}
	resp, err := a.AddDoctor(context.Background(), &pb.AddDoctorRequest{
		Email:            reqBody.Email,
		Password:         reqBody.Password,
		Name:             reqBody.Name,
		SpecializationId: int32(reqBody.SpecializationId),
		Phone:            int32(reqBody.Phone),
	})
	if err != nil {
		utils.JSONResponse(w, resp, http.StatusBadGateway, req)
		return
	}
	utils.JSONResponse(w, resp, int(resp.StatusCode), req)

}

func (a *AdminServerClient) PatientCreate(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Phone    int    `json:"phone"`
		Age      int    `json:"age"`
		Gender   string `json:"gender"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		utils.JSONResponse(w, "failed to decode request", http.StatusBadRequest, r)
		return
	}

	// Call the Patient gRPC Create method
	resp, err := a.AddPatient(context.Background(), &pb.AddPatientRequest{
		Name:     reqBody.Name,
		Email:    reqBody.Email,
		Phone:    int32(reqBody.Phone),
		Password: reqBody.Password,
	})
	if err != nil {
		utils.JSONResponse(w, "GRPC error", http.StatusInternalServerError, r)
		return
	}

	utils.JSONResponse(w, resp, int(resp.StatusCode), r)
}

func (a *AdminServerClient) PatientDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	patientId := vars["ID"]

	// Call the Patient gRPC Delete method
	resp, err := a.DeletePatient(context.Background(), &pb.DeletePatientRequest{
		PatientId: patientId,
	})
	if err != nil {
		utils.JSONResponse(w, "GRPC error", http.StatusInternalServerError, r)
		return
	}

	utils.JSONResponse(w, resp, int(resp.StatusCode), r)
}

func (a *AdminServerClient) DoctorDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	doctorID := vars["ID"]

	if doctorID == "" {
		utils.JSONResponse(w, "Doctor ID is missing", http.StatusBadRequest, r)
		return
	}

	// Call the Doctor gRPC Delete method
	resp, err := a.DeleteDoctor(context.Background(), &pb.DeleteDoctorRequest{
		DoctorId: doctorID,
	})
	if err != nil {
		utils.JSONResponse(w, "GRPC error", http.StatusInternalServerError, r)
		return
	}

	utils.JSONResponse(w, resp, int(resp.StatusCode), r)
}

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

	// Call the Patient gRPC Block method (this updates the patient status)
	resp, err := a.BlockPatient(context.Background(), &pb.BlockPatientRequest{
		PatientId: patientId,
		Reason:    reqBody.Reason,
	})
	if err != nil {
		utils.JSONResponse(w, "GRPC error", http.StatusInternalServerError, r)
		return
	}

	utils.JSONResponse(w, resp, int(resp.StatusCode), r)
}
func (a *AdminServerClient) ListDoctorsHandler(w http.ResponseWriter, req *http.Request) {
	resp, err := a.ListDoctors(context.Background(), &pb.Empty{})
	if err != nil || resp.Status != "success" {
		utils.JSONResponse(w, "Failed to list doctors", http.StatusInternalServerError, req)
		return
	}
	fmt.Println("docotrs",resp)
	utils.JSONResponse(w, resp, http.StatusOK, req)
}

// ListPatients handles the request for listing all patients
func (a *AdminServerClient) ListPatientsHandler(w http.ResponseWriter, req *http.Request) {
	resp, err := a.ListPatients(context.Background(), &pb.Empty{})
	if err != nil || resp.Status != "success" {
		utils.JSONResponse(w, "Failed to list patients", http.StatusInternalServerError, req)
		return
	}

	utils.JSONResponse(w, resp, http.StatusOK, req)
}
func (a *AdminServerClient) AddDoctorSpecialization(w http.ResponseWriter, req *http.Request) {
	var reqBody struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		utils.JSONResponse(w, "failed to decode request", http.StatusBadRequest, req)
		return
	}
	resp, err := a.AddSpecialization(context.Background(), &pb.AddSpecializationRequest{
		Name:        reqBody.Name,
		Description: reqBody.Description,
	})
	if err != nil {
		utils.JSONResponse(w, "GRPC erro", http.StatusInternalServerError, req)
		return
	}
	utils.JSONResponse(w, resp, 200, req)
}
