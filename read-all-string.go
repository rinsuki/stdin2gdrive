package main

import (
	"io"
	"strings"
)

func readAllString(reader io.Reader) string {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, reader)
	// check errors
	if err != nil {
		panic(err)
	}
	return buf.String()
}
