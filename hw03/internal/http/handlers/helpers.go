package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	apperrors "github.com/course-go/reelgoofy/internal/errors"
	"github.com/course-go/reelgoofy/internal/http/dto"
	"github.com/course-go/reelgoofy/internal/status"
	"github.com/google/uuid"
)

const (
	defaultLimit  = 20
	defaultOffset = 0
)

func writeJSON(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		panic(err)
	}
}

func writeSuccess[T any](w http.ResponseWriter, statusCode int, data T) {
	resp := dto.Response[T]{
		Status: status.Success,
		Data:   data,
	}
	writeJSON(w, statusCode, resp)
}

func writeFail(w http.ResponseWriter, statusCode int, data map[string]string) {
	resp := dto.FailResponse{
		Status: status.Fail,
		Data:   data,
	}
	writeJSON(w, statusCode, resp)
}

func writeError(w http.ResponseWriter, statusCode int, msg string) {
	resp := dto.ErrorResponse{
		Status:  status.Error,
		Message: msg,
		Code:    statusCode,
		Data:    nil,
	}
	writeJSON(w, statusCode, resp)
}

func respondWithError(w http.ResponseWriter, err error) {
	var svcErr apperrors.ServiceError

	if errors.As(err, &svcErr) {
		writeFail(w, svcErr.Code, dto.FailDataDTO{
			svcErr.Field: svcErr.Message,
		})
		return
	}

	writeError(w, http.StatusInternalServerError, "internal server error")
}

func getPaginationParams(r *http.Request) (int, int) {
	limit := getQueryInt(r, "limit", defaultLimit)
	offset := getQueryInt(r, "offset", defaultOffset)
	return limit, offset
}

func getQueryInt(r *http.Request, key string, defaultValue int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}

	return intVal
}

func validateUUID(field, id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return apperrors.NewValidationError(field, "must be a valid UUID")
	}
	return nil
}
