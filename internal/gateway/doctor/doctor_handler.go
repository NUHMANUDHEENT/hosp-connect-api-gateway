package doctor

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/NUHMANUDHEENT/hosp-connect-pb/proto/doctor"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/utils"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/middleware"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const role = "doctor"

func (d *DoctorServerClient) DoctorSignIn(w http.ResponseWriter, req *http.Request) {
	d.Logger.Info("Received sign-in request")

	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		d.Logger.WithFields(logrus.Fields{
			"function": "DoctorSignIn",
			"error":    err.Error(),
		}).Error("Failed to decode request body")
		utils.JSONResponse(w, "Failed to decode request", http.StatusBadRequest, req)
		return
	}

	resp, err := d.DoctorClient.SignIn(context.Background(), &doctor.SignInRequest{
		Email:    reqBody.Email,
		Password: reqBody.Password,
	})
	if err != nil {
		d.Logger.WithFields(logrus.Fields{
			"function": "DoctorSignIn",
			"email":    reqBody.Email,
			"error":    err.Error(),
		}).Error("gRPC error during doctor sign-in")
		utils.JSONResponse(w, err.Error(), http.StatusUnauthorized, req)
		return
	}

	d.Logger.WithFields(logrus.Fields{
		"function": "DoctorSignIn",
		"doctorId": resp.DoctorId,
		"status":   resp.Status,
	}).Info("Doctor signed in successfully")

	if resp.Status == "success" {
		jwtToken, err := middleware.CreateJWTToken(resp.DoctorId, role)
		if err != nil {
			d.Logger.WithFields(logrus.Fields{
				"function": "DoctorSignIn",
				"doctorId": resp.DoctorId,
				"error":    err.Error(),
			}).Error("Failed to create JWT token")
			utils.JSONResponse(w, "Failed to create JWT token", http.StatusInternalServerError, req)
			return
		}

		// Set JWT token in cookie
		middleware.SetJWTToken(w, jwtToken, role)
		d.Logger.Info("JWT token set in cookie for doctor")
	}

	utils.JSONResponse(w, resp, http.StatusOK, req)
	d.Logger.Info("Sign-in response sent")
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

	d.Logger.Info("Doctor logged out successfully")
	utils.JSONStandardResponse(w, "success", "", "Doctor logged out successfully", http.StatusOK, r)
}

func (d *DoctorServerClient) GetDoctorProfile(w http.ResponseWriter, req *http.Request) {
	d.Logger.Info("Received request to get doctor profile")

	claims, err := middleware.ExtractClaimsFromCookie(req, "doctor")
	if err != nil {
		d.Logger.WithFields(logrus.Fields{
			"function": "GetDoctorProfile",
			"error":    err.Error(),
		}).Error("Unauthorized access")
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, req)
		return
	}

	doctorId := claims.UserId
	d.Logger.WithFields(logrus.Fields{
		"function": "GetDoctorProfile",
		"doctorId": doctorId,
	}).Info("Fetching profile for doctor ID")

	// Call gRPC to fetch the profile
	resp, err := d.DoctorClient.GetProfile(context.Background(), &doctor.GetProfileRequest{
		DoctorId: doctorId,
	})
	if err != nil || resp.Status != "success" {
		d.Logger.WithFields(logrus.Fields{
			"function": "GetDoctorProfile",
			"doctorId": doctorId,
			"error":    err.Error(),
		}).Error("Failed to get profile")
		utils.JSONResponse(w, "Failed to get profile", http.StatusInternalServerError, req)
		return
	}

	d.Logger.WithFields(logrus.Fields{
		"function": "GetDoctorProfile",
		"doctorId": doctorId,
		"status":   resp.Status,
	}).Info("Successfully fetched profile for doctor")
	// Return profile information
	utils.JSONResponse(w, resp, http.StatusOK, req)
	d.Logger.Info("Doctor profile response sent")
}
func (d *DoctorServerClient) UpdateDoctorProfile(w http.ResponseWriter, req *http.Request) {
	d.Logger.Info("Received request to update doctor profile")

	claims, err := middleware.ExtractClaimsFromCookie(req, "doctor")
	if err != nil {
		d.Logger.WithFields(logrus.Fields{
			"function": "UpdateDoctorProfile",
			"error":    err.Error(),
		}).Error("Unauthorized access")
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, req)
		return
	}

	doctorId := claims.UserId
	d.Logger.WithFields(logrus.Fields{
		"function":  "UpdateDoctorProfile",
		"doctorId":  doctorId,
	}).Info("Parsing request body for profile update")

	// Parse request body to update profile
	var reqBody struct {
		Name             string `json:"name"`
		SpecializationId int32  `json:"specialization"`
		Phone            int    `json:"phone"`
	}
	err = json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		d.Logger.WithFields(logrus.Fields{
			"function": "UpdateDoctorProfile",
			"error":    err.Error(),
		}).Error("Invalid request body")
		utils.JSONResponse(w, "Invalid request body", http.StatusBadRequest, req)
		return
	}

	// Create doctor update request
	updateReq := &doctor.UpdateProfileRequest{
		Doctor: &doctor.Doctor{
			DoctorId:         doctorId,
			Name:             reqBody.Name,
			SpecializationId: reqBody.SpecializationId,
			Phone:            int32(reqBody.Phone),
		},
	}

	// Call gRPC service to update the profile
	resp, err := d.DoctorClient.UpdateProfile(context.Background(), updateReq)
	if err != nil || resp.Status != "success" {
		d.Logger.WithFields(logrus.Fields{
			"function":  "UpdateDoctorProfile",
			"doctorId":  doctorId,
			"error":     err.Error(),
		}).Error("Failed to update profile")
		utils.JSONResponse(w, "Failed to update profile", http.StatusInternalServerError, req)
		return
	}

	d.Logger.WithFields(logrus.Fields{
		"function":  "UpdateDoctorProfile",
		"doctorId":  doctorId,
		"status":    resp.Status,
	}).Info("Successfully updated doctor profile")

	// Return success response
	utils.JSONResponse(w, resp, http.StatusOK, req)
}

func (d *DoctorServerClient) DoctorStoreAccessToken(email string, token *oauth2.Token) error {
	d.Logger.WithFields(logrus.Fields{
		"function": "DoctorStoreAccessToken",
		"email":    email,
	}).Info("Storing access token for doctor")

	_, err := d.DoctorClient.StoreAccessToken(context.Background(), &doctor.StoreAccessTokenRequest{
		Email:        email,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry.String(),
	})
	if err != nil {
		d.Logger.WithFields(logrus.Fields{
			"function": "DoctorStoreAccessToken",
			"email":    email,
			"error":    err.Error(),
		}).Error("Failed to store access token")
		return err
	}

	d.Logger.WithFields(logrus.Fields{
		"function": "DoctorStoreAccessToken",
		"email":    email,
	}).Info("Access token stored successfully")
	return nil
}

func (d *DoctorServerClient) ConfirmScheduleHandler(w http.ResponseWriter, req *http.Request) {
	d.Logger.Info("Received request to confirm schedule")

	// Extract the doctor claims to verify authentication
	claims, err := middleware.ExtractClaimsFromCookie(req, "doctor")
	if err != nil {
		d.Logger.WithFields(logrus.Fields{
			"function": "ConfirmScheduleHandler",
			"error":    err.Error(),
		}).Error("Unauthorized access")
		utils.JSONResponse(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized, req)
		return
	}

	doctorID := claims.UserId // Assuming claims contain the doctor ID
	if doctorID == "" {
		d.Logger.WithFields(logrus.Fields{
			"function": "ConfirmScheduleHandler",
			"error":    "Missing doctor ID",
		}).Error("Missing doctor ID")
		utils.JSONResponse(w, "Missing doctor ID", http.StatusBadRequest, req)
		return
	}

	d.Logger.WithFields(logrus.Fields{
		"function":  "ConfirmScheduleHandler",
		"doctorId":  doctorID,
	}).Info("Making gRPC call to confirm schedule")

	// Make gRPC call to confirm the doctor's schedule
	grpcReq := &doctor.ConfirmScheduleRequest{
		DoctorId: doctorID,
	}

	grpcResp, err := d.DoctorClient.ConfirmSchedule(context.Background(), grpcReq)
	if err != nil {
		d.Logger.WithFields(logrus.Fields{
			"function": "ConfirmScheduleHandler",
			"doctorId": doctorID,
			"error":    err.Error(),
		}).Error("Failed to confirm schedule")
		utils.JSONStandardResponse(w, "fail", "Failed to confirm schedule: "+err.Error(), "", http.StatusInternalServerError, req)
		return
	}

	if grpcResp.Status != "success" {
		d.Logger.WithFields(logrus.Fields{
			"function":  "ConfirmScheduleHandler",
			"doctorId":  doctorID,
			"error":     grpcResp.Error,
		}).Error("Failed to confirm schedule: " + grpcResp.Error)
		utils.JSONResponse(w, "Failed to confirm schedule: "+grpcResp.Error, http.StatusInternalServerError, req)
		return
	}

	d.Logger.WithFields(logrus.Fields{
		"function":  "ConfirmScheduleHandler",
		"doctorId":  doctorID,
	}).Info("Schedule confirmed successfully")

	utils.JSONResponse(w, grpcResp.Schedules, http.StatusOK, req)
}
