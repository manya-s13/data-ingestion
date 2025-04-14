package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No .env file loaded, using defaults or environment variables: %v", err)
	}

	url := getEnv("CLICKHOUSE_URL", "tcp://localhost:9000")
	db := getEnv("CLICKHOUSE_DB", "default")
	user := getEnv("CLICKHOUSE_USER", "default")
	pass := getEnv("CLICKHOUSE_PASS", "")
	jwt := os.Getenv("CLICKHOUSE_JWT")

	dsn := fmt.Sprintf("%s?database=%s&username=%s&password=%s", url, db, user, pass)
	if jwt != "" {
		dsn += fmt.Sprintf("&auth_jwt=%s", jwt)
	}

	conn, err := sql.Open("clickhouse", dsn)
	if err != nil {
		log.Fatal("Failed to open ClickHouse connection: ", err)
	}
	defer conn.Close()

	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(0)

	if err := conn.Ping(); err != nil {
		log.Fatal("Failed to ping ClickHouse: ", err)
	}
	log.Println("Connected to ClickHouse!")

	r := mux.NewRouter()
	r.HandleFunc("/ingest", handleIngest).Methods("POST")

	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./build"))))

	log.Printf("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

	func handleIngest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Ingest endpoint - implementation pending"))
}