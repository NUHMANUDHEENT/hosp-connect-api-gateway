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
	"github.com/nuhmanudheent/hosp-connect-api-gateway/logs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of requests processed",
		},
		[]string{"method"},
	)
)

func init() {
	prometheus.MustRegister(requestCount)
}

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

	// Wrap the router with CORS middleware
	corsHandler := di.CORS(router)

	http.Handle("/metrics", promhttp.Handler())
	log.Println("API Gateway running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
