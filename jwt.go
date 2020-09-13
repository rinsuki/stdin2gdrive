package main

import (
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func getClientFromJWTConfig(path string) *http.Client {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	jwtConf, err := google.JWTConfigFromJSON(bytes, "https://www.googleapis.com/auth/drive")
	return jwtConf.Client(oauth2.NoContext)
}
