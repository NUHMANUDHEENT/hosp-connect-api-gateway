package main

import (
	"log"
	"net/http"

	pbAdmin "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/admin"
	pbAppointment "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/appointment"
	pbDoctor "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/doctor"
	pbPatient "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/patient"
	pbPayment "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/payment"
	"github.com/gorilla/mux"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/di"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/admin"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/appointment"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/doctor"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/patient"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/gateway/payment"
	"google.golang.org/grpc"
)

func main() {
	userConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal("Failed to connect to admin service:", err)
	}
	appointmentConn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatal("Failed to connect to appointment service:", err)
	}
	paymentConn, err := grpc.Dial("localhost:50054", grpc.WithInsecure())
	if err != nil {
		log.Fatal("Failed to connect to payment service:", err)
	}
	defer userConn.Close()
	defer appointmentConn.Close()
	defer paymentConn.Close()

	router := mux.NewRouter()
	adminClient := &admin.AdminServerClient{
		AdminServiceClient: pbAdmin.NewAdminServiceClient(userConn),
	}
	doctorClient := &doctor.DoctorServerClient{
		DoctorServiceClient: pbDoctor.NewDoctorServiceClient(userConn),
	}
	patientClient := &patient.PatientServerClient{
		PatientServiceClient: pbPatient.NewPatientServiceClient(userConn),
	}
	appointmentClient := &appointment.AppointmentServerClient{
		AppointmentServiceClient: pbAppointment.NewAppointmentServiceClient(appointmentConn),
	}
	paymentClient := &payment.PaymentServerClient{
		PaymentServiceClient: pbPayment.NewPaymentServiceClient(paymentConn),
	}

	admin.RegisterAdminRoutes(router, adminClient)
	doctor.RegisterDoctorRoutes(router, doctorClient, patientClient,appointmentClient)
	patient.RegisterPatientRoutes(router, patientClient, appointmentClient)
	payment.RegisterPaymentRouters(router, paymentClient)

	// Wrap the router with CORS middleware
	corsHandler := di.CORS(router)

	log.Println("API Gateway running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
