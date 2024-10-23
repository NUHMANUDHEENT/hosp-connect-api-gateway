package patient

import (
	pb "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/patient"
	"github.com/gorilla/mux"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/di"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/appointment"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/middleware"
)

type PatientServerClient struct {
	pb.PatientServiceClient
}

func RegisterPatientRoutes(router *mux.Router, patientClient *PatientServerClient, appointmentClient *appointment.AppointmentServerClient) {
	// Public routes
	publicRouter := router.PathPrefix("/api/v1/patient").Subrouter()
	publicRouter.HandleFunc("/signup", patientClient.PatientSignUp).Methods("POST")
	publicRouter.HandleFunc("/signup/verify-email", patientClient.SignUpverify).Methods("GET")
	publicRouter.HandleFunc("/signin", patientClient.PatientSignIn).Methods("POST")
	publicRouter.HandleFunc("/logout", patientClient.PatientLogout).Methods("POST")
	publicRouter.HandleFunc("/help-desk/callback", di.HelpDeskHandler).Methods("POST")
	publicRouter.HandleFunc("/video-call", patientClient.VideoCallRender)
	publicRouter.HandleFunc("/ws", patientClient.PatientChatHandler)
	
	// Private routes that require JWT middleware
	privateRouter := router.PathPrefix("/api/v1/patient").Subrouter()
	privateRouter.Use(middleware.JWTMiddleware("patient"))
	privateRouter.HandleFunc("/profile", patientClient.GetPatientProfile).Methods("GET")
	privateRouter.HandleFunc("/profile", patientClient.UpdatePatientProfile).Methods("PUT")
	privateRouter.HandleFunc("/get-availability", appointmentClient.GetAvailability).Methods("GET")
	privateRouter.HandleFunc("/get-appointments", appointmentClient.GetAppointments).Methods("GET")
	privateRouter.HandleFunc("/get-doctor-availability", appointmentClient.GetAvailabilityByDoctorId).Methods("GET")
	privateRouter.HandleFunc("/confirm-appointment", appointmentClient.ConfirmPatientAppointment).Methods("POST")
	privateRouter.HandleFunc("/get-prescription", patientClient.GetPrescriptions)
	privateRouter.HandleFunc("/help-desk", di.HelpDeskRender)
	privateRouter.HandleFunc("/customer-care", patientClient.PatientChatRender)
}
