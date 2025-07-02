package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"go-airflow-trigger/handlers"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env for local development
	_ = godotenv.Load()

	r := mux.NewRouter()

	// Upload endpoints for each CSV
	r.HandleFunc("/upload/bond", handlers.UploadHandler("Bond.csv")).Methods("POST")
	r.HandleFunc("/upload/interest-rate", handlers.UploadHandler("InterestRate.csv")).Methods("POST")
	r.HandleFunc("/upload/inflation-index", handlers.UploadHandler("InflationIndex.csv")).Methods("POST")
	r.HandleFunc("/upload/inflation-expectation", handlers.UploadHandler("InflationExpectation.csv")).Methods("POST")
	r.HandleFunc("/upload/exchange-rate", handlers.UploadHandler("ExchangeRate.csv")).Methods("POST")
	r.HandleFunc("/upload/foreign-exchange", handlers.UploadHandler("ForeignExchange.csv")).Methods("POST")
	r.HandleFunc("/upload/forward-curve", handlers.UploadHandler("ForwardCurve.csv")).Methods("POST")
	// Bond Enrichment DAG
	r.HandleFunc("/status/bond_enrichment_dag", handlers.GetBondEnrichmentStatus).Methods("GET")
	r.HandleFunc("/trigger/bond_enrichment_dag", handlers.TriggerBondEnrichment).Methods("POST")
	// Bond Results & Cashflows
	r.HandleFunc("/bond_results", handlers.GetBondResults).Methods("GET")
	r.HandleFunc("/bond_cashflows", handlers.GetBondCashflows).Methods("GET")
	// Trigger DAGs
	r.HandleFunc("/trigger/{dag}", handlers.TriggerHandler).Methods("POST")

	// DAG status
	r.HandleFunc("/status/{dag_id}", handlers.StatusHandler).Methods("GET")

	// Bond query
	r.HandleFunc("/bonds", handlers.GetBondsHandler).Methods("GET")
	http.Handle("/", corsMiddleware(r))

	log.Println("Go DAG Trigger API running on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
