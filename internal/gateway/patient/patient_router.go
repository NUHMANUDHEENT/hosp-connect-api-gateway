package patient

import (
	pb "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/patient"
	"github.com/gorilla/mux"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/di"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/appointment"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/middleware"
	"github.com/sirupsen/logrus"
)

type PatientServerClient struct {
	PatientClient pb.PatientServiceClient
	Logger        *logrus.Logger
}

func NewPatientClient(patientClient pb.PatientServiceClient, logger *logrus.Logger) *PatientServerClient {
	return &PatientServerClient{
		PatientClient: patientClient,
		Logger:        logger,
	}
}
func RegisterPatientRoutes(router *mux.Router, patientClient *PatientServerClient, AppointmentClient *appointment.AppointmentServerClient) {
	// Public routes
	publicRouter := router.PathPrefix("/api/v1/patient").Subrouter()
	publicRouter.HandleFunc("/signup", patientClient.PatientSignUp).Methods("POST")
	publicRouter.HandleFunc("/signup/verify-email", patientClient.SignUpVerify).Methods("GET")
	publicRouter.HandleFunc("/signin", patientClient.PatientSignIn).Methods("POST")
	publicRouter.HandleFunc("/logout", patientClient.PatientLogout).Methods("POST")
	publicRouter.HandleFunc("/help-desk/callback", di.HelpDeskHandler).Methods("POST")
	publicRouter.HandleFunc("/ws", patientClient.PatientChatHandler)

	// Private routes that require JWT middleware
	privateRouter := router.PathPrefix("/api/v1/patient").Subrouter()
	privateRouter.Use(middleware.JWTMiddleware("patient"))
	privateRouter.HandleFunc("/profile", patientClient.GetPatientProfile).Methods("GET")
	privateRouter.HandleFunc("/profile", patientClient.UpdatePatientProfile).Methods("PUT")
	privateRouter.HandleFunc("/get-availability", AppointmentClient.GetAvailability).Methods("GET")
	privateRouter.HandleFunc("/get-appointments", AppointmentClient.GetAppointments).Methods("GET")
	privateRouter.HandleFunc("/get-doctor-availability", AppointmentClient.GetAvailabilityByDoctorId).Methods("GET")
	privateRouter.HandleFunc("/confirm-appointment", AppointmentClient.ConfirmPatientAppointment).Methods("POST")
	privateRouter.HandleFunc("/cancel-appointment", AppointmentClient.CancelPatientAppointment).Methods("POST")
	privateRouter.HandleFunc("/get-prescription", patientClient.GetPrescriptions)
	privateRouter.HandleFunc("/help-desk", di.HelpDeskRender)
	privateRouter.HandleFunc("/customer-care", patientClient.PatientChatRender)
	privateRouter.HandleFunc("/video-call/{room}", patientClient.VideoCallRender)

}
