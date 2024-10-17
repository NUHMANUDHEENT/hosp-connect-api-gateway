package payment

import (
	"github.com/NUHMANUDHEENT/hosp-connect-pb/proto/payment"
	"github.com/gorilla/mux"
)

type PaymentServerClient struct {
	payment.PaymentServiceClient
}

func RegisterPaymentRouters(router *mux.Router, paymentClient *PaymentServerClient) {
	privateRouter := router.PathPrefix("/api/v1/payment").Subrouter()
	privateRouter.HandleFunc("", paymentClient.LoadPaymentPage).Methods("GET")
	privateRouter.HandleFunc("/callback", paymentClient.PaymentCallBack)
}
