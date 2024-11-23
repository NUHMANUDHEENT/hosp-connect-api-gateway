package payment

import (
	"github.com/NUHMANUDHEENT/hosp-connect-pb/proto/payment"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type PaymentServerClient struct {
	PaymentClient payment.PaymentServiceClient
	logger        *logrus.Logger
}

func NewPaymentClient(paymentClient payment.PaymentServiceClient, logger *logrus.Logger) *PaymentServerClient {
	return &PaymentServerClient{
		PaymentClient: paymentClient,
		logger:        logger,
	}

}
func RegisterPaymentRouters(router *mux.Router, paymentClient *PaymentServerClient) {
	router.HandleFunc("/payment-success", paymentClient.PaymentSucces)
	privateRouter := router.PathPrefix("/api/v1/payment").Subrouter()
	privateRouter.HandleFunc("", paymentClient.LoadPaymentPage).Methods("GET")
	privateRouter.HandleFunc("/callback", paymentClient.PaymentCallBack)
}
