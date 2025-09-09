package strings

import (
	"fmt"
	"math/big"
	"strings"
)

func ToString(val interface{}) string {
	switch value := val.(type) {
	case string:
		return value
	case int:
		return fmt.Sprintf("%d", value)
	case int64:
		return fmt.Sprintf("%d", value)
	case float64:
		return fmt.Sprintf("%f", value)
	case big.Int:
		return value.String()
	case *big.Int:
		return value.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

func ParsePropertyMap(propertyString string) map[string]string {
	if propertyString == "" {
		return make(map[string]string)
	}

	// Parse property string in format "key1=val1,key2=val2,..."
	propertyMap := make(map[string]string)
	for _, pair := range strings.Split(propertyString, ",") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			propertyMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		} else if len(parts) == 1 {
			propertyMap[strings.TrimSpace(parts[0])] = "true"
		}
	}
	return propertyMap
}

func Extract(str string, start, end string) (string, string) {
	startIdx := strings.Index(str, start)
	if startIdx == -1 {
		return "", str
	}
	endIdx := strings.Index(str, end)
	if endIdx == -1 {
		return "", str
	}
	return str[startIdx+len(start) : endIdx], str[endIdx+len(end):]
}

// propertyString is a string in format "key1=val1,key2=val2,..."
func GetPropertyValue(propertyString string, key string, defaultValue string) string {
	if propertyString == "" {
		return defaultValue
	}

	// Parse property string in format "key1=val1,key2=val2,..."
	propertyMap := make(map[string]string)
	for _, pair := range strings.Split(propertyString, ",") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			propertyMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		} else if len(parts) == 1 {
			propertyMap[strings.TrimSpace(parts[0])] = "true"
		}
	}

	// Return the value for the specified key or the default value if not found
	if val, exists := propertyMap[key]; exists {
		return val
	}
	return defaultValue
}
