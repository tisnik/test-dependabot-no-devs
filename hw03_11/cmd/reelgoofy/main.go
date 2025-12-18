package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/medvedovan/reelgoofy-hw3/internal/api"
	"github.com/medvedovan/reelgoofy-hw3/internal/config"
	"github.com/medvedovan/reelgoofy-hw3/internal/handler"
	"github.com/medvedovan/reelgoofy-hw3/internal/middleware"
	"github.com/medvedovan/reelgoofy-hw3/internal/repository"
	"github.com/medvedovan/reelgoofy-hw3/internal/server"
	"github.com/medvedovan/reelgoofy-hw3/internal/service"
)

var configPathFlag = flag.String("config", "config.yaml", "path to config file")

func main() {
	flag.Parse()

	config, err := config.NewConfig(configPathFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run ReelGoofy app: %v", err)
		os.Exit(1)
	}

	repo := repository.NewRepository()
	validate := validator.New(validator.WithRequiredStructEnabled())
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	revService := service.NewReviewService(logger, validate, repo)
	recService := service.NewRecommendationService(logger, validate, repo)

	recHandler := handler.NewRecommendationHandler(logger, recService)
	revHandler := handler.NewReviewHandler(logger, revService)
	serverHandler := server.NewServerHandler(recHandler, revHandler)

	middlewareFuncs := []api.MiddlewareFunc{
		middleware.ContentType,
	}

	serverOptions := api.ChiServerOptions{
		BaseURL:     "/" + config.Service.BaseURL,
		Middlewares: middlewareFuncs,
	}

	chiHandler := api.HandlerWithOptions(serverHandler, serverOptions)

	server := &http.Server{
		Addr:         ":" + config.Service.Port,
		Handler:      chiHandler,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	logger.Info("Server is listening.", "port", config.Service.Port)
	err = server.ListenAndServe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run ReelGoofy app: %v", err)
		os.Exit(1)
	}
}
