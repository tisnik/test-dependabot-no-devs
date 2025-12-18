package recommendation

type Recommendation struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

type ScoredRecommendation struct {
	Recommendation

	Score int
}

type ContentPathParams struct {
	ContentId string `json:"contentId"`
	Offset    string `json:"offset"`
	Limit     string `json:"limit"`
}

type UserPathParams struct {
	UserId string `json:"userId"`
	Offset string `json:"offset"`
	Limit  string `json:"limit"`
}

type UserProfile struct {
	FavoriteGenres    map[string]int
	FavoriteActors    map[string]int
	FavoriteDirectors map[string]int
}

type Film struct {
	Id       string
	Title    string
	Genres   []string
	Actors   []string
	Director string
}
