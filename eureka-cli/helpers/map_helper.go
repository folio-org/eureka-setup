package helpers

func GetString(entry map[string]any, key string) string {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return ""
	}

	value, ok := rawValue.(string)
	if !ok {
		return ""
	}

	return value
}

func GetBool(entry map[string]any, key string) bool {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return false
	}

	value, ok := rawValue.(bool)
	if !ok {
		return false
	}

	return value
}

func GetAnyOrDefault(entry map[string]any, key string, defaultValue any) any {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return defaultValue
	}

	return rawValue
}

func GetStringOrDefault(entry map[string]any, key string, defaultValue string) string {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return defaultValue
	}

	value, ok := rawValue.(string)
	if !ok {
		return defaultValue
	}

	return value
}

func GetIntOrDefault(entry map[string]any, key string, defaultValue int64) int64 {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return defaultValue
	}

	newValue, ok := rawValue.(int)
	if !ok {
		return defaultValue
	}

	return int64(newValue)
}

func GetBoolOrDefault(entry map[string]any, key string, defaultValue bool) bool {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return defaultValue
	}

	newValue, ok := rawValue.(bool)
	if !ok {
		return defaultValue
	}

	return newValue
}

func SetBool(entry map[string]any, key string, value *bool) {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return
	}

	newValue, ok := rawValue.(bool)
	if !ok {
		return
	}
	*value = newValue
}

func SetString(entry map[string]any, key string, value *string) {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return
	}

	newValue, ok := rawValue.(string)
	if !ok {
		return
	}
	*value = newValue
}
