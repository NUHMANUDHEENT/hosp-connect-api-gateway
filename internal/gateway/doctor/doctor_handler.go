package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/NUHMANUDHEENT/hosp-connect-pb/proto/doctor"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/utils"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/middleware"
	"golang.org/x/oauth2"
)

const role = "doctor"

func (d *DoctorServerClient) DoctorSignIn(w http.ResponseWriter, req *http.Request) {
	log.Println("signin request")
	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := d.SignIn(context.Background(), &doctor.SignInRequest{
		Email:    reqBody.Email,
		Password: reqBody.Password,
	})
	if err != nil {
		utils.JSONResponse(w, err.Error(), http.StatusUnauthorized, req)
		return
	}
	fmt.Println("doctor id ", resp.DoctorId)
	if resp.Status == "success" {
		jwtToken, err := middleware.CreateJWTToken(resp.DoctorId, role)
		if err != nil {
			utils.JSONResponse(w, "Failed to create JWT token", http.StatusInternalServerError, req)
			return
		}
		// Set JWT token in cookie
		middleware.SetJWTToken(w, jwtToken, role)
	}

	utils.JSONResponse(w, resp, http.StatusOK, req)

}
func (d *DoctorServerClient) DoctorLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		MaxAge:   -1,
		Name:     "doctortoken",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	utils.JSONStandardResponse(w, "success", "", "Doctor logged out successfully", http.StatusOK, r)
}
func (d *DoctorServerClient) GetDoctorProfile(w http.ResponseWriter, req *http.Request) {
	claims, err := middleware.ExtractClaimsFromCookie(req, "doctor")
	if err != nil {
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, req)
		return
	}

	doctorId := claims.UserId

	// Call gRPC to fetch the profile
	resp, err := d.GetProfile(context.Background(), &doctor.GetProfileRequest{
		DoctorId: doctorId,
	})
	if err != nil || resp.Status != "success" {
		utils.JSONResponse(w, "Failed to get profile", http.StatusInternalServerError, req)
		return
	}

	// Return profile information
	utils.JSONResponse(w, resp, http.StatusOK, req)
}
func (d *DoctorServerClient) UpdateDoctorProfile(w http.ResponseWriter, req *http.Request) {
	claims, err := middleware.ExtractClaimsFromCookie(req, "doctor")
	if err != nil {
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, req)
		return
	}

	doxtorId := claims.UserId
	// Parse request body to update profile
	var reqBody struct {
		Name             string `json:"name"`
		SpecializationId int32  `json:"specialization"`
		phone            int    `json:"phone"`
	}
	err = json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		utils.JSONResponse(w, "Invalid request body", http.StatusBadRequest, req)
		return
	}

	// Create doctor update request
	updateReq := &doctor.UpdateProfileRequest{
		Doctor: &doctor.Doctor{
			DoctorId:         doxtorId,
			Name:             reqBody.Name,
			SpecializationId: reqBody.SpecializationId,
			Phone:            int32(reqBody.phone),
		},
	}

	// Call gRPC service to update the profile
	resp, err := d.UpdateProfile(context.Background(), updateReq)
	if err != nil || resp.Status != "success" {
		utils.JSONResponse(w, "Failed to update profile", http.StatusInternalServerError, req)
		return
	}

	// Return success response
	utils.JSONResponse(w, resp, http.StatusOK, req)
}

func (d *DoctorServerClient) DoctorStoreAccessToken(email string, token *oauth2.Token) error {
	_, err := d.StoreAccessToken(context.Background(), &doctor.StoreAccessTokenRequest{
		Email:        email,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry.String(),
	})
	if err != nil {
		return err
	}
	return nil
}
func (d *DoctorServerClient) ConfirmScheduleHandler(w http.ResponseWriter, req *http.Request) {
	// Extract the doctor claims to verify authentication
	claims, err := middleware.ExtractClaimsFromCookie(req, "doctor")
	if err != nil {
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, req)
		return
	}

	doctorID := claims.UserId // Assuming claims contain the doctor ID
	if doctorID == "" {
		utils.JSONResponse(w, "Missing doctor ID", http.StatusBadRequest, req)
		return
	}

	// Make gRPC call to confirm the doctor's schedule
	grpcReq := &doctor.ConfirmScheduleRequest{
		DoctorId: doctorID,
	}

	grpcResp, err := d.ConfirmSchedule(context.Background(), grpcReq)
	if err != nil {
		utils.JSONStandardResponse(w, "fail", "Failed to confirm schedule: "+err.Error(), "", http.StatusInternalServerError, req)
		return
	}

	if grpcResp.Status != "success" {
		utils.JSONResponse(w, "Failed to confirm schedule: "+grpcResp.Error, http.StatusInternalServerError, req)
		return
	}

	utils.JSONResponse(w, grpcResp.Schedules, http.StatusOK, req)
}
