package main

import (
	"context"
	"log"

	"github.com/course-go/reelgoofy/internal/app"
)

func main() {
	ctx := context.Background()
	err := app.Run(ctx)
	if err != nil {
		log.Printf("reelgoofy service stopped: %s", err)
	}
}
