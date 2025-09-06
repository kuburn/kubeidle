package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// DBConnector is an interface for database operations
// that can be implemented for different databases.
type DBConnector interface {
	Connect() error
	Close() error
	InsertScaleDownState(state ScaleDownState) error
}

// MySQLConnector implements the DBConnector interface for MySQL.
type MySQLConnector struct {
	DSN string
	DB  *sql.DB
}

// ScaleDownState represents the state of a scaled-down object.
type ScaleDownState struct {
	ClusterName     string
	ApplicationName string
	Namespace       string
	ObjectType      string
	Replicas        int
	CPURequests     string
	MemoryRequests  string
	CPULimits       string
	MemoryLimits    string
	ScaleDownTime   string
	ScaleUpTime     string
}

// Connect establishes a connection to the MySQL database.
func (m *MySQLConnector) Connect() error {
	var err error
	m.DB, err = sql.Open("mysql", m.DSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	return nil
}

// Close closes the database connection.
func (m *MySQLConnector) Close() error {
	if m.DB != nil {
		return m.DB.Close()
	}
	return nil
}

// InsertScaleDownState inserts the scale-down state into the database.
func (m *MySQLConnector) InsertScaleDownState(state ScaleDownState) error {
	query := `INSERT INTO scale_down_states (cluster_name, application_name, namespace, object_type, replicas, cpu_requests, memory_requests, cpu_limits, memory_limits, scale_down_time, scale_up_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := m.DB.Exec(query, state.ClusterName, state.ApplicationName, state.Namespace, state.ObjectType, state.Replicas, state.CPURequests, state.MemoryRequests, state.CPULimits, state.MemoryLimits, state.ScaleDownTime, state.ScaleUpTime)
	if err != nil {
		return fmt.Errorf("failed to insert scale-down state: %w", err)
	}
	return nil
}
