package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadSlackToken_fails_when_secret_file_is_not_found(t *testing.T) {
	_, err := readSlackToken("unknown.file")
	assert.Error(t, err)
}

func TestReadSlackToken_fails_when_secret_file_is_empty(t *testing.T) {
	filename := "empty.file"
	defer removeFile(filename)

	err := ioutil.WriteFile(filename, nil, 0644)
	assert.NoError(t, err)

	_, err = readSlackToken(filename)
	assert.Equal(t, "invalid token", err.Error())
}

func TestReadSlackToken_returns_token_stored_in_file(t *testing.T) {
	filename := "token.file"
	defer removeFile(filename)

	b := []byte("secret!!")
	err := ioutil.WriteFile(filename, b, 0644)

	token, err := readSlackToken(filename)
	assert.NoError(t, err)
	assert.Equal(t, "secret!!", token)
}

func removeFile(filename string) {
	os.Remove(filename)
}
