package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// handleLogin processes login requests and issues JWT tokens
func handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Simple authentication for demo purposes
	// In production, use a secure authentication method
	if creds.Username == "admin" && creds.Password == "password" {
		token, err := GenerateJWT(creds.Username, "admin")
		if err != nil {
			http.Error(w, "Token generation failed", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
		return
	}
	
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// handleIngest handles ingestion of data into ClickHouse
func handleIngest(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Value string `json:"value"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if err := insertEvent(payload.Value); err != nil {
		http.Error(w, "ClickHouse insert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	fmt.Fprintf(w, "Ingested: %s", payload.Value)
}

// handleGetTables returns a list of tables from ClickHouse
func handleGetTables(w http.ResponseWriter, r *http.Request) {
	var config ClickHouseConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	db, err := getClickHouseConnection(config)
	if err != nil {
		http.Error(w, "Failed to connect: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()
	
	tables, err := getClickHouseTables(db)
	if err != nil {
		http.Error(w, "Failed to get tables: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tables)
}

// handleGetTableSchema returns the schema for a ClickHouse table
func handleGetTableSchema(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Config    ClickHouseConfig `json:"config"`
		TableName string           `json:"table_name"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	db, err := getClickHouseConnection(req.Config)
	if err != nil {
		http.Error(w, "Failed to connect: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()
	
	schema, err := getTableSchema(db, req.TableName)
	if err != nil {
		http.Error(w, "Failed to get schema: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

// handleGetFlatFileSchema returns the schema for a flat file
func handleGetFlatFileSchema(w http.ResponseWriter, r *http.Request) {
	var config FlatFileConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	schema, err := getFlatFileSchema(config)
	if err != nil {
		http.Error(w, "Failed to get schema: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

// handleStartIngest handles the main ingestion process
func handleStartIngest(w http.ResponseWriter, r *http.Request) {
	var request IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	var result *IngestResult
	var err error
	
	switch request.SourceType {
	case "clickhouse":
		// ClickHouse to Flat File
		db, err := getClickHouseConnection(request.ClickHouseConfig)
		if err != nil {
			http.Error(w, "Failed to connect to ClickHouse: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()
		
		result, err = exportToFile(db, request)
		
	case "flatfile":
		// Flat File to ClickHouse
		db, err := getClickHouseConnection(request.ClickHouseConfig)
		if err != nil {
			http.Error(w, "Failed to connect to ClickHouse: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()
		
		result, err = importFromFile(db, request)
		
	default:
		http.Error(w, "Invalid source type", http.StatusBadRequest)
		return
	}
	
	if err != nil {
		http.Error(w, "Ingestion failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handlePreviewData returns a preview of the data source
func handlePreviewData(w http.ResponseWriter, r *http.Request) {
	var request struct {
		SourceType       string           `json:"source_type"`
		ClickHouseConfig ClickHouseConfig `json:"clickhouse_config,omitempty"`
		FlatFileConfig   FlatFileConfig   `json:"flat_file_config,omitempty"`
		TableName        string           `json:"table_name,omitempty"`
		Columns          []string         `json:"columns,omitempty"`
		Limit            int              `json:"limit"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if request.Limit <= 0 {
		request.Limit = 100 // Default limit
	}
	
	var result *PreviewResult
	var err error
	
	switch request.SourceType {
	case "clickhouse":
		var db *sql.DB
		db, err = getClickHouseConnection(request.ClickHouseConfig)
		if err != nil {
			break
		}
		defer db.Close()
		
		result, err = previewClickHouseData(db, request.TableName, request.Columns, request.Limit)
		
	case "flatfile":
		result, err = previewFlatFileData(request.FlatFileConfig, request.Limit)
		
	default:
		err = fmt.Errorf("invalid source type")
	}
	
	if err != nil {
		http.Error(w, "Preview failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}