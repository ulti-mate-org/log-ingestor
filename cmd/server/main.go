package main

import (
	"fmt"
	"github.com/ulti-mate-org/logging/internal/api"
)

func main() {
	fmt.Println("logging service starting...")
	api.ListenAndServe(":8080")
	fmt.Println("logging service stopped.")
}
