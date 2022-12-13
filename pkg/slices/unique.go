package slices

func UniqueStringSlice(slice []string) []string {
	if slice == nil {
		return nil
	}

	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, exists := keys[entry]; !exists {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
