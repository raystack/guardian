package utils

func ContainsOrdered(ss []string, lookingFor []string) (bool, int) {
	var headFound bool
	var headFoundIndex int

	if len(lookingFor) <= 0 {
		return true, headFoundIndex
	}
	head := lookingFor[0]

	for i := 0; i < len(ss); i++ {
		if !headFound && ss[i] != head {
			// skip head mismatch
			continue
		} else if !headFound {
			// mark position in ss when found matching head
			headFound = true
			headFoundIndex = i
			continue
		} else if ss[i] != lookingFor[i-headFoundIndex] {
			// return false on first mismatch after head
			return false, headFoundIndex
		} else if len(lookingFor)-1 == i-headFoundIndex {
			return true, headFoundIndex
		}
	}
	return false, headFoundIndex
}
