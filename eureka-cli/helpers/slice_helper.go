package helpers

func ConvertMapKeysToSlice(inputMap map[string]any) []string {
	keys := make([]string, 0, len(inputMap))
	for key := range inputMap {
		keys = append(keys, key)
	}
	return keys
}
