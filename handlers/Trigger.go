package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func TriggerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	dagID := mux.Vars(r)["dag"]
	if dagID == "" {
		http.Error(w, "DAG ID is required", http.StatusBadRequest)
		return
	}

	if !triggerDAG(dagID) {
		http.Error(w, "Failed to trigger DAG", http.StatusBadGateway)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("DAG %s triggered successfully", dagID),
	})
}

func triggerDAG(dagID string) bool {
	url := fmt.Sprintf("http://host.docker.internal:8080/api/v1/dags/%s/dagRuns", dagID)

	payload, _ := json.Marshal(map[string]interface{}{
		"conf": map[string]interface{}{},
	})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return false
	}

	auth := base64.StdEncoding.EncodeToString([]byte("admin:admin"))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error triggering DAG:", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var body bytes.Buffer
		_, _ = body.ReadFrom(resp.Body)
		fmt.Printf("DAG trigger failed: %s\nResponse: %s\n", resp.Status, body.String())
		return false
	}

	return true
}
