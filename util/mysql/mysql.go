package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"main/config"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLClient encapsulates MySQL database operations
type MySQLClient struct {
	db      *sql.DB
	timeout time.Duration
}

// Initialize creates MySQL clients with the given configurations
func Initialize() error {
	var err error
	once.Do(func() {
		// Initialize the clients map
		MYSQL_CLIENTS = make(map[string]*MySQLClient)
		cfg := config.CONFIG

		// Initialize clients from MySQLList
		if len(cfg.Database.MySQLList) == 0 {
			log.Println("No MySQL configurations found in MySQLList")
			return
		}

		// Find default client or use first one as default
		var defaultConfig *config.MySQLConfig
		for i, mysqlConfig := range cfg.Database.MySQLList {
			if defaultConfig == nil {
				defaultConfig = &cfg.Database.MySQLList[i]
			} else if mysqlConfig.Name == "default" {
				defaultConfig = &cfg.Database.MySQLList[i]
				break
			}
		}

		// If no default config found, use the first one
		if defaultConfig == nil && len(cfg.Database.MySQLList) > 0 {
			defaultConfig = &cfg.Database.MySQLList[0]
		}

		// Initialize all MySQL clients from the configuration
		for i, mysqlConfig := range cfg.Database.MySQLList {
			name := mysqlConfig.Name
			if name == "" {
				name = fmt.Sprintf("mysql_%d", i)
				cfg.Database.MySQLList[i].Name = name
			}

			if mysqlConfig.ConnString == "" {
				log.Printf("MySQL client '%s' has no connection string, skipping", name)
				continue
			}

			db, dbErr := sql.Open("mysql", mysqlConfig.ConnString)
			if dbErr != nil {
				log.Printf("Failed to connect to MySQL '%s': %v", name, dbErr)
				continue
			}

			// Set connection pool settings
			db.SetMaxOpenConns(25)
			db.SetMaxIdleConns(5)
			db.SetConnMaxLifetime(1 * time.Hour)

			// Test the connection
			timeout := 10 // Default timeout
			if mysqlConfig.Timeout > 0 {
				timeout = mysqlConfig.Timeout
			}

			pingCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()

			if pingErr := db.PingContext(pingCtx); pingErr != nil {
				log.Printf("Failed to ping MySQL '%s': %v", name, pingErr)
				continue
			}

			client := &MySQLClient{
				db:      db,
				timeout: time.Duration(timeout) * time.Second,
			}

			// Add to the clients map
			MYSQL_CLIENTS[name] = client

			// Set as default client if this is the default config
			if defaultConfig != nil && name == defaultConfig.Name {
				MYSQL_CLIENT = client
				log.Printf("Default MySQL client set to '%s': %s@%s", name, mysqlConfig.DB, mysqlConfig.Host)
			} else {
				log.Printf("MySQL client '%s' initialized successfully: %s@%s", name, mysqlConfig.DB, mysqlConfig.Host)
			}
		}

		// If MYSQL_CLIENT is still nil but we have clients, set the first one as default
		if MYSQL_CLIENT == nil && len(MYSQL_CLIENTS) > 0 {
			for name, client := range MYSQL_CLIENTS {
				MYSQL_CLIENT = client
				log.Printf("Setting '%s' as default MySQL client since no 'default' client was found", name)
				break
			}
		}
	})

	return err
}

// Close closes the MySQL database connection
func (c *MySQLClient) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Query executes a query that returns rows
func (c *MySQLClient) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if c.db == nil {
		return nil, fmt.Errorf("MySQL client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	return c.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (c *MySQLClient) QueryRow(query string, args ...interface{}) *sql.Row {
	if c.db == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	return c.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows
func (c *MySQLClient) Exec(query string, args ...interface{}) (sql.Result, error) {
	if c.db == nil {
		return nil, fmt.Errorf("MySQL client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	return c.db.ExecContext(ctx, query, args...)
}

// QueryToMap executes a query and returns the results as a slice of maps
func (c *MySQLClient) QueryToMap(query string, args ...interface{}) ([]map[string]interface{}, error) {
	if c.db == nil {
		return nil, fmt.Errorf("MySQL client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Create a slice of interface{} to hold each row's values
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Create the result slice
	var result []map[string]interface{}

	// Iterate through the rows
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		// Create a map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Handle nil values
			if val == nil {
				row[col] = nil
				continue
			}

			// Handle different types
			switch v := val.(type) {
			case []byte:
				// Convert []byte to string for text data
				row[col] = string(v)
			default:
				row[col] = v
			}
		}

		result = append(result, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// Begin starts a new transaction
func (c *MySQLClient) Begin() (*sql.Tx, error) {
	if c.db == nil {
		return nil, fmt.Errorf("MySQL client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	return c.db.BeginTx(ctx, nil)
}
