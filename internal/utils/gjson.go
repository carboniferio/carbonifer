package utils

import "github.com/tidwall/gjson"

func GetOr(resource *gjson.Result, paths []string) gjson.Result {
	for _, path := range paths {
		if resource.Get(path).Exists() {
			return resource.Get(path)
		}
		if resource.Get("values." + path).Exists() {
			return resource.Get("values." + path)
		}
	}
	return gjson.Result{}
}
