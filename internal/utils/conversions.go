package utils

import (
	"fmt"
	"strconv"
)

// ParseToInt converts to an int an interface that could be int, float or string
func ParseToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		var err error
		intValue, err := strconv.Atoi(v)
		if err == nil {
			return intValue, nil
		}
		floatValue, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return int(floatValue), nil
		}
		return 0, err

	default:
		return 0, fmt.Errorf("Cannot convert interface to int: %v", value)
	}
}

// ConvertInterfaceListToStringList converts a list of interfaces to a list of strings
func ConvertInterfaceListToStringList(list []interface{}) []string {
	stringList := []string{}
	for _, item := range list {
		stringList = append(stringList, item.(string))
	}
	return stringList
}
