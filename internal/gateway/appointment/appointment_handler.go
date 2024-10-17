package appointment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pb "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/appointment"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/utils"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/middleware"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (p *AppointmentServerClient) GetAvailability(w http.ResponseWriter, r *http.Request) {
	var reqbody struct {
		CategoryID        int       `json:"categoryid"`
		RequestedDateTime time.Time `json:"requesteddatetime"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqbody)
	if err != nil {
		utils.JSONStandardResponse(w, "fail", "Failed to bind json data", "", 500, r)
		return
	}
	fmt.Println("reqqqq", reqbody.CategoryID)
	resp, err := p.CheckAvailability(context.Background(), &pb.GetAvailabilityRequest{
		RequestedDateTime: timestamppb.New(reqbody.RequestedDateTime),
		CategoryId:        int32(reqbody.CategoryID),
	})
	if err != nil {
		utils.JSONStandardResponse(w, "fail", "Failed to call service", "", 500, r)
		return
	}
	utils.JSONResponse(w, resp, 200, r)
}
func (p *AppointmentServerClient) GetAvailabilityByDoctorId(w http.ResponseWriter, r *http.Request) {
	var reqbody struct {
		DoctorId string `json:"doctorid"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqbody)
	if err != nil {
		utils.JSONStandardResponse(w, "fail", "Failed to bind json data", "", 500, r)
		return
	}
	resp, err := p.CheckAvailabilityByDoctorId(context.Background(), &pb.CheckAvailabilityByDoctorIdRequest{
		DoctorId: reqbody.DoctorId,
	})
	if err != nil {
		utils.JSONStandardResponse(w, "fail", "Failed to call appointment grpc", "", 500, r)
		return
	}
	fmt.Println("gyyy", resp)
	utils.JSONResponse(w, resp, 200, r)
}
func (p *AppointmentServerClient) ConfirmPatientAppointment(w http.ResponseWriter, r *http.Request) {
	var reqbody struct {
		SpecializationId int       `json:"specializationid"`
		AppointmentTime  time.Time `json:"appointmenttime"`
		DoctorId         string    `json:"doctorid"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqbody)
	if err != nil {
		utils.JSONStandardResponse(w, "fail", "Failed to bind json data", "", 500, r)
		return
	}
	parsedTime := reqbody.AppointmentTime.UTC()

	claims, err := middleware.ExtractClaimsFromCookie(r, "patient")
	if err != nil {
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, r)
		return
	}
	appointmentReq := &pb.ConfirmAppointmentRequest{
		DoctorId:          reqbody.DoctorId,
		SpecializationId:  int32(reqbody.SpecializationId),
		ConfirmedDateTime: timestamppb.New(parsedTime),
		PatientId:         claims.UserId,
	}
	fmt.Println("datas", appointmentReq)
	resp, err := p.ConfirmAppointment(context.Background(), appointmentReq)
	if err != nil {
		utils.JSONStandardResponse(w, "fail", "Failed to call service", "", 500, r)
		return
	}
	utils.JSONResponse(w, resp, 200, r)

}
func (a *AppointmentServerClient) GetAppointments(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.ExtractClaimsFromCookie(r, "patient")
	if err != nil {
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, r)
		return
	}
	resp, err := a.GetUpcomingAppointments(context.Background(), &pb.GetAppointmentsRequest{
		PatientId: claims.UserId,
	})
	if err != nil {
		utils.JSONResponse(w, resp, 400, r)
		return
	}
	utils.JSONResponse(w, resp, 200, r)
}
func (d *AppointmentServerClient) CreateRoomForVideoTreatments(w http.ResponseWriter, req *http.Request) {
	claims, err := middleware.ExtractClaimsFromCookie(req, "doctor")
	if err != nil {
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, req)
		return
	}
	var reqBody struct {
		PatientId        string `json:"patientid"`
		SpecializationId int64  `json:"specialization"`
	}
	err = json.NewDecoder(req.Body).Decode(&reqBody)
                          if err != nil {
		utils.JSONResponse(w, "Invalid request body", http.StatusBadRequest, req)
		return
	}
	resp, err := d.CreateRoomForVideoTreatment(context.Background(), &pb.VideoRoomRequest{
		PatientId:        reqBody.PatientId,
		SpecializationId: reqBody.SpecializationId,
		DoctorId:         claims.UserId,
	})
	if err != nil {
		utils.JSONStandardResponse(w, "fail", "Failed to create room for video treatment: "+err.Error(), "", http.StatusInternalServerError, req)
		return
	}
	utils.JSONResponse(w, resp, 200, req)

}

