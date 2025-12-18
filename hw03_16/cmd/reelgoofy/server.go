package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

const (
	genreWeight     = 3.0
	tagWeight       = 2.0
	directorWeight  = 1.5
	scoreNormalizer = 100.0
)

type jsonResponse struct {
	Status string `json:"status"`
	Data   any    `json:"data,omitempty"`
}

type Server struct {
	storage *Storage
}

func NewServer() *Server {
	return &Server{
		storage: NewStorage(),
	}
}

func (s *Server) IngestReviews(w http.ResponseWriter, r *http.Request) {
	var req RawReviewsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendFail(w, http.StatusBadRequest, map[string]any{
			"request": "Invalid JSON",
		})
		return
	}

	if req.Data == nil || req.Data.Reviews == nil {
		sendFail(w, http.StatusBadRequest, map[string]any{
			"reviews": "Missing reviews",
		})
		return
	}

	for _, rawReview := range *req.Data.Reviews {
		_, err = uuid.Parse(rawReview.ContentId)
		if err != nil {
			sendFail(w, http.StatusBadRequest, map[string]any{
				"contentId": "ID is not a valid UUID.",
			})
			return
		}
		_, err = uuid.Parse(rawReview.UserId)
		if err != nil {
			sendFail(w, http.StatusBadRequest, map[string]any{
				"userId": "ID is not a valid UUID.",
			})
			return
		}
		if rawReview.Review == "" {
			sendFail(w, http.StatusBadRequest, map[string]any{
				"review": "Review is required",
			})
			return
		}
	}

	reviews := make([]Review, 0)
	for _, rawReview := range *req.Data.Reviews {
		review := Review{
			Id:          uuid.New().String(),
			ContentId:   rawReview.ContentId,
			UserId:      rawReview.UserId,
			Title:       rawReview.Title,
			Genres:      rawReview.Genres,
			Tags:        rawReview.Tags,
			Description: rawReview.Description,
			Director:    rawReview.Director,
			Actors:      rawReview.Actors,
			Origins:     rawReview.Origins,
			Duration:    rawReview.Duration,
			Released:    rawReview.Released,
			Review:      rawReview.Review,
			Score:       rawReview.Score,
		}
		s.storage.Add(&review)
		reviews = append(reviews, review)
	}

	sendSuccess(w, http.StatusCreated, Reviews{Reviews: &reviews})
}

func (s *Server) DeleteReview(w http.ResponseWriter, r *http.Request, reviewId string) {
	_, err := uuid.Parse(reviewId)
	if err != nil {
		sendFail(w, http.StatusBadRequest, map[string]any{
			"reviewId": "ID is not a valid UUID.",
		})
		return
	}

	if !s.storage.Delete(reviewId) {
		sendFail(w, http.StatusNotFound, map[string]any{
			"reviewId": "Review with such ID not found.",
		})
		return
	}

	sendSuccess(w, http.StatusOK, nil)
}

func (s *Server) RecommendContentToContent(
	w http.ResponseWriter,
	r *http.Request,
	contentId string,
	params RecommendContentToContentParams,
) {
	_, err := uuid.Parse(contentId)
	if err != nil {
		sendFail(w, http.StatusBadRequest, map[string]any{
			"contentId": "ID is not a valid UUID.",
		})
		return
	}

	limit := 100
	offset := 0
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}

	recommendations := s.getContentRecommendations(contentId, limit, offset)
	sendSuccess(w, http.StatusOK, Recommendations{Recommendations: &recommendations})
}

func (s *Server) RecommendContentToUser(
	w http.ResponseWriter,
	r *http.Request,
	userId string,
	params RecommendContentToUserParams,
) {
	_, err := uuid.Parse(userId)
	if err != nil {
		sendFail(w, http.StatusBadRequest, map[string]any{
			"userId": "ID is not a valid UUID.",
		})
		return
	}

	limit := 100
	offset := 0
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}

	recommendations := s.getUserRecommendations(userId, limit, offset)
	sendSuccess(w, http.StatusOK, Recommendations{Recommendations: &recommendations})
}

func sendSuccess(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(jsonResponse{
		Status: "success",
		Data:   data,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sendFail(w http.ResponseWriter, code int, data map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(jsonResponse{
		Status: "fail",
		Data:   data,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) getContentRecommendations(contentId string, limit, offset int) []Recommendation {
	sourceReviews := s.storage.GetByContentID(contentId)
	if len(sourceReviews) == 0 {
		return []Recommendation{}
	}

	allReviews := s.storage.GetAll()
	scores := make(map[string]float64)
	titles := make(map[string]string)

	for _, review := range allReviews {
		if review.ContentId == contentId {
			continue
		}

		for _, source := range sourceReviews {
			score := calculateSimilarity(source, review)
			if score > 0 {
				scores[review.ContentId] += score
				if review.Title != nil {
					titles[review.ContentId] = *review.Title
				}
			}
		}
	}

	return buildRecommendations(scores, titles, limit, offset)
}

func (s *Server) getUserRecommendations(userId string, limit, offset int) []Recommendation {
	userReviews := s.storage.GetByUserID(userId)
	if len(userReviews) == 0 {
		return []Recommendation{}
	}

	allReviews := s.storage.GetAll()
	scores := make(map[string]float64)
	titles := make(map[string]string)

	for _, review := range allReviews {
		alreadyReviewed := false
		for _, userReview := range userReviews {
			if userReview.ContentId == review.ContentId {
				alreadyReviewed = true
				break
			}
		}
		if alreadyReviewed {
			continue
		}

		for _, userReview := range userReviews {
			score := calculateSimilarity(userReview, review)
			if score > 0 {
				score *= float64(userReview.Score) / scoreNormalizer
				scores[review.ContentId] += score
				if review.Title != nil {
					titles[review.ContentId] = *review.Title
				}
			}
		}
	}

	return buildRecommendations(scores, titles, limit, offset)
}

func calculateSimilarity(r1, r2 *Review) float64 {
	score := 0.0

	if r1.Genres != nil && r2.Genres != nil {
		matches := 0
		for _, g1 := range *r1.Genres {
			for _, g2 := range *r2.Genres {
				if g1 == g2 {
					matches++
				}
			}
		}
		if len(*r1.Genres) > 0 {
			score += float64(matches) / float64(len(*r1.Genres)) * genreWeight
		}
	}

	if r1.Tags != nil && r2.Tags != nil {
		matches := 0
		for _, t1 := range *r1.Tags {
			for _, t2 := range *r2.Tags {
				if t1 == t2 {
					matches++
				}
			}
		}
		if len(*r1.Tags) > 0 {
			score += float64(matches) / float64(len(*r1.Tags)) * tagWeight
		}
	}

	if r1.Director != nil && r2.Director != nil && *r1.Director == *r2.Director {
		score += directorWeight
	}

	return score
}

func buildRecommendations(scores map[string]float64, titles map[string]string, limit, offset int) []Recommendation {
	type scored struct {
		id    string
		score float64
	}

	items := make([]scored, 0, len(scores))
	for id, score := range scores {
		if score > 0 {
			items = append(items, scored{id: id, score: score})
		}
	}

	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].score > items[i].score {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	start := offset
	if start > len(items) {
		return []Recommendation{}
	}

	end := min(start+limit, len(items))

	recommendations := make([]Recommendation, 0, end-start)
	for i := start; i < end; i++ {
		rec := Recommendation{
			Id: items[i].id,
		}
		if title, ok := titles[items[i].id]; ok {
			rec.Title = &title
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}
