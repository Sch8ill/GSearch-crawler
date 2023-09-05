package parseUtils

import (
	"strings"
)

// removes duplicate items from a slice
func RemoveDuplicateItems[T string | int](rawSlice []T) []T {
	allKeys := make(map[T]bool)
	slice := []T{}
	for _, item := range rawSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			slice = append(slice, item)
		}
	}
	return slice
}

// removes tags from an url
func RemoveTagsfromUrl(url string) string {
	if strings.Contains(url, "#") {
		url = strings.Split(url, "#")[0]
	}
	return url
}

// checks if a given item is in a given slice
func Contains[T string | int](list []T, wanted T) bool {
	for _, item := range list {
		if item == wanted {
			return true
		}
	}
	return false
}
