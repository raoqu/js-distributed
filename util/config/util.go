package config

import "strings"

// propertyString is a string in format "key1=val1,key2=val2,..."
func ConstructPropertyMap(propertyString string) map[string]string {
	if propertyString == "" {
		return make(map[string]string)
	}

	// Parse property string in format "key1=val1,key2=val2,..."
	propertyMap := make(map[string]string)
	for _, pair := range strings.Split(propertyString, ",") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			propertyMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return propertyMap
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
