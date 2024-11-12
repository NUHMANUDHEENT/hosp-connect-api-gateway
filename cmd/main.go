package main

import (
	"log"
	"net/http"
	"os"

	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/config"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/di"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	config.LoadEnv()
	port := os.Getenv("SERVER_PORT")
	router := config.GrpcSetUp()
	corsHandler := di.CORS(router)
	http.Handle("/metrics", promhttp.Handler())
	log.Println("API Gateway running on port 8080")
	log.Fatal(http.ListenAndServe(port, corsHandler))
}
