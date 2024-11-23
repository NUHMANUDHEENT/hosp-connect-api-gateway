package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/NUHMANUDHEENT/hosp-connect-pb/proto/payment"
)

type RazorpayWebhookRequest struct {
	Entity    string          `json:"entity"`
	AccountID string          `json:"account_id"`
	Event     string          `json:"event"`
	Contains  []string        `json:"contains"`
	Payload   RazorpayPayload `json:"payload"`
	CreatedAt int64           `json:"created_at"`
}

type RazorpayPayload struct {
	Payment RazorpayPayment `json:"payment"`
}

type RazorpayPayment struct {
	Entity RazorpayPaymentEntity `json:"entity"`
}

type RazorpayPaymentEntity struct {
	ID          string                 `json:"id"`
	Entity      string                 `json:"entity"`
	Amount      int64                  `json:"amount"`
	Currency    string                 `json:"currency"`
	Status      string                 `json:"status"`
	OrderID     string                 `json:"order_id"`
	Method      string                 `json:"method"`
	Email       string                 `json:"email"`
	Contact     string                 `json:"contact"`
	Description string                 `json:"description"`
	Captured    bool                   `json:"captured"`
	CreatedAt   int64                  `json:"created_at"`
	Notes       map[string]interface{} `json:"notes"`
}

// LoadPaymentPage serves the payment page to the client
func (p *PaymentServerClient) LoadPaymentPage(w http.ResponseWriter, r *http.Request) {
	paymentPagePath := filepath.Join("templates", "payment.html")
	fmt.Println("hiiiii")
	http.ServeFile(w, r, paymentPagePath)
}
func (p *PaymentServerClient) PaymentCallBack(w http.ResponseWriter, r *http.Request) {
	var webhookReq RazorpayWebhookRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	// signature := r.Header.Get("x-razorpay-signature")
	// fmt.Println("sign", signature)
	// if !verifySignature("ciivOcvarUcV6uSV7WniDwfj", body, signature) {
	// 	log.Println("Invalid signature")
	// 	return
	// }

	// Unmarshal the JSON request
	if err := json.Unmarshal(body, &webhookReq); err != nil {
		log.Println("Failed to parse request body", http.StatusBadRequest)
		return
	}
	// Extract payment details
	paymentID := webhookReq.Payload.Payment.Entity.ID
	orderID := webhookReq.Payload.Payment.Entity.OrderID
	status := webhookReq.Payload.Payment.Entity.Status
	patientId := ""
	if webhookReq.Payload.Payment.Entity.Notes != nil {
		patientId, _ = webhookReq.Payload.Payment.Entity.Notes["patientId"].(string)
	}
	amount := webhookReq.Payload.Payment.Entity.Amount
	if status == "authorized" {
		return
	}
	log.Printf("Received Razorpay payment callback: OrderID: %s, PaymentID: %s,patientId: %s, Status: %s", orderID, paymentID, patientId, status)

	resp, err := p.PaymentClient.PaymentCallback(context.Background(), &payment.PaymentCallBackRequest{
		PaymentId: paymentID,
		Status:    status,
		OrderId:   orderID,
		PatientId: patientId,
		Amount:    float64(amount) / 100,
	})
	if err != nil {
		log.Printf("Order ID %v payment failed with status %v", orderID, status)
		return
	}

	log.Printf("Payment confirmation successful. Status: %s", resp)

	// Respond to Razorpay with a success status
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Webhook received"))
	if err != nil {
		log.Println("Failed to write response", err)
		return
	}

}

type PaymentSuccessResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	Status     string `json:"status"`
}

func (p *PaymentServerClient) PaymentSucces(w http.ResponseWriter, r *http.Request) {
	response := PaymentSuccessResponse{
		Message:    "Payment processed successfully!",
		StatusCode: http.StatusOK,
		Status:     "success",
	}

	// Set the response header to indicate JSON content
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Set the HTTP status code

	// Encode the response to JSON and send it
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
