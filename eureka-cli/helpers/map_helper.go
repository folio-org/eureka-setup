package helpers

func GetBool(mapEntry map[string]any, key string) bool {
	value := mapEntry[key]
	return value != nil && value.(bool)
}

func GetAnyOrDefault(mapEntry map[string]any, key string, defaultValue any) any {
	if mapEntry[key] == nil {
		return defaultValue
	}

	return mapEntry[key]
}

func GetIntOrDefault(mapEntry map[string]any, key string, defaultValue int64) int64 {
	value, ok := mapEntry[key].(int)
	if !ok || mapEntry[key] == nil {
		return int64(defaultValue)
	}

	return int64(value)
}

func GetBoolOrDefault(mapEntry map[string]any, key string, defaultValue bool) bool {
	value, ok := mapEntry[key].(bool)
	if !ok || mapEntry[key] == nil {
		return defaultValue
	}

	return value
}
