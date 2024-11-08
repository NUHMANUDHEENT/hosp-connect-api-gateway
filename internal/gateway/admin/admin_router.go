package admin

import (
	pb "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/admin"
	"github.com/gorilla/mux"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/appointment"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/middleware"
	"github.com/sirupsen/logrus"
)

type AdminServerClient struct {
	AdminClient pb.AdminServiceClient
	Logger      *logrus.Logger
}

func NewAdminClient(adminClient pb.AdminServiceClient, logger *logrus.Logger) *AdminServerClient {
	return &AdminServerClient{
		AdminClient: adminClient,
		Logger:      logger,
	}
}

func RegisterAdminRoutes(router *mux.Router, adminclient *AdminServerClient, appointmentClient *appointment.AppointmentServerClient) {
	publicRouter := router.PathPrefix("/api/v1/admin").Subrouter()
	publicRouter.HandleFunc("/signin", adminclient.AdminSignIn).Methods("POST")
	publicRouter.HandleFunc("/logout", adminclient.AdminLogout).Methods("POST")
	publicRouter.HandleFunc("/ws", adminclient.CustomerCareChatHandler)

	privateRouter := router.PathPrefix("/api/v1/admin").Subrouter()
	privateRouter.Use(middleware.JWTMiddleware("admin"))
	privateRouter.HandleFunc("/doctor/register", adminclient.DoctorRegister).Methods("POST")
	privateRouter.HandleFunc("/doctor/delete/{ID}", adminclient.DoctorDelete).Methods("DELETE")
	privateRouter.HandleFunc("/patient/register", adminclient.PatientCreate).Methods("POST")
	privateRouter.HandleFunc("/patient/delete/{ID}", adminclient.PatientDelete).Methods("DELETE")
	privateRouter.HandleFunc("/patient/block/{ID}", adminclient.PatientBlock)
	privateRouter.HandleFunc("/patient/list", adminclient.ListPatientsHandler)
	privateRouter.HandleFunc("/doctor/list", adminclient.ListDoctorsHandler)
	privateRouter.HandleFunc("/doctor/addcategory", appointmentClient.AddDoctorSpecialization).Methods("POST")
	privateRouter.HandleFunc("/customer-support", adminclient.AdminChatRender)
	privateRouter.HandleFunc("/dashboard", appointmentClient.Dashboard)
	privateRouter.HandleFunc("/dashboard/fetch", appointmentClient.DashboardResponse)

}
