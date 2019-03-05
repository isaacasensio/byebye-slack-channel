package main

import (
	"errors"
	"io/ioutil"
	"strings"
)

func readSlackToken(filename string) (string, error) {

	buff, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	token := string(buff)
	if strings.TrimSpace(token) == "" {
		return "", errors.New("invalid token")
	}

	return token, nil
}