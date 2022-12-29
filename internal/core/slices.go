package core

func GetValueAt(slice []string, index int) string {
	if len(slice) >= index+1 {
		return slice[index]
	}

	return ""
}

func MergeValues(slice []string) string {
	if slice == nil || len(slice) == 0 {
		return ""
	}

	result := ""
	for _, item := range slice {
		result += item
	}

	return result
}
