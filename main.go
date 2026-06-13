package main

//go:generate go run github.com/swaggo/swag/cmd/swag@latest init -d api/

import (
	"github.com/maslotwi/graph-auth/api"
	"github.com/maslotwi/graph-auth/helpers/environment"
)

func main() {
	environment.LoadEnv()
	api.RunAPIServer()
}
