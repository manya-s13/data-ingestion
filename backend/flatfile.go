package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// getFlatFileSchema attempts to determine the schema of a flat file
func getFlatFileSchema(config FlatFileConfig) (*TableSchema, error) {
	file, err := os.Open(config.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	if config.Delimiter != "" {
		reader.Comma = rune(config.Delimiter[0])
	}
	
	// Read header row
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header row: %w", err)
	}
	
	// Read a sample row to guess column types
	sampleRow, err := reader.Read()
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read sample row: %w", err)
	}
	
	schema := &TableSchema{
		Name: getFileBaseName(config.Filename),
	}
	
	// Create columns
	for i, header := range headers {
		colType := "String" // Default type
		
		// Simple type inference if we have a sample row
		if err != io.EOF && i < len(sampleRow) {
			colType = inferColumnType(sampleRow[i])
		}
		
		schema.Columns = append(schema.Columns, ColumnInfo{
			Name:     header,
			Type:     colType,
			Selected: true, // Default to selected
		})
	}
	
	return schema, nil
}

// writeRowsToFile writes SQL query results to a CSV file
func writeRowsToFile(rows *sql.Rows, config FlatFileConfig, columns []string) (int, error) {
	file, err := os.Create(config.Filename)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	if config.Delimiter != "" {
		writer.Comma = rune(config.Delimiter[0])
	}
	defer writer.Flush()
	
	// Write header row
	if err := writer.Write(columns); err != nil {
		return 0, fmt.Errorf("failed to write header: %w", err)
	}
	
	// Get column types for proper conversion
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return 0, fmt.Errorf("failed to get column types: %w", err)
	}
	
	// Create a slice of interface{} to hold the values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	
	// Create pointers to each interface{}
	for i := range columns {
		valuePtrs[i] = &values[i]
	}
	
	recordCount := 0
	for rows.Next() {
		// Scan the values from the row into the interface slice
		if err := rows.Scan(valuePtrs...); err != nil {
			return recordCount, fmt.Errorf("failed to scan row: %w", err)
		}
		
		// Convert values to strings
		rowStrings := make([]string, len(columns))
		for i, val := range values {
			rowStrings[i] = fmt.Sprintf("%v", val)
		}
		
		// Write the row to the CSV file
		if err := writer.Write(rowStrings); err != nil {
			return recordCount, fmt.Errorf("failed to write row: %w", err)
		}
		
		recordCount++
	}
	
	return recordCount, nil
}

// importFromFile imports data from a CSV file to ClickHouse
func importFromFile(db *sql.DB, request IngestRequest) (*IngestResult, error) {
	// Read file schema
	schema, err := getFlatFileSchema(request.FlatFileConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get file schema: %w", err)
	}
	
	// Filter columns based on selection
	var selectedColumns []string
	var selectedTypes []string
	
	if len(request.SelectedColumns) > 0 {
		for _, col := range schema.Columns {
			for _, selected := range request.SelectedColumns {
				if col.Name == selected {
					selectedColumns = append(selectedColumns, col.Name)
					selectedTypes = append(selectedTypes, col.Type)
					break
				}
			}
		}
	} else {
		// If no columns explicitly selected, use all
		for _, col := range schema.Columns {
			selectedColumns = append(selectedColumns, col.Name)
			selectedTypes = append(selectedTypes, col.Type)
		}
	}
	
	// Create table if requested
	if request.CreateTargetTable {
		if err := createClickHouseTable(db, request.TargetTable, selectedColumns, selectedTypes); err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}
	
	// Read file data
	file, err := os.Open(request.FlatFileConfig.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	if request.FlatFileConfig.Delimiter != "" {
		reader.Comma = rune(request.FlatFileConfig.Delimiter[0])
	}
	
	// Skip header row
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}
	
	// Prepare batch insert statement
	columnsStr := strings.Join(selectedColumns, ", ")
	placeholders := strings.Repeat("?, ", len(selectedColumns))
	placeholders = placeholders[:len(placeholders)-2] // Remove trailing ", "
	
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", 
		request.TargetTable, columnsStr, placeholders)
	
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	
	// Process rows
	recordCount := 0
	batch := make([][]interface{}, 0, 1000) // Process in batches
	
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to read row: %w", err)
		}
		
		// Extract selected columns
		values := make([]interface{}, len(selectedColumns))
		for i, colName := range selectedColumns {
			colIndex := -1
			for j, header := range schema.Columns {
				if header.Name == colName && j < len(record) {
					colIndex = j
					break
				}
			}
			
			if colIndex >= 0 && colIndex < len(record) {
				values[i] = record[colIndex]
			} else {
				values[i] = nil
			}
		}
		
		batch = append(batch, values)
		
		// Execute batch if size limit reached
		if len(batch) >= 1000 {
			for _, row := range batch {
				if _, err := stmt.Exec(row...); err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("failed to insert row: %w", err)
				}
				recordCount++
			}
			batch = batch[:0] // Clear batch
		}
	}
	
	// Insert any remaining rows
	for _, row := range batch {
		if _, err := stmt.Exec(row...); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to insert row: %w", err)
		}
		recordCount++
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return &IngestResult{
		Success:     true,
		RecordCount: recordCount,
	}, nil
}

// createClickHouseTable creates a new table in ClickHouse
func createClickHouseTable(db *sql.DB, tableName string, columns []string, types []string) error {
	if len(columns) != len(types) {
		return fmt.Errorf("column and type counts don't match")
	}
	
	// Build column definitions
	var columnDefs []string
	for i := range columns {
		columnDefs = append(columnDefs, fmt.Sprintf("%s %s", columns[i], types[i]))
	}
	
	// Default engine
	engineClause := "ENGINE = MergeTree() ORDER BY tuple()"
	
	// Create table
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s) %s", 
		tableName, strings.Join(columnDefs, ", "), engineClause)
	
	_, err := db.Exec(query)
	return err
}

// previewFlatFileData gets a preview of data from a flat file
func previewFlatFileData(config FlatFileConfig, limit int) (*PreviewResult, error) {
	file, err := os.Open(config.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	if config.Delimiter != "" {
		reader.Comma = rune(config.Delimiter[0])
	}
	
	// Read header row
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header row: %w", err)
	}
	
	// Read preview rows
	var rows [][]interface{}
	for i := 0; i < limit; i++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}
		
		// Convert to interface{}
		row := make([]interface{}, len(record))
		for j, val := range record {
			row[j] = val
		}
		
		rows = append(rows, row)
	}
	
	return &PreviewResult{
		Headers: headers,
		Rows:    rows,
	}, nil
}

// getFileBaseName gets the base name of a file without extension
func getFileBaseName(filename string) string {
	base := filename
	
	// Remove path
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	if i := strings.LastIndex(base, "\\"); i >= 0 {
		base = base[i+1:]
	}
	
	// Remove extension
	if i := strings.LastIndex(base, "."); i >= 0 {
		base = base[:i]
	}
	
	return base
}

// inferColumnType attempts to guess the data type from a sample value
func inferColumnType(sample string) string {
	// Simple inference rules
	if sample == "" {
		return "String"
	}
	
	// Try to interpret as integer
	if strings.ContainsAny(sample, "0123456789") && !strings.ContainsAny(sample, ".,e") {
		return "Int64"
	}
	
	// Try to interpret as float
	if strings.ContainsAny(sample, "0123456789") && strings.ContainsAny(sample, ".e") {
		return "Float64"
	}
	
	// Try to interpret as date/datetime
	if strings.Contains(sample, "-") && strings.Contains(sample, ":") {
		return "DateTime"
	}
	if strings.Contains(sample, "-") && !strings.Contains(sample, ":") {
		return "Date"
	}
	
	// Default to string
	return "String"
}