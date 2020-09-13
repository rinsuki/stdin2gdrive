package main

import (
	"log"
	"net/http"
)

func httpPanic(res *http.Response) {
	log.Panicln(res.Status, readAllString(res.Body))
}
