package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	utils "go-airflow-trigger/Utils"
)

func GetBondEnrichmentStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	airflowURL := fmt.Sprintf("http://%s:8080/api/v1/dags/bond_enrichment_dag/dagRuns", "host.docker.internal")

	req, err := http.NewRequest("GET", airflowURL, nil)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to create request: %v", err)})
		return
	}

	auth := base64.StdEncoding.EncodeToString([]byte("admin:admin"))
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		respondJSON(w, http.StatusBadGateway, map[string]string{"error": "Failed to fetch DAG statusss" + err.Error()})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to parse response"})
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func TriggerBondEnrichment(w http.ResponseWriter, r *http.Request) {
	airflowURL := fmt.Sprintf("http://%s:8080/api/v1/dags/bond_enrichment_dag/dagRuns", "host.docker.internal")

	payload := []byte(`{"conf": {}}`)

	req, err := http.NewRequest("POST", airflowURL, bytes.NewBuffer(payload))
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to create request: %v", err)})
		return
	}

	req.SetBasicAuth("admin", "admin")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, map[string]string{"error": fmt.Sprintf("Failed to trigger DAG: %v", err)})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Write(body)
}

func GetBondResults(w http.ResponseWriter, r *http.Request) {
	db, err := utils.Connect()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Database connection failed"})
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM bond_results")
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to query bond_results"})
		return
	}
	defer rows.Close()

	var results []map[string]interface{}
	cols, _ := rows.Columns()

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			continue
		}

		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})

			switch v := (*val).(type) {
			case []byte:
				strVal := string(v)
				if f, err := strconv.ParseFloat(strVal, 64); err == nil {
					rowMap[colName] = f
				} else {
					rowMap[colName] = strVal
				}
			default:
				rowMap[colName] = v
			}
		}
		results = append(results, rowMap)
	}

	respondJSON(w, http.StatusOK, results)
}

func GetBondCashflows(w http.ResponseWriter, r *http.Request) {
	db, err := utils.Connect()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Database connection failed"})
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM bond_cashflows")
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to query bond_cashflows"})
		return
	}
	defer rows.Close()

	var cashflows []map[string]interface{}
	cols, _ := rows.Columns()

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			continue
		}

		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})

			switch v := (*val).(type) {
			case []byte:
				strVal := string(v)
				if f, err := strconv.ParseFloat(strVal, 64); err == nil {
					rowMap[colName] = f
				} else {
					rowMap[colName] = strVal
				}
			default:
				rowMap[colName] = v
			}
		}
		cashflows = append(cashflows, rowMap)
	}

	respondJSON(w, http.StatusOK, cashflows)
}
