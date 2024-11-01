package appointment

import (
	"github.com/NUHMANUDHEENT/hosp-connect-pb/proto/appointment"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type AppointmentServerClient struct {
	appointment.AppointmentServiceClient
	Logger *logrus.Logger
}

func NewAppointmentClient(appointmentClient appointment.AppointmentServiceClient, logger *logrus.Logger) *AppointmentServerClient {
	return &AppointmentServerClient{
		AppointmentServiceClient: appointmentClient,
		Logger:                   logger,
	}
}
func RegisterAppointmentRouters(router *mux.Router, appointmentClient *AppointmentServerClient) {
	// router.HandleFunc("/appointment", appointmentClient.GetAppointment).Methods("GET")
}
