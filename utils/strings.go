package utils

func ContainsString(arr []string, v string) bool {
	for _, item := range arr {
		if item == v {
			return true
		}
	}
	return false
}
