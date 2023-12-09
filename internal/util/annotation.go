package util

func GetAnnotationStringValue(annotation string, annotations map[string]string, defaultValue string) string {
	if val, ok := annotations[annotation]; ok {
		return val
	} else if defaultValue != "" {
		return defaultValue
	}
	return ""
}

func GetAnnotationBoolValue(annotation string, annotations map[string]string, defaultValue bool) bool {
	if val, ok := annotations[annotation]; ok {
		if val == "true" {
			return true
		}
	} else if defaultValue {
		return true
	}
	return false
}
