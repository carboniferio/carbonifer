package utils

import (
	"fmt"
	"strconv"
)

// Convert to an int an interface that could be int, float or string
func ParseToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		if intValue, err := strconv.Atoi(v); err == nil {
			return intValue, nil
		} else {
			return 0, err
		}
	default:
		return 0, fmt.Errorf("Cannot convert interface to int: %v", value)
	}
}

func ConvertInterfaceListToStringList(list []interface{}) []string {
	stringList := []string{}
	for _, item := range list {
		stringList = append(stringList, item.(string))
	}
	return stringList
}
