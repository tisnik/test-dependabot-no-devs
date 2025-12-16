package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/course-go/reelgoofy/internal/domain"
)

func WriteFailResponse(w http.ResponseWriter, statusCode int, errors map[string]string) {
	w.WriteHeader(statusCode)

	response := domain.FailResponse{
		Status: "fail",
		Data:   errors,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode JSON response: %v\n", err)
	}
}

func WriteErrorResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)

	response := domain.ErrorResponse{
		Status:  "error",
		Message: "Unable to communicate with database",
		Code:    http.StatusInternalServerError,
		Data:    map[string]string{},
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode JSON response: %v\n", err)
	}
}
