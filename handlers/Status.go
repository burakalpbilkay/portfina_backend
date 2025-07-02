package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	dagID := mux.Vars(r)["dag_id"]
	//url := fmt.Sprintf("http://airflow-webserver:8080/api/v1/dags/%s/dagRuns?order_by=-execution_date&limit=5", dagID)
	url := fmt.Sprintf("http://host.docker.internal:8080/api/v1/dags/%s/dagRuns", dagID)

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
