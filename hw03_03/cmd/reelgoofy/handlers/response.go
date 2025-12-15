package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/course-go/reelgoofy/cmd/reelgoofy/structs"
)

func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func RespondSuccess(w http.ResponseWriter, status int, reviews []structs.Review) {
	response := structs.ReviewsResponse{
		Status: "success",
	}
	response.Data.Reviews = reviews
	RespondJSON(w, status, response)
}

func RespondRecommendations(w http.ResponseWriter, recommendations []structs.Recommendation) {
	response := structs.RecommendationsResponse{
		Status: "success",
	}
	response.Data.Recommendations = recommendations
	RespondJSON(w, http.StatusOK, response)
}

func RespondFail(w http.ResponseWriter, status int, data map[string]any) {
	response := structs.FailResponse{
		Status: "fail",
		Data:   data,
	}
	RespondJSON(w, status, response)
}

func RespondError(w http.ResponseWriter, message string) {
	const internalServerError = 500
	response := structs.ErrorResponse{
		Status:  "error",
		Message: message,
		Code:    internalServerError,
	}
	RespondJSON(w, http.StatusInternalServerError, response)
}

func MethodNotAllowed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}
