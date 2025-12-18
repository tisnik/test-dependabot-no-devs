package repository

import (
	"sync"

	"github.com/google/uuid"
	"github.com/medvedovan/reelgoofy-hw3/internal/model"
	"github.com/medvedovan/reelgoofy-hw3/internal/utils"
)

type Repository struct {
	mu              sync.Mutex
	reviews         []model.Review
	contentProfiles map[string]*ContentProfile
}

type ContentProfile struct {
	Actors   *[]string
	Director *string
	Genres   *[]string
	Score    int
	Tags     *[]string
	Title    *string
}

type UserPreferences struct {
	Genres        *[]string
	Directors     *[]string
	Actors        *[]string
	Tags          *[]string
	SeenContentId *[]string
}

const MinScore = 70

func NewRepository() *Repository {
	return &Repository{
		reviews:         make([]model.Review, 0),
		contentProfiles: make(map[string]*ContentProfile),
	}
}

func (r *Repository) GetReviews() (reviews []model.Review) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]model.Review(nil), r.reviews...)
}

func (r *Repository) GetContentProfiles() (contentProfiles map[string]*ContentProfile) {
	r.mu.Lock()
	defer r.mu.Unlock()

	contentProfiles = make(map[string]*ContentProfile, len(r.contentProfiles))
	for k, v := range r.contentProfiles {
		if v != nil {
			contentProfiles[k] = v
		}
	}
	return contentProfiles
}

func (r *Repository) CreateReview(rawReview *model.RawReview) (review model.Review) {
	if rawReview == nil {
		return model.Review{}
	}

	review = model.Review{
		Id:          uuid.NewString(),
		Actors:      rawReview.Actors,
		ContentId:   rawReview.ContentId,
		Description: rawReview.Description,
		Director:    rawReview.Director,
		Duration:    rawReview.Duration,
		Genres:      rawReview.Genres,
		Origins:     rawReview.Origins,
		Released:    rawReview.Released,
		Review:      rawReview.Review,
		Score:       rawReview.Score,
		Tags:        rawReview.Tags,
		Title:       rawReview.Title,
		UserId:      rawReview.UserId,
	}

	r.mu.Lock()
	r.reviews = append(r.reviews, review)
	r.mu.Unlock()

	r.updateContentProfile(review.ContentId)

	return review
}

func (r *Repository) DeleteReview(id string) (deleted bool) {
	r.mu.Lock()

	index := -1
	for i, review := range r.reviews {
		if review.Id == id {
			index = i
			break
		}
	}

	if index == -1 {
		r.mu.Unlock()
		return false
	}

	reviewToDelete := r.reviews[index]

	last := len(r.reviews) - 1
	r.reviews[index] = r.reviews[last]
	r.reviews = r.reviews[:last]
	r.mu.Unlock()

	r.updateContentProfile(reviewToDelete.ContentId)

	return true
}

func (r *Repository) AddReviews(reviews *[]model.Review) {
	if reviews == nil {
		return
	}

	for _, review := range *reviews {
		r.mu.Lock()
		r.reviews = append(r.reviews, review)
		r.mu.Unlock()
		r.updateContentProfile(review.ContentId)
	}
}

func (r *Repository) FilterReviewsByContentId(contentId string) *[]model.Review {
	filteredReviews := make([]model.Review, 0, len(r.reviews))

	r.mu.Lock()
	for _, review := range r.reviews {
		if review.ContentId == contentId {
			filteredReviews = append(filteredReviews, review)
		}
	}
	r.mu.Unlock()

	return &filteredReviews
}

func (r *Repository) FilterReviewsByUserId(userId string) *[]model.Review {
	filteredReviews := make([]model.Review, 0, len(r.reviews))

	r.mu.Lock()
	for _, review := range r.reviews {
		if review.UserId == userId {
			filteredReviews = append(filteredReviews, review)
		}
	}
	r.mu.Unlock()

	return &filteredReviews
}

func (r *Repository) GetUserPreferences(userId string) *UserPreferences {
	relevantReviews := r.FilterReviewsByUserId(userId)

	if len(*relevantReviews) == 0 {
		return nil
	}

	userPreferences := UserPreferences{
		Genres:        &[]string{},
		Tags:          &[]string{},
		Actors:        &[]string{},
		Directors:     &[]string{},
		SeenContentId: &[]string{},
	}

	for _, review := range *relevantReviews {
		userPreferences.SeenContentId = utils.MergeUnique(userPreferences.SeenContentId, &[]string{review.ContentId})

		if review.Score < MinScore {
			continue
		}

		userPreferences.Directors = utils.MergeUnique(userPreferences.Directors, &[]string{*review.Director})
		userPreferences.Genres = utils.MergeUnique(userPreferences.Genres, review.Genres)
		userPreferences.Actors = utils.MergeUnique(userPreferences.Actors, review.Actors)
		userPreferences.Tags = utils.MergeUnique(userPreferences.Tags, review.Tags)
	}

	return &userPreferences
}

func (r *Repository) GetContentProfileById(contentId string) *ContentProfile {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.contentProfiles[contentId]
}

func (r *Repository) updateContentProfile(contentId string) {
	relevantReviews := r.FilterReviewsByContentId(contentId)

	if len(*relevantReviews) == 0 {
		delete(r.contentProfiles, contentId)
		return
	}

	newProfile := ContentProfile{
		Genres: &[]string{},
		Tags:   &[]string{},
		Actors: &[]string{},
	}

	var totalScore int

	titles := map[string]int{}
	directors := map[string]int{}

	for _, review := range *relevantReviews {
		totalScore += review.Score

		if review.Title != nil && *review.Title != "" {
			titles[*review.Title]++
		}
		if review.Director != nil && *review.Director != "" {
			directors[*review.Director]++
		}

		newProfile.Genres = utils.MergeUnique(newProfile.Genres, review.Genres)
		newProfile.Actors = utils.MergeUnique(newProfile.Actors, review.Actors)
		newProfile.Tags = utils.MergeUnique(newProfile.Tags, review.Tags)
	}

	newProfile.Score = totalScore / len(*relevantReviews)

	newProfile.Title = utils.ChooseMostFrequent(titles)
	newProfile.Director = utils.ChooseMostFrequent(directors)

	r.mu.Lock()
	r.contentProfiles[contentId] = &newProfile
	r.mu.Unlock()
}
