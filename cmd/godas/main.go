package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := NewRootCmd()
	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
