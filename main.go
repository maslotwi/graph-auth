package main

import (
	"fmt"

	"github.com/maslotwi/graph-auth/api"
)

func main() {
	port := 8080
	fmt.Printf("Listening on port %d\n", port)
	api.RunAPIServer(port)
}
