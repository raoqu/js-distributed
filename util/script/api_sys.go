package script

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
)

// CommandOptions holds all the options for a command execution
type CommandOptions struct {
	// Command is the command to execute
	Command string
	// Args contains command arguments
	Args []string
	// WorkDir specifies the working directory for command execution
	WorkDir string
}

// Sys_command implements a function to execute shell commands from scripts
// This function provides a JavaScript-friendly API for executing shell commands
// and capturing their output.
//
// Usage in JS:
//
//	sys.command(command, {
//	  args: ["arg1", "arg2"],   // Command arguments (optional)
//	  workDir: "/path/to/dir"   // Working directory (optional)
//	})
//
// Returns an object with the following properties:
//
//	{
//	  success: boolean,         // Whether the command executed successfully
//	  output: string,           // Combined stdout and stderr output
//	  error: string,            // Error message if command failed, null otherwise
//	  exitCode: number          // Command exit code
//	}
func Sys_command(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	if len(call.Arguments) < 1 {
		return nil, fmt.Errorf("sys.command requires at least a command argument")
	}

	// Get command
	cmdStr := call.Arguments[0].String()
	if cmdStr == "" {
		return nil, fmt.Errorf("command cannot be empty")
	}

	// Parse options
	opts := parseCommandOptions(rt, call)

	// Execute the command
	result := executeCommand(cmdStr, opts)

	return rt.ToValue(result), nil
}

// parseCommandOptions extracts and processes options from JavaScript arguments
func parseCommandOptions(rt *goja.Runtime, call goja.FunctionCall) CommandOptions {
	// Default values
	opts := CommandOptions{
		Args:    []string{},
		WorkDir: "",
	}

	// Process options if provided
	if len(call.Arguments) > 1 && !goja.IsUndefined(call.Arguments[1]) && !goja.IsNull(call.Arguments[1]) {
		options := call.Arguments[1].ToObject(rt)
		if options != nil {
			// Extract args
			argsVal := options.Get("args")
			if argsVal != nil && !goja.IsUndefined(argsVal) && !goja.IsNull(argsVal) {
				argsObj := argsVal.Export()
				if argsArr, ok := argsObj.([]interface{}); ok {
					for _, arg := range argsArr {
						opts.Args = append(opts.Args, fmt.Sprintf("%v", arg))
					}
				}
			}

			// Extract workDir
			workDirVal := options.Get("workDir")
			if workDirVal != nil && !goja.IsUndefined(workDirVal) && !goja.IsNull(workDirVal) {
				opts.WorkDir = workDirVal.String()
			}
		}
	}

	return opts
}

// executeCommand performs the actual command execution
func executeCommand(cmdStr string, opts CommandOptions) map[string]interface{} {
	// Create result object
	result := make(map[string]interface{})
	result["success"] = false
	result["output"] = ""
	result["error"] = nil
	result["exitCode"] = -1

	// Create command
	cmd := exec.Command(cmdStr, opts.Args...)

	// Set working directory if provided
	if opts.WorkDir != "" {
		// Ensure the working directory exists and is absolute
		absWorkDir, err := filepath.Abs(opts.WorkDir)
		if err != nil {
			result["error"] = fmt.Sprintf("Failed to resolve working directory: %v", err)
			return result
		}

		// Check if directory exists
		if _, err := os.Stat(absWorkDir); os.IsNotExist(err) {
			result["error"] = fmt.Sprintf("Working directory does not exist: %s", absWorkDir)
			return result
		}

		cmd.Dir = absWorkDir
	}

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	
	// Set output
	result["output"] = strings.TrimSpace(string(output))
	
	// Handle error and exit code
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result["exitCode"] = exitError.ExitCode()
			result["error"] = fmt.Sprintf("Command failed with exit code %d", exitError.ExitCode())
		} else {
			result["error"] = fmt.Sprintf("Failed to execute command: %v", err)
		}
		return result
	}

	// Command succeeded
	result["success"] = true
	result["exitCode"] = 0
	
	return result
}
