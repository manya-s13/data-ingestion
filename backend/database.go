package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

// Global database connection
var conn *sql.DB

// initDatabase initializes the database connection
func initDatabase() error {
	url := getEnv("CLICKHOUSE_URL", "tcp://localhost:9000")
	db := getEnv("CLICKHOUSE_DB", "default")
	user := getEnv("CLICKHOUSE_USER", "default")
	pass := getEnv("CLICKHOUSE_PASS", "")
	
	dsn := fmt.Sprintf("%s?database=%s&username=%s&password=%s", url, db, user, pass)
	
	var err error
	conn, err = sql.Open("clickhouse", dsn)
	if err != nil {
		return fmt.Errorf("ClickHouse connection failed: %w", err)
	}
	
	if err := conn.Ping(); err != nil {
		return fmt.Errorf("ClickHouse ping failed: %w", err)
	}
	
	log.Println("Connected to ClickHouse")
	return nil
}

// closeDatabase closes the database connection
func closeDatabase() {
	if conn != nil {
		conn.Close()
	}
}

// insertEvent inserts a new event into the database
func insertEvent(value string) error {
	_, err := conn.Exec("INSERT INTO events (value) VALUES (?)", value)
	return err
}