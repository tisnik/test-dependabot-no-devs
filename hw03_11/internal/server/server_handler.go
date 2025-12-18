package server

import (
	"net/http"

	"github.com/medvedovan/reelgoofy-hw3/internal/handler"
	"github.com/medvedovan/reelgoofy-hw3/internal/model"
)

type ServerHandler struct {
	recommendationHandler *handler.RecommendationHandler
	reviewHandler         *handler.ReviewHandler
}

func NewServerHandler(
	recHandler *handler.RecommendationHandler,
	revHandler *handler.ReviewHandler,
) (server *ServerHandler) {
	return &ServerHandler{
		recommendationHandler: recHandler,
		reviewHandler:         revHandler,
	}
}

func (s *ServerHandler) RecommendContentToContent(
	w http.ResponseWriter,
	r *http.Request,
	contentId string,
	params model.RecommendContentToContentParams,
) {
	s.recommendationHandler.RecommendContentToContent(w, r, contentId, params)
}

func (s *ServerHandler) RecommendContentToUser(
	w http.ResponseWriter,
	r *http.Request,
	userId string,
	params model.RecommendContentToUserParams,
) {
	s.recommendationHandler.RecommendContentToUser(w, r, userId, params)
}

func (s *ServerHandler) IngestReviews(w http.ResponseWriter, r *http.Request) {
	s.reviewHandler.IngestReviews(w, r)
}

func (s *ServerHandler) DeleteReview(w http.ResponseWriter, r *http.Request, reviewId string) {
	s.reviewHandler.DeleteReview(w, r, reviewId)
}
