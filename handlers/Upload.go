package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func UploadHandler(targetFile string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20)
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "File upload error: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		if handler.Size > 10<<20 {
			http.Error(w, "File exceeds 10MB limit", http.StatusBadRequest)
			return
		}

		if !strings.HasSuffix(handler.Filename, ".csv") {
			http.Error(w, "Only .csv files are allowed", http.StatusBadRequest)
			return
		}

		_ = os.MkdirAll("/app/data/incoming", os.ModePerm)
		outputPath := fmt.Sprintf("/app/data/incoming/%s", targetFile)
		outFile, err := os.Create(outputPath)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, file)
		if err != nil {
			http.Error(w, "Failed to write file", http.StatusInternalServerError)
			return
		}

		dagID := strings.ToLower(strings.TrimSuffix(targetFile, ".csv")) + "_ingestion_dag"
		if !triggerDAG(dagID) {
			http.Error(w, fmt.Sprintf("File uploaded, but DAG trigger failed for %s", dagID), http.StatusBadGateway)
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{
			"message": "File uploaded and DAG triggered successfully",
			"dag":     dagID,
		})
	}
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
