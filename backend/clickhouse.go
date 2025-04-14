package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

// getClickHouseConnection returns a new connection to ClickHouse
func getClickHouseConnection(config ClickHouseConfig) (*sql.DB, error) {
	protocol := "tcp"
	if strings.HasPrefix(strings.ToLower(config.Port), "843") || 
	   strings.HasPrefix(strings.ToLower(config.Port), "944") {
		protocol = "https"
	}
	
	dsn := fmt.Sprintf("%s://%s:%s?database=%s&username=%s&password=%s", 
		protocol, config.Host, config.Port, config.Database, config.Username, config.Password)
	
	// Add JWT token if provided
	if config.JWTToken != "" {
		dsn += fmt.Sprintf("&access_token=%s", config.JWTToken)
	}
	
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open ClickHouse connection: %w", err)
	}
	
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}
	
	return db, nil
}

// getClickHouseTables returns a list of tables from ClickHouse
func getClickHouseTables(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()
	
	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}
	
	return tables, nil
}

// getTableSchema returns the schema for a ClickHouse table
func getTableSchema(db *sql.DB, tableName string) (*TableSchema, error) {
	query := fmt.Sprintf("DESCRIBE TABLE %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table schema: %w", err)
	}
	defer rows.Close()
	
	schema := &TableSchema{
		Name: tableName,
	}
	
	for rows.Next() {
		var name, colType, defaultType, defaultExpression string
		var comment sql.NullString
		
		if err := rows.Scan(&name, &colType, &defaultType, &defaultExpression, &comment); err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}
		
		schema.Columns = append(schema.Columns, ColumnInfo{
			Name: name,
			Type: colType,
			Selected: true, // Default to selected
		})
	}
	
	return schema, nil
}

// exportToFile exports data from ClickHouse to a flat file
func exportToFile(db *sql.DB, request IngestRequest) (*IngestResult, error) {
	// Validate request
	if len(request.SelectedColumns) == 0 {
		return nil, fmt.Errorf("no columns selected for export")
	}
	
	// Build query based on selected columns
	columnsStr := strings.Join(request.SelectedColumns, ", ")
	
	// Start with basic query
	query := fmt.Sprintf("SELECT %s FROM %s", columnsStr, request.SourceTable)
	
	// Add joins if specified
	if len(request.JoinConditions) > 0 {
		for _, join := range request.JoinConditions {
			query += fmt.Sprintf(" %s JOIN %s ON %s", join.Type, join.Table, join.Condition)
		}
	}
	
	log.Printf("Executing query: %s", query)
	
	// Execute query
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	
	// Write results to file
	recordCount, err := writeRowsToFile(rows, request.FlatFileConfig, request.SelectedColumns)
	if err != nil {
		return nil, fmt.Errorf("failed to write data to file: %w", err)
	}
	
	return &IngestResult{
		Success:     true,
		RecordCount: recordCount,
	}, nil
}

// previewClickHouseData gets a preview of data from ClickHouse
func previewClickHouseData(db *sql.DB, tableName string, columns []string, limit int) (*PreviewResult, error) {
	columnsStr := "*"
	if len(columns) > 0 {
		columnsStr = strings.Join(columns, ", ")
	}
	
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT %d", columnsStr, tableName, limit)
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to preview data: %w", err)
	}
	defer rows.Close()
	
	// Get column names
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("failed to get column types: %w", err)
	}
	
	headers := make([]string, len(columnTypes))
	for i, col := range columnTypes {
		headers[i] = col.Name()
	}
	
	// Read rows
	var resultRows [][]interface{}
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(headers))
		valuePtrs := make([]interface{}, len(headers))
		
		// Create pointers to each interface{}
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		resultRows = append(resultRows, values)
	}
	
	return &PreviewResult{
		Headers: headers,
		Rows:    resultRows,
	}, nil
}