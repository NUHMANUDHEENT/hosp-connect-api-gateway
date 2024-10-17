package appointment

import (
	"github.com/NUHMANUDHEENT/hosp-connect-pb/proto/appointment"
	"github.com/gorilla/mux"
)

type AppointmentServerClient struct {
	appointment.AppointmentServiceClient
}
func RegisterAppointmentRouters(router *mux.Router, appointmentClient *AppointmentServerClient){
	// router.HandleFunc("/appointment", appointmentClient.GetAppointment).Methods("GET")
}