package doctor

import (
	"net/http"

	pb "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/doctor"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/appointment"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/patient"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/middleware"
)

type DoctorServerClient struct {
	DoctorClient pb.DoctorServiceClient
	Logger       *logrus.Logger
}

func NewDoctorClient(doctorClient pb.DoctorServiceClient, logger *logrus.Logger) *DoctorServerClient {
	return &DoctorServerClient{
		DoctorClient: doctorClient,
		Logger:       logger,
	}
}
func RegisterDoctorRoutes(router *mux.Router, DoctorClient *DoctorServerClient, PatientClient *patient.PatientServerClient, AppointmentClient *appointment.AppointmentServerClient) {
	publicRouter := router.PathPrefix("/api/v1/doctor").Subrouter()
	publicRouter.HandleFunc("/signin", DoctorClient.DoctorSignIn).Methods("POST")
	publicRouter.HandleFunc("/logout", DoctorClient.DoctorLogout).Methods("POST")
	publicRouter.HandleFunc("/auth/login", DoctorClient.HandleGoogleLogin).Methods("GET")
	publicRouter.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		DoctorClient.HandleGoogleCallback(w, r)
	}).Methods("GET")

	privateRouter := router.PathPrefix("/api/v1/doctor").Subrouter()
	privateRouter.Use(middleware.JWTMiddleware("doctor"))
	privateRouter.HandleFunc("/profile", DoctorClient.GetDoctorProfile).Methods("GET")
	privateRouter.HandleFunc("/profile", DoctorClient.UpdateDoctorProfile).Methods("PUT")
	privateRouter.HandleFunc("/add-prescription", PatientClient.AddPrescriptionForPatient).Methods("POST")
	privateRouter.HandleFunc("/get-prescription", PatientClient.GetPrescriptionsForDoctors)
	privateRouter.HandleFunc("/schedule/confirm", DoctorClient.ConfirmScheduleHandler).Methods("POST")
	privateRouter.HandleFunc("/video-room-create", AppointmentClient.CreateRoomForVideoTreatments).Methods("POST")
	privateRouter.HandleFunc("/video-call/{room}", AppointmentClient.VideoCallRender)
}
