package apihelpers

// StringMapToAnyMap converts map[string]string to map[string]any for API request bodies.
func StringMapToAnyMap(m map[string]string) map[string]any {
	if m == nil {
		return nil
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
