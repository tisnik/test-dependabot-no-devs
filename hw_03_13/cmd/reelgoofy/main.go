package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/course-go/reelgoofy/internal/containers/reviews/repository"
	reviewSeeders "github.com/course-go/reelgoofy/internal/containers/reviews/seeders"
	"github.com/course-go/reelgoofy/internal/core/server"
)

const seedReviews = 10

// main is the application entrypoint.
func main() {
	repo := repository.NewReviewRepository()
	config := argParse(repo)

	s := server.NewServer(*config, repo)
	s.Run()
}

// argParse is responsible for parsing arguments, mainly used for seeding and server settings.
func argParse(repo repository.ReviewRepository) *server.Config {
	config := &server.Config{
		Port:    server.DefaultPort,
		Timeout: server.DefaultTimeout,
	}

	if len(os.Args) <= 1 {
		return config
	}

	// Using for since its safer for this use-case (using index + 1)
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		switch arg {
		case "seed-reviews":
			seeder := reviewSeeders.NewReviewSeeder(repo)
			seeder.Seed(seedReviews)

		case "--port", "-p":
			if i+1 >= len(os.Args) {
				fmt.Fprintf(os.Stderr, "Error: %s requires a value\n", arg)
				os.Exit(1)
			}
			portStr := os.Args[i+1]
			port, err := strconv.ParseUint(portStr, 10, 16)
			if err != nil || port < 1 || port > 65535 {
				fmt.Fprintf(os.Stderr, "Error: Port must be within range 1-65535\n")
				os.Exit(1)
			}
			config.Port = portStr
			i++

		case "--timeout", "-t":
			if i+1 >= len(os.Args) {
				fmt.Fprintf(os.Stderr, "Error: %s requires a value\n", arg)
				os.Exit(1)
			}
			timeoutStr := os.Args[i+1]
			timeout, err := strconv.ParseUint(timeoutStr, 10, 32)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Timeout has to be a real number\n")
				os.Exit(1)
			}
			config.Timeout = uint(timeout)
			i++
		}
	}

	return config
}
