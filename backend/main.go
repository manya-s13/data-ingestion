// package main

// import (
// 	"database/sql"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"strings"
// 	"time"

// 	"github.com/golang-jwt/jwt/v5"
// 	"github.com/gorilla/mux"
// 	"github.com/joho/godotenv"
// 	_ "github.com/ClickHouse/clickhouse-go/v2"
// )

// var conn *sql.DB // shared connection
// var jwtSecret []byte

// func main() {
// 	if err := godotenv.Load(); err != nil {
// 		log.Println("No .env file found")
// 	}

// 	url := getEnv("CLICKHOUSE_URL", "tcp://localhost:9000")
// 	db := getEnv("CLICKHOUSE_DB", "default")
// 	user := getEnv("CLICKHOUSE_USER", "default")
// 	pass := getEnv("CLICKHOUSE_PASS", "")
// 	jwtSecret = []byte(getEnv("JWT_SECRET", ""))

// 	if len(jwtSecret) == 0 {
// 		log.Fatal("JWT_SECRET is required")
// 	}

// 	dsn := fmt.Sprintf("%s?database=%s&username=%s&password=%s", url, db, user, pass)
// 	var err error
// 	conn, err = sql.Open("clickhouse", dsn)
// 	if err != nil {
// 		log.Fatal("ClickHouse connection failed:", err)
// 	}
// 	defer conn.Close()

// 	if err := conn.Ping(); err != nil {
// 		log.Fatal("ClickHouse ping failed:", err)
// 	}
// 	log.Println("Connected to ClickHouse")

// 	r := mux.NewRouter()
// 	r.HandleFunc("/login", handleLogin).Methods("POST")
// 	r.HandleFunc("/ingest", authMiddleware(handleIngest)).Methods("POST")

// 	log.Println("Server running on :8080")
// 	log.Fatal(http.ListenAndServe(":8080", r))
// }

// func getEnv(key, fallback string) string {
// 	if val, ok := os.LookupEnv(key); ok {
// 		return val
// 	}
// 	return fallback
// }

// // ---------------- JWT HANDLING ----------------

// func GenerateJWT(username, role string) (string, error) {
// 	claims := jwt.MapClaims{
// 		"username": username,
// 		"role":     role,
// 		"exp":      time.Now().Add(time.Hour * 24).Unix(),
// 	}
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	return token.SignedString(jwtSecret)
// }

// func ValidateJWT(tokenStr string) (*jwt.Token, error) {
// 	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
// 		return jwtSecret, nil
// 	})
// }

// // ---------------- ROUTES ----------------

// func handleLogin(w http.ResponseWriter, r *http.Request) {
// 	var creds struct {
// 		Username string `json:"username"`
// 		Password string `json:"password"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
// 		http.Error(w, "Invalid JSON", http.StatusBadRequest)
// 		return
// 	}

// 	if creds.Username == "admin" && creds.Password == "password" {
// 		token, err := GenerateJWT(creds.Username, "admin")
// 		if err != nil {
// 			http.Error(w, "Token generation failed", http.StatusInternalServerError)
// 			return
// 		}
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(map[string]string{"token": token})
// 		return
// 	}

// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
// }

// func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		auth := r.Header.Get("Authorization")
// 		if !strings.HasPrefix(auth, "Bearer ") {
// 			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
// 			return
// 		}
// 		tokenStr := strings.TrimPrefix(auth, "Bearer ")
// 		token, err := ValidateJWT(tokenStr)
// 		if err != nil || !token.Valid {
// 			http.Error(w, "Invalid token", http.StatusUnauthorized)
// 			return
// 		}
// 		next(w, r)
// 	}
// }

// func handleIngest(w http.ResponseWriter, r *http.Request) {
// 	var payload struct {
// 		Value string `json:"value"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
// 		http.Error(w, "Invalid JSON", http.StatusBadRequest)
// 		return
// 	}

// 	_, err := conn.Exec("INSERT INTO events (value) VALUES (?)", payload.Value)
// 	if err != nil {
// 		http.Error(w, "ClickHouse insert failed: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	fmt.Fprintf(w, "Ingested: %s", payload.Value)
// }

// package main

// import (
// 	"log"
// 	"net/http"
// 	"os"

// 	"github.com/gorilla/mux"
// 	"github.com/joho/godotenv"
// )

// func main() {
// 	// Load environment variables
// 	if err := godotenv.Load(); err != nil {
// 		log.Println("No .env file found")
// 	}
	
// 	// Initialize JWT secret
// 	jwtSecret = []byte(getEnv("JWT_SECRET", ""))
// 	if len(jwtSecret) == 0 {
// 		log.Fatal("JWT_SECRET is required")
// 	}

// 	// Initialize database
// 	if err := initDatabase(); err != nil {
// 		log.Fatal("Database initialization failed:", err)
// 	}
// 	defer closeDatabase()
	
// 	// Setup and start server
// 	r := mux.NewRouter()
	
// 	// Routes
// 	r.HandleFunc("/login", handleLogin).Methods("POST")
// 	r.HandleFunc("/ingest", authMiddleware(handleIngest)).Methods("POST")
	
// 	// Start server
// 	log.Println("Server running on :8080")
// 	log.Fatal(http.ListenAndServe(":8080", r))
// }

// // getEnv retrieves environment variables with fallback
// func getEnv(key, fallback string) string {
// 	if val, ok := os.LookupEnv(key); ok {
// 		return val
// 	}
// 	return fallback
// }
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Set JWT secret from env
	handler.JwtSecret = []byte(getEnv("JWT_SECRET", ""))
	if len(handler.JwtSecret) == 0 {
		log.Fatal("JWT_SECRET is required")
	}

	// Initialize ClickHouse DB
	if err := clickhouse.InitClickHouse(); err != nil {
		log.Fatal("Database initialization failed:", err)
	}
	defer clickhouse.CloseClickHouse()

	// Setup router
	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/register", handler.Register).Methods("POST")
	r.HandleFunc("/login", handler.Login).Methods("POST")

	// Protected routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(handler.AuthMiddleware)
	api.HandleFunc("/ingest", handler.IngestFlatFile).Methods("POST")
	api.HandleFunc("/data", handler.GetData).Methods("GET") // optional if needed

	// Start server
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
