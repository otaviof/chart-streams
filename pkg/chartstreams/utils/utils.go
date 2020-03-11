package utils

// ContainsStringSlice given a string slice and a string, returns boolean when is contained.
func ContainsStringSlice(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
