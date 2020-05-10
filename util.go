package apimdb

import (
	"fmt"
	"strings"
)

const imdbSearchURL = "https://www.imdb.com/find?s="

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func splitIMDBName(s string) (string, error) {
	nameID := strings.Split(s, "/")
	if strings.HasPrefix(nameID[2], "nm") || strings.HasPrefix(nameID[2], "tt") {
		return nameID[2], nil
	}
	return "", fmt.Errorf("Could not find nameID")
}
