package main

import (
	"os"
	"strings"
)

func readVideoLinkFromFile(filepath string) ([]string, error) {
	var res []string
	dat, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	res = append(res, strings.Split(string(dat), "\n")...)
	return res, nil
}
