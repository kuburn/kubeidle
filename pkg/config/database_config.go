package config

import (
	"fmt"
)

// DatabaseConfig holds the configuration for database connections
type DatabaseConfig struct {
	Type     string // e.g., "mysql", "postgresql"
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

// NewMySQLConfig creates a new MySQL database configuration
func NewMySQLConfig(host string, port int, username, password, database string) *DatabaseConfig {
	return &DatabaseConfig{
		Type:     "mysql",
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
	}
}

// GetDSN returns the Data Source Name for the database connection
func (c *DatabaseConfig) GetDSN() string {
	switch c.Type {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	default:
		return ""
	}
}
