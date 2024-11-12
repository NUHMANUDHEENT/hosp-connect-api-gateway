package config

import (
	"fmt"
	"log"
	"os"

	pbAdmin "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/admin"
	pbAppointment "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/appointment"
	pbDoctor "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/doctor"
	pbPatient "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/patient"
	pbPayment "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/payment"
	"github.com/gorilla/mux"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/admin"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/appointment"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/doctor"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/patient"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/payment"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/logs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func GrpcSetUp() *mux.Router {

	userConn, err := grpc.NewClient(os.Getenv("USER_GRPC_SERVER"), grpc.WithInsecure())
	if err != nil {
		log.Fatal("Failed to connect to user service:", err)
	}
	fmt.Println("users", os.Getenv("USER_GRPC_SERVER"))
	appointmentConn, err := grpc.NewClient(os.Getenv("APPOINTMENT_GRPC_SERVER"),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to connect to appointment service:", err)
	}

	paymentConn, err := grpc.NewClient(os.Getenv("PAYMENT_GRPC_SERVER"),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to connect to payment service:", err)
	}

	logger := logs.NewLogger()
	router := mux.NewRouter()
	adminClient := admin.NewAdminClient(pbAdmin.NewAdminServiceClient(userConn), logger)
	doctorClient := doctor.NewDoctorClient(pbDoctor.NewDoctorServiceClient(userConn), logger)
	patientClient := patient.NewPatientClient(pbPatient.NewPatientServiceClient(userConn), logger)
	appointmentClient := appointment.NewAppointmentClient(pbAppointment.NewAppointmentServiceClient(appointmentConn), logger)
	paymentClient := payment.NewPaymentClient(pbPayment.NewPaymentServiceClient(paymentConn), logger)

	admin.RegisterAdminRoutes(router, adminClient, appointmentClient)
	doctor.RegisterDoctorRoutes(router, doctorClient, patientClient, appointmentClient)
	patient.RegisterPatientRoutes(router, patientClient, appointmentClient)
	payment.RegisterPaymentRouters(router, paymentClient)
	return router
}
