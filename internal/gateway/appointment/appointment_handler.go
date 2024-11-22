package appointment

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	pb "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/appointment"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/utils"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/middleware"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetAvailability handles getting availability for a specific category and requested date
func (p *AppointmentServerClient) GetAvailability(w http.ResponseWriter, r *http.Request) {
	p.Logger.Info("Received request to get availability")

	var reqbody struct {
		CategoryID        int       `json:"categoryid" validate:"required"`
		RequestedDateTime time.Time `json:"requesteddatetime" validate:"required"`
	}

	err := json.NewDecoder(r.Body).Decode(&reqbody)
	if err != nil {
		utils.JSONStandardResponse(w, "fail", "Failed to bind json data", "", http.StatusBadRequest, r)
		return
	}
	ok, er := utils.ValidateInput(reqbody)
	if !ok {
		utils.JSONStandardResponse(w, "Fail", er, "", http.StatusBadRequest, r)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function":      "GetAvailability",
		"categoryID":    reqbody.CategoryID,
		"requestedTime": reqbody.RequestedDateTime,
	}).Info("Processing availability check")

	resp, err := p.CheckAvailability(context.Background(), &pb.GetAvailabilityRequest{
		RequestedDateTime: timestamppb.New(reqbody.RequestedDateTime),
		CategoryId:        int32(reqbody.CategoryID),
	})
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "GetAvailability",
			"error":    err.Error(),
		}).Error("Failed to call availability service")
		utils.JSONStandardResponse(w, "fail", "Failed to call service", "", http.StatusInternalServerError, r)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function": "GetAvailability",
		"response": resp,
	}).Info("Availability fetched successfully")
	utils.JSONResponse(w, resp, http.StatusOK, r)
}

// GetAvailabilityByDoctorId handles availability check for a specific doctor
func (p *AppointmentServerClient) GetAvailabilityByDoctorId(w http.ResponseWriter, r *http.Request) {
	p.Logger.Info("Received request to get availability by doctor ID")

	var reqbody struct {
		DoctorId string `json:"doctorid" validate:"required"`
	}

	err := json.NewDecoder(r.Body).Decode(&reqbody)
	if err != nil {
		utils.JSONStandardResponse(w, "fail", "Failed to bind json data", "", http.StatusBadRequest, r)
		return
	}

	ok, er := utils.ValidateInput(reqbody)
	if !ok {
		utils.JSONStandardResponse(w, "Fail", er, "", http.StatusBadRequest, r)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function": "GetAvailabilityByDoctorId",
		"doctorId": reqbody.DoctorId,
	}).Info("Checking availability for doctor")

	resp, err := p.CheckAvailabilityByDoctorId(context.Background(), &pb.CheckAvailabilityByDoctorIdRequest{
		DoctorId: reqbody.DoctorId,
	})
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "GetAvailabilityByDoctorId",
			"error":    err.Error(),
		}).Error("Failed to call appointment service")
		utils.JSONStandardResponse(w, "fail", "Failed to call appointment grpc", "", http.StatusInternalServerError, r)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function": "GetAvailabilityByDoctorId",
		"response": resp,
	}).Info("Doctor availability fetched successfully")
	utils.JSONResponse(w, resp, http.StatusOK, r)
}

// ConfirmPatientAppointment handles confirming a patient appointment
func (p *AppointmentServerClient) ConfirmPatientAppointment(w http.ResponseWriter, r *http.Request) {
	p.Logger.Info("Received request to confirm patient appointment")

	var reqbody struct {
		SpecializationId int       `json:"specializationid" validate:"required"`
		AppointmentTime  time.Time `json:"appointmenttime" validate:"required"`
		DoctorId         string    `json:"doctorid" validate:"required"`
		Type             string    `json:"type" validate:"required"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqbody)
	if err != nil {
		utils.JSONStandardResponse(w, "fail", "Failed to bind json data", "", http.StatusBadRequest, r)
		return
	}

	ok, er := utils.ValidateInput(reqbody)
	if !ok {
		utils.JSONStandardResponse(w, "Fail", er, "", http.StatusBadRequest, r)
		return
	}

	parsedTime := reqbody.AppointmentTime.UTC()

	claims, err := middleware.ExtractClaimsFromCookie(r, "patient")
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "ConfirmPatientAppointment",
			"error":    err.Error(),
		}).Error("Unauthorized access attempt")
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, r)
		return
	}

	appointmentReq := &pb.ConfirmAppointmentRequest{
		DoctorId:          reqbody.DoctorId,
		SpecializationId:  int32(reqbody.SpecializationId),
		ConfirmedDateTime: timestamppb.New(parsedTime),
		PatientId:         claims.UserId,
		Type:              reqbody.Type,
	}

	p.Logger.WithFields(logrus.Fields{
		"function":  "ConfirmPatientAppointment",
		"patientId": claims.UserId,
		"request":   appointmentReq,
	}).Info("Confirming patient appointment")

	resp, err := p.ConfirmAppointment(context.Background(), appointmentReq)
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "ConfirmPatientAppointment",
			"error":    err.Error(),
		}).Error("Failed to call confirm appointment service")
		utils.JSONStandardResponse(w, "fail", "Failed to call appointment service", "", http.StatusInternalServerError, r)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function": "ConfirmPatientAppointment",
		"response": resp,
	}).Info("Appointment confirmed successfully")
	utils.JSONResponse(w, resp, http.StatusOK, r)

}
func (p *AppointmentServerClient) CancelPatientAppointment(w http.ResponseWriter, r *http.Request) {
	var reqbody struct {
		AppointmentId int32  `json:"appointmentId" validate:"required"`
		Reason        string `json:"reason" validate:"required"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqbody)
	if err != nil {
		utils.JSONStandardResponse(w, "fail", "Failed to bind json data", "", http.StatusBadRequest, r)
		return
	}

	ok, er := utils.ValidateInput(reqbody)
	if !ok {
		utils.JSONStandardResponse(w, "Fail", er, "", http.StatusBadRequest, r)
		return
	}

	claims, err := middleware.ExtractClaimsFromCookie(r, "patient")
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "CancelAppointment",
			"error":    err.Error(),
		}).Error("Unauthorized access attempt")
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, r)
		return
	}
	resp, err := p.CancelAppointment(context.Background(), &pb.CancelAppointmentRequest{
		PatientId:     claims.UserId,
		AppointmentId: reqbody.AppointmentId,
		Reason:        reqbody.Reason,
	})
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "CancelAppointment",
			"error":    err.Error(),
		}).Error("Failed to call cancel appointment service")
		utils.JSONStandardResponse(w, "fail", "Failed to call service", "", http.StatusInternalServerError, r)
	}
	p.Logger.WithFields(logrus.Fields{
		"function": "CancelAppointment",
		"response": resp,
	}).Info("Appointment cancelled successfully")
	utils.JSONResponse(w, resp, http.StatusOK, r)

}

// GetAppointments handles fetching upcoming appointments for a patient
func (p *AppointmentServerClient) GetAppointments(w http.ResponseWriter, r *http.Request) {
	p.Logger.Info("Received request to get upcoming appointments")

	claims, err := middleware.ExtractClaimsFromCookie(r, "patient")
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "GetAppointments",
			"error":    err.Error(),
		}).Error("Unauthorized access attempt")
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, r)
		return
	}

	resp, err := p.GetUpcomingAppointments(context.Background(), &pb.GetAppointmentsRequest{
		PatientId: claims.UserId,
	})
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "GetAppointments",
			"error":    err.Error(),
		}).Error("Failed to retrieve appointments")
		utils.JSONResponse(w, resp, http.StatusBadRequest, r)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function": "GetAppointments",
		"response": resp,
	}).Info("Upcoming appointments fetched successfully")
	utils.JSONResponse(w, resp, http.StatusOK, r)
}

// CreateRoomForVideoTreatments handles the creation of a video treatment room via gRPC
func (d *AppointmentServerClient) CreateRoomForVideoTreatments(w http.ResponseWriter, req *http.Request) {
	d.Logger.Info("Received request to create a video treatment room")

	claims, err := middleware.ExtractClaimsFromCookie(req, "doctor")
	if err != nil {
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, req)
		return
	}

	var reqBody struct {
		PatientId        string `json:"patientid" validate:"required"`
		SpecializationId int64  `json:"specialization" validate:"required"`
	}
	err = json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		utils.JSONResponse(w, "Invalid request body", http.StatusBadRequest, req)
		return
	}

	ok, er := utils.ValidateInput(reqBody)
	if !ok {
		utils.JSONStandardResponse(w, "Fail", er, "", http.StatusBadRequest, req)
		return
	}

	resp, err := d.CreateRoomForVideoTreatment(context.Background(), &pb.VideoRoomRequest{
		PatientId:        reqBody.PatientId,
		SpecializationId: reqBody.SpecializationId,
		DoctorId:         claims.UserId,
	})
	if err != nil {
		d.Logger.WithFields(logrus.Fields{
			"function": "CreateRoomForVideoTreatments",
			"error":    err.Error(),
		}).Error("Failed to create room for video treatment")
		utils.JSONStandardResponse(w, "fail", "Failed to create room for video treatment: "+err.Error(), "", http.StatusInternalServerError, req)
		return
	}

	d.Logger.WithFields(logrus.Fields{
		"function": "CreateRoomForVideoTreatments",
		"response": resp,
	}).Info("Video treatment room created successfully")
	utils.JSONResponse(w, resp, http.StatusOK, req)
}

// VideoCallRender serves the video call HTML page for doctors
func (d *AppointmentServerClient) VideoCallRender(w http.ResponseWriter, r *http.Request) {
	d.Logger.Info("Serving video call page for doctor")

	_, err := middleware.ExtractClaimsFromCookie(r, "doctor")
	if err != nil {
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, r)
		return
	}

	videocallhtml := filepath.Join("..", "templates", "video_call_jitsi.html")
	http.ServeFile(w, r, videocallhtml)
	d.Logger.Info("Video call page served successfully")
}

// Dashboard serves the admin dashboard HTML page
func (d *AppointmentServerClient) Dashboard(w http.ResponseWriter, r *http.Request) {
	d.Logger.Info("Serving admin dashboard")

	_, err := middleware.ExtractClaimsFromCookie(r, "admin")
	if err != nil {
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, r)
		return
	}

	dashboardHTML := filepath.Join("templates", "dashboard.html")
	http.ServeFile(w, r, dashboardHTML)
	d.Logger.Info("Dashboard page served successfully")
}

// DashboardResponse fetches statistics details for the dashboard
func (p *AppointmentServerClient) DashboardResponse(w http.ResponseWriter, r *http.Request) {
	p.Logger.Info("Fetching dashboard statistics")

	filterParam := r.URL.Query().Get("filter")
	if filterParam == "" {
		utils.JSONResponse(w, "Missing filter parameter", http.StatusBadRequest, r)
		return
	}

	resp, err := p.AppointmentServiceClient.FetchStatisticsDetails(context.Background(), &pb.StatisticsRequest{
		Param: filterParam,
	})
	if err != nil {
		p.Logger.WithFields(logrus.Fields{
			"function": "DashboardResponse",
			"error":    err.Error(),
		}).Error("Failed to fetch dashboard stats")
		http.Error(w, "Failed to fetch dashboard stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	p.Logger.WithFields(logrus.Fields{
		"function": "DashboardResponse",
		"response": resp,
	}).Info("Dashboard statistics fetched successfully")
}

// AddDoctorSpecialization adds a new specialization for doctors via gRPC
func (a *AppointmentServerClient) AddDoctorSpecialization(w http.ResponseWriter, req *http.Request) {
	a.Logger.Info("Received request to add a new doctor specialization")

	var reqBody struct {
		Name        string `json:"name" validate:"required"`
		Description string `json:"description" validate:"required"`
	}
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		utils.JSONResponse(w, "Failed to decode request body", http.StatusBadRequest, req)
		return
	}

	ok, er := utils.ValidateInput(reqBody)
	if !ok {
		utils.JSONStandardResponse(w, "Fail", er, "", http.StatusBadRequest, req)
		return
	}

	resp, err := a.AddSpecialization(context.Background(), &pb.AddSpecializationRequest{
		Name:        reqBody.Name,
		Description: reqBody.Description,
	})
	if err != nil {
		a.Logger.WithFields(logrus.Fields{
			"function": "AddDoctorSpecialization",
			"error":    err.Error(),
		}).Error("Failed to add specialization")
		utils.JSONResponse(w, "Failed to add specialization: "+err.Error(), http.StatusInternalServerError, req)
		return
	}

	a.Logger.WithFields(logrus.Fields{
		"function": "AddDoctorSpecialization",
		"response": resp,
	}).Info("Doctor specialization added successfully")
	utils.JSONResponse(w, resp, http.StatusOK, req)
}
