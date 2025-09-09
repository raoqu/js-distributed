package script

import (
	"fmt"
	"log"

	"github.com/dop251/goja"
)

func Console_log(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	args := make([]interface{}, len(call.Arguments))
	for i, arg := range call.Arguments {
		args[i] = arg.Export()
	}

	// Format the output
	output := fmt.Sprint(args...)
	log.Printf("[SCRIPT] %s", output)

	return goja.Undefined(), nil
}

func Console_error(rt *goja.Runtime, call goja.FunctionCall) (goja.Value, error) {
	args := make([]interface{}, len(call.Arguments))
	for i, arg := range call.Arguments {
		args[i] = arg.Export()
	}

	// Format the output
	output := fmt.Sprint(args...)

	log.Printf("[SCRIPT ERROR] %s", output)

	return goja.Undefined(), nil
}
