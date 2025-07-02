package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func main_old() {

	r := mux.NewRouter()

	// DAG trigger endpoints
	r.HandleFunc("/trigger/bond", triggerHandler("bond_ingestion_dag")).Methods("POST")
	r.HandleFunc("/trigger/interest-rate", triggerHandler("interest_rate_ingestion_dag")).Methods("POST")
	r.HandleFunc("/trigger/inflation-index", triggerHandler("inflation_index_ingestion_dag")).Methods("POST")
	r.HandleFunc("/trigger/inflation-expectation", triggerHandler("inflation_expectation_ingestion_dag")).Methods("POST")
	r.HandleFunc("/trigger/exchange-rate", triggerHandler("exchange_rate_ingestion_dag")).Methods("POST")
	r.HandleFunc("/trigger/foreign-exchange", triggerHandler("foreign_exchange_ingestion_dag")).Methods("POST")
	r.HandleFunc("/trigger/forward-curve", triggerHandler("forward_curve_ingestion_dag")).Methods("POST")

	// File upload + auto-trigger
	r.HandleFunc("/upload/bond", uploadHandler("Bond.csv")).Methods("POST")
	r.HandleFunc("/upload/interest-rate", uploadHandler("InterestRate.csv")).Methods("POST")
	r.HandleFunc("/upload/inflation-index", uploadHandler("InflationIndex.csv")).Methods("POST")
	r.HandleFunc("/upload/inflation-expectation", uploadHandler("InflationExpectation.csv")).Methods("POST")
	r.HandleFunc("/upload/exchange-rate", uploadHandler("ExchangeRate.csv")).Methods("POST")
	r.HandleFunc("/upload/foreign-exchange", uploadHandler("ForeignExchange.csv")).Methods("POST")
	r.HandleFunc("/upload/forward-curve", uploadHandler("ForwardCurve.csv")).Methods("POST")

	r.HandleFunc("/status/{dag_id}", dagStatusHandler).Methods("GET")
	r.HandleFunc("/bonds", getBondsHandler).Methods("GET")

	fmt.Println("Go DAG Trigger API running at http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

func triggerHandler(dagID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		success := triggerDAG(dagID)
		if success {
			respondJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("DAG %s triggered successfully", dagID)})
		} else {
			respondJSON(w, http.StatusBadGateway, map[string]string{"error": fmt.Sprintf("Failed to trigger DAG %s", dagID)})
		}
	}
}

func uploadHandler(targetFile string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20) // 10MB

		file, handler, err := r.FormFile("file")
		if err != nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "file upload error: " + err.Error()})
			return
		}
		defer file.Close()

		if handler.Size > 10<<20 {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "file exceeds 10MB limit"})
			return
		}

		if !strings.HasSuffix(handler.Filename, ".csv") {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "only .csv files are allowed"})
			return
		}

		outputPath := fmt.Sprintf("data/incoming/%s", targetFile)
		outFile, err := os.Create(outputPath)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save file"})
			return
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, file)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to write file"})
			return
		}

		dagMap := map[string]string{
			"Bond.csv":                 "bond_ingestion_dag",
			"InterestRate.csv":         "interest_rate_ingestion_dag",
			"InflationIndex.csv":       "inflation_index_ingestion_dag",
			"InflationExpectation.csv": "inflation_expectation_ingestion_dag",
			"ExchangeRate.csv":         "exchange_rate_ingestion_dag",
			"ForeignExchange.csv":      "foreign_exchange_ingestion_dag",
			"ForwardCurve.csv":         "forward_curve_ingestion_dag",
		}
		dagID := dagMap[targetFile]
		if dagID == "" {
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error": "No DAG mapped for file: " + targetFile,
			})
			return
		}

		if triggerDAG(dagID) {
			respondJSON(w, http.StatusOK, map[string]string{
				"message": "file uploaded and DAG triggered successfully",
				"dag":     dagID,
			})
		} else {
			respondJSON(w, http.StatusBadGateway, map[string]string{
				"error": "file uploaded, but DAG trigger failed",
				"dag":   dagID,
			})
		}
	}
}

func triggerDAG(dagID string) bool {
	url := fmt.Sprintf("http://localhost:8080/api/v1/dags/%s/dagRuns", dagID)
	payload, _ := json.Marshal(map[string]interface{}{"conf": map[string]interface{}{}})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return false
	}

	auth := base64.StdEncoding.EncodeToString([]byte("admin:admin"))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func dagStatusHandler(w http.ResponseWriter, r *http.Request) {
	dagID := mux.Vars(r)["dag_id"]
	url := fmt.Sprintf("http://localhost:8080/api/v1/dags/%s/dagRuns?order_by=-execution_date&limit=5", dagID)

	req, _ := http.NewRequest("GET", url, nil)
	auth := base64.StdEncoding.EncodeToString([]byte("admin:admin"))
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		respondJSON(w, http.StatusBadGateway, map[string]string{"error": "Failed to fetch DAG status"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	respondJSON(w, http.StatusOK, result)
}

func getBondsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("postgres", "host=postgres port=5432 user=airflow password=airflow dbname=finance_db sslmode=disable")
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Database connection failed"})
		return
	}
	defer db.Close()

	rows, err := db.Query(`
        SELECT id, type, coupon_rate, currency, buy_sell, discount_curve, maturity_date, term, notional, clean_price,
               issue_date, inflation_curve, inflation_expectation_curve, underlying_curve_index, last_reset_rate,
               interest_rate, theoretical_model, compounding, accrual_day_count_convention, day_count_convention,
               calendar, indexation, created_at
        FROM bonds
    `)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to query bonds"})
		return
	}
	defer rows.Close()

	var bonds []map[string]interface{}

	for rows.Next() {
		var (
			id, bondType, currency, buySell, discountCurve, term, inflationCurve,
			inflationExpectationCurve, underlyingCurveIndex, theoreticalModel,
			compounding, accrualConvention, dayCountConvention, calendar, indexation string

			couponRate, notional, cleanPrice, lastResetRate, interestRate float64
			maturityDate, issueDate, createdAt                            time.Time
		)

		err := rows.Scan(
			&id, &bondType, &couponRate, &currency, &buySell, &discountCurve,
			&maturityDate, &term, &notional, &cleanPrice,
			&issueDate, &inflationCurve, &inflationExpectationCurve, &underlyingCurveIndex,
			&lastResetRate, &interestRate, &theoreticalModel, &compounding,
			&accrualConvention, &dayCountConvention, &calendar, &indexation, &createdAt,
		)
		if err != nil {
			continue
		}

		bond := map[string]interface{}{
			"id":                           id,
			"type":                         bondType,
			"coupon_rate":                  couponRate,
			"currency":                     currency,
			"buy_sell":                     buySell,
			"discount_curve":               discountCurve,
			"maturity_date":                maturityDate.Format("2006-01-02"),
			"term":                         term,
			"notional":                     notional,
			"clean_price":                  cleanPrice,
			"issue_date":                   issueDate.Format("2006-01-02"),
			"inflation_curve":              inflationCurve,
			"inflation_expectation_curve":  inflationExpectationCurve,
			"underlying_curve_index":       underlyingCurveIndex,
			"last_reset_rate":              lastResetRate,
			"interest_rate":                interestRate,
			"theoretical_model":            theoreticalModel,
			"compounding":                  compounding,
			"accrual_day_count_convention": accrualConvention,
			"day_count_convention":         dayCountConvention,
			"calendar":                     calendar,
			"indexation":                   indexation,
			"created_at":                   createdAt.Format(time.RFC3339),
		}

		bonds = append(bonds, bond)
	}

	respondJSON(w, http.StatusOK, bonds)
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(payload)
}
