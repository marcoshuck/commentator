package main

import "net/http"

type Service interface {
	Routes() http.Handler
}
