package helpers

func GetBool(entry map[string]any, key string) bool {
	if entry[key] == nil {
		return false
	}
	newValue, ok := entry[key].(bool)
	if !ok {
		return false
	}

	return newValue
}

func GetAnyOrDefault(entry map[string]any, key string, defaultValue any) any {
	if entry[key] == nil {
		return defaultValue
	}

	return entry[key]
}

func GetIntOrDefault(entry map[string]any, key string, defaultValue int64) int64 {
	if entry[key] == nil {
		return defaultValue
	}
	newValue, ok := entry[key].(int)
	if !ok {
		return defaultValue
	}

	return int64(newValue)
}

func GetBoolOrDefault(entry map[string]any, key string, defaultValue bool) bool {
	if entry[key] == nil {
		return defaultValue
	}
	newValue, ok := entry[key].(bool)
	if !ok {
		return defaultValue
	}

	return newValue
}

func SetBool(entry map[string]any, key string, value *bool) {
	if entry[key] != nil {
		newValue, ok := entry[key].(bool)
		if !ok {
			return
		}
		*value = newValue
	}
}

func SetString(entry map[string]any, key string, value *string) {
	if entry[key] != nil {
		newValue, ok := entry[key].(string)
		if !ok {
			return
		}
		*value = newValue
	}
}
