package script

import (
	"fmt"
	"main/util/mysql"
	"main/util/strings"

	"github.com/dop251/goja"
)

// MySQL_query executes a SQL query and returns the results as an array of objects
// Usage in JS:
//
//	mysql.query("SELECT * FROM users WHERE age > ?", [25])
//
// Returns an array of objects with column names as keys
func MySQL_query(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	if len(call.Arguments) < 1 {
		return nil, fmt.Errorf("mysql.query requires at least a query string")
	}

	// var query string
	// var dbName string
	// argIndex := 1
	param := call.Arguments[0].String()
	db, query := strings.Extract(param, "[", "]")

	// Check if MySQL client is initialized
	if mysql.MYSQL_CLIENT == nil {
		return nil, fmt.Errorf("MySQL client is not initialized")
	}

	// Process arguments if provided
	var args []interface{}
	if len(call.Arguments) > 1 && !goja.IsUndefined(call.Arguments[1]) && !goja.IsNull(call.Arguments[1]) {
		// Extract arguments array
		argsValue := call.Arguments[1].Export()
		if argsArray, ok := argsValue.([]interface{}); ok {
			args = argsArray
		} else {
			return nil, fmt.Errorf("second argument must be an array of query parameters")
		}
	}

	client := mysql.GetClient(db)
	if client == nil {
		return nil, fmt.Errorf("MySQL client is not initialized")
	}

	results, err := client.QueryToMap(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	// Convert the results to a JavaScript array
	return rt.ToValue(results), nil
}

// MySQL_exec executes a SQL statement that doesn't return rows (INSERT, UPDATE, DELETE)
// Usage in JS:
//
//	mysql.exec("INSERT INTO users (name, age) VALUES (?, ?)", ["John", 30])
//
// Returns an object with lastInsertId and rowsAffected properties
func MySQL_exec(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	if len(call.Arguments) < 1 {
		return nil, fmt.Errorf("mysql.exec requires at least a query string")
	}

	// Get the query string
	query := call.Arguments[0].String()

	// Check if MySQL client is initialized
	if mysql.MYSQL_CLIENT == nil {
		return nil, fmt.Errorf("MySQL client is not initialized")
	}

	// Process arguments if provided
	var args []interface{}
	if len(call.Arguments) > 1 && !goja.IsUndefined(call.Arguments[1]) && !goja.IsNull(call.Arguments[1]) {
		// Extract arguments array
		argsValue := call.Arguments[1].Export()
		if argsArray, ok := argsValue.([]interface{}); ok {
			args = argsArray
		} else {
			return nil, fmt.Errorf("second argument must be an array of query parameters")
		}
	}

	// Execute the statement
	result, err := mysql.MYSQL_CLIENT.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("statement execution failed: %w", err)
	}

	// Get affected rows and last insert ID
	lastInsertID, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()

	// Create a result object
	resultObj := map[string]interface{}{
		"lastInsertId": lastInsertID,
		"rowsAffected": rowsAffected,
	}

	return rt.ToValue(resultObj), nil
}

// MySQL_queryRow executes a SQL query and returns a single row as an object
// Usage in JS:
//
//	mysql.queryRow("SELECT * FROM users WHERE id = ?", [1])
//
// Returns an object with column names as keys or null if no rows found
func MySQL_queryRow(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	if len(call.Arguments) < 1 {
		return nil, fmt.Errorf("mysql.queryRow requires at least a query string")
	}

	// Get the query string
	query := call.Arguments[0].String()

	// Check if MySQL client is initialized
	if mysql.MYSQL_CLIENT == nil {
		return nil, fmt.Errorf("MySQL client is not initialized")
	}

	// Process arguments if provided
	var args []interface{}
	if len(call.Arguments) > 1 && !goja.IsUndefined(call.Arguments[1]) && !goja.IsNull(call.Arguments[1]) {
		// Extract arguments array
		argsValue := call.Arguments[1].Export()
		if argsArray, ok := argsValue.([]interface{}); ok {
			args = argsArray
		} else {
			return nil, fmt.Errorf("second argument must be an array of query parameters")
		}
	}

	// Execute the query
	results, err := mysql.MYSQL_CLIENT.QueryToMap(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	// If no results, return null
	if len(results) == 0 {
		return goja.Null(), nil
	}

	// Return the first row
	return rt.ToValue(results[0]), nil
}

// MySQL_transaction executes a function within a transaction
// Usage in JS:
//
//	mysql.transaction(function() {
//	  mysql.exec("INSERT INTO users (name) VALUES (?)", ["John"]);
//	  mysql.exec("UPDATE counters SET value = value + 1 WHERE name = ?", ["user_count"]);
//	});
//
// Returns true if the transaction was committed successfully
func MySQL_transaction(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	if len(call.Arguments) < 1 {
		return nil, fmt.Errorf("mysql.transaction requires a callback function")
	}

	// Get the callback function
	callback, ok := goja.AssertFunction(call.Arguments[0])
	if !ok {
		return nil, fmt.Errorf("first argument must be a function")
	}

	// Check if MySQL client is initialized
	if mysql.MYSQL_CLIENT == nil {
		return nil, fmt.Errorf("MySQL client is not initialized")
	}

	// Start a transaction
	tx, err := mysql.MYSQL_CLIENT.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	// Create a deferred function to handle rollback in case of error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Re-throw the panic after rollback
		}
	}()

	// Execute the callback function
	_, err = callback(goja.Undefined())
	if err != nil {
		tx.Rollback()
		return rt.ToValue(false), err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return rt.ToValue(false), fmt.Errorf("failed to commit transaction: %w", err)
	}

	return rt.ToValue(true), nil
}
