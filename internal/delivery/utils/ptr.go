package utils

func GetDefaultIfNull(types []string) []string {
	if types == nil {
		return []string{}
	}

	return types
}
