package admin

import (
	pb "github.com/NUHMANUDHEENT/hosp-connect-pb/proto/admin"
	"github.com/gorilla/mux"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/middleware"
)

type AdminServerClient struct {
	pb.AdminServiceClient
}

func RegisterAdminRoutes(router *mux.Router, adminclient *AdminServerClient) {
	publicRouter := router.PathPrefix("/api/v1/admin").Subrouter()
	publicRouter.HandleFunc("/signin", adminclient.AdminSignIn).Methods("POST")
	publicRouter.HandleFunc("/logout", adminclient.AdminLogout).Methods("POST")

	privateRouter := router.PathPrefix("/api/v1/admin").Subrouter()
	privateRouter.Use(middleware.JWTMiddleware("admin"))
	privateRouter.HandleFunc("/doctor/register", adminclient.DoctorRegister).Methods("POST")
	privateRouter.HandleFunc("/doctor/delete/{ID}", adminclient.DoctorDelete).Methods("POST")
	privateRouter.HandleFunc("/patient/register", adminclient.PatientCreate).Methods("POST")
	privateRouter.HandleFunc("/patient/delete/{ID}", adminclient.PatientDelete).Methods("POST")
	privateRouter.HandleFunc("/patient/block/{ID}", adminclient.PatientBlock)
	privateRouter.HandleFunc("/patient/list", adminclient.ListPatientsHandler)
	privateRouter.HandleFunc("/doctor/list", adminclient.ListDoctorsHandler)
	privateRouter.HandleFunc("/doctor/addcategory", adminclient.AddDoctorSpecialization).Methods("POST")

}
