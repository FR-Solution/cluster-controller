package utils

func MergeValues(values, extraValues map[string]interface{}) {
	for key := range extraValues {
		_, isExist := values[key]
		_, ok := extraValues[key].(map[string]interface{})
		if !ok || !isExist {
			values[key] = extraValues[key]
			continue
		}
		MergeValues(values[key].(map[string]interface{}), extraValues[key].(map[string]interface{}))
	}
}
