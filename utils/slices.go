package utils

func SubsliceExists(slice, subslice []string) (bool, int) {
	// check for empty subslice slice
	if len(subslice) == 0 {
		return true, 0
	}

	// check if subslice slice is longer than slice
	if len(subslice) > len(slice) {
		return false, 0
	}

	// looking for a subslice
	for i := 0; i < len(slice); i++ {
		if slice[i] == subslice[0] {
			found := true
			for j := 1; j < len(subslice); j++ {
				if slice[i+j] != subslice[j] {
					found = false
					break
				}
			}
			if found {
				return true, i
			}
		}
	}

	return false, 0
}
