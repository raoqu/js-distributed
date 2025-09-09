package config

import (
	"strconv"
	"strings"
)

// MySQLConfig holds MySQL database connection details
type MySQLConfig struct {
	Name       string `yaml:"name,omitempty"`
	ConnString string `yaml:"connString"`
	Host       string
	Port       int
	User       string
	Password   string
	DB         string
	Timeout    int
}

// ParseMySQLConnString parses a MySQL connection string and returns a MySQLConfig structure
// Example: "root:password@tcp(localhost:3306)/test?parseTime=true&timeout=10s"
func (c *MySQLConfig) ParseMySQLConnString(connString string) MySQLConfig {
	// If connection string is empty, return defaults
	if connString == "" {
		return *c
	}

	c.ConnString = connString

	// Split the connection string into parts
	// Format: username:password@protocol(host:port)/dbname?param=value

	// First, split by @ to separate credentials from the rest
	parts := strings.Split(connString, "@")
	if len(parts) > 1 {
		// Extract username and password
		credentials := strings.Split(parts[0], ":")
		if len(credentials) > 0 {
			c.User = credentials[0]
		}
		if len(credentials) > 1 {
			c.Password = credentials[1]
		}

		// Process the rest (protocol, host, port, db, params)
		rest := parts[1]

		// Extract protocol and address
		protocolParts := strings.Split(rest, "(")
		if len(protocolParts) > 1 {
			// Extract host and port
			addrPart := strings.Split(protocolParts[1], ")")
			if len(addrPart) > 0 {
				hostPort := strings.Split(addrPart[0], ":")
				if len(hostPort) > 0 {
					c.Host = hostPort[0]
				}
				if len(hostPort) > 1 {
					// Convert port string to int
					if port, err := strconv.Atoi(hostPort[1]); err == nil {
						c.Port = port
					}
				}
			}

			// Extract database and parameters
			if len(addrPart) > 1 {
				dbParams := strings.Split(addrPart[1], "?")
				if len(dbParams) > 0 && len(dbParams[0]) > 1 {
					// Remove leading / from database name
					c.DB = strings.TrimPrefix(dbParams[0], "/")
				}

				// Parse parameters
				if len(dbParams) > 1 {
					paramPairs := strings.Split(dbParams[1], "&")
					for _, pair := range paramPairs {
						keyValue := strings.Split(pair, "=")
						if len(keyValue) == 2 && keyValue[0] == "timeout" {
							// Extract timeout value (e.g., "10s")
							timeoutStr := keyValue[1]
							// Remove any unit suffix and convert to seconds
							timeoutStr = strings.TrimSuffix(timeoutStr, "s")
							if timeout, err := strconv.Atoi(timeoutStr); err == nil {
								c.Timeout = timeout
							}
						}
					}
				}
			}
		}
	}

	return *c
}
