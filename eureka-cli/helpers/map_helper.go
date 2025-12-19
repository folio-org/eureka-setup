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

func GetInt(entry map[string]any, key string) int {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return 0
	}

	value, ok := rawValue.(int)
	if !ok {
		return 0
	}

	return value
}

func GetIntPtr(entry map[string]any, key string) *int {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return nil
	}

	value, ok := rawValue.(int)
	if !ok {
		return nil
	}

	return IntPtr(value)
}

func GetBoolPtr(entry map[string]any, key string) *bool {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return nil
	}

	value, ok := rawValue.(bool)
	if !ok {
		return nil
	}

	return BoolPtr(value)
}

func GetStringSlice(entry map[string]any, key string) []string {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return []string{}
	}

	anySlice, ok := rawValue.([]any)
	if !ok {
		return []string{}
	}

	result := make([]string, 0, len(anySlice))
	for _, item := range anySlice {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}

	return result
}

func GetMap(entry map[string]any, key string) map[string]any {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return make(map[string]any)
	}

	value, ok := rawValue.(map[string]any)
	if !ok {
		return make(map[string]any)
	}

	return value
}

// TODO Add tests
func GetMapOrDefault(entry map[string]any, key string, defaultValue map[string]any) map[string]any {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return defaultValue
	}

	value, ok := rawValue.(map[string]any)
	if !ok {
		return defaultValue
	}

	return value
}

// TODO Add tests
func GetAnySlice(entry map[string]any, key string) []any {
	rawValue, exists := entry[key]
	if !exists || rawValue == nil {
		return []any{}
	}

	value, ok := rawValue.([]any)
	if !ok {
		return []any{}
	}

	return value
}
