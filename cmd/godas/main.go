package main

import (
	"context"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := NewRootCmd()
	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}