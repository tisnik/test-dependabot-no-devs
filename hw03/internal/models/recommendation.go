package models

type Recommendation struct {
	ContentId string `json:"id"`
	Title     string `json:"title"`
}

type RecommedationParameters struct {
	ContentId       string
	Title           string
	Actors          []string
	Origins         []string
	Genres          []string
	Tags            []string
	Directors       []string
	AvgScore        float64
	SimilarityScore float64
}

type RecommedationParametersForUser struct {
	ActorsWithAvgScore    map[string]float64
	OriginsWithAvgScore   map[string]float64
	GenresWithAvgScore    map[string]float64
	TagsWithAvgScore      map[string]float64
	DirectorsWithAvgScore map[string]float64
}
