package main

//go:generate go run github.com/swaggo/swag/cmd/swag@latest init -d api/

import (
	"github.com/maslotwi/graph-auth/api"
)

func main() {
	port := 8080
	api.RunAPIServer(port)
}
