package main

// ClickHouseConfig holds the connection parameters for ClickHouse
type ClickHouseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	JWTToken string `json:"jwt_token"`
}

// FlatFileConfig holds the configuration for flat file operations
type FlatFileConfig struct {
	Filename  string `json:"filename"`
	Delimiter string `json:"delimiter"`
}

// TableSchema represents database table schema with columns
type TableSchema struct {
	Name    string     `json:"name"`
	Columns []ColumnInfo `json:"columns"`
}

// ColumnInfo represents information about a database column
type ColumnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Selected bool   `json:"selected"`
}

// IngestRequest represents a data ingestion request
type IngestRequest struct {
	SourceType        string           `json:"source_type"` // "clickhouse" or "flatfile"
	ClickHouseConfig  ClickHouseConfig `json:"clickhouse_config,omitempty"`
	FlatFileConfig    FlatFileConfig   `json:"flat_file_config,omitempty"`
	SourceTable       string           `json:"source_table,omitempty"`
	TargetTable       string           `json:"target_table,omitempty"`
	CreateTargetTable bool             `json:"create_target_table"`
	SelectedColumns   []string         `json:"selected_columns"`
	JoinConditions    []JoinCondition  `json:"join_conditions,omitempty"`
}

// JoinCondition represents a join between tables
type JoinCondition struct {
	Table      string `json:"table"`
	Type       string `json:"type"` // INNER, LEFT, RIGHT, FULL
	Condition  string `json:"condition"`
}

// IngestResult represents the result of a data ingestion operation
type IngestResult struct {
	Success      bool   `json:"success"`
	RecordCount  int    `json:"record_count,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// PreviewResult represents data preview results
type PreviewResult struct {
	Headers []string        `json:"headers"`
	Rows    [][]interface{} `json:"rows"`
	Error   string          `json:"error,omitempty"`
}