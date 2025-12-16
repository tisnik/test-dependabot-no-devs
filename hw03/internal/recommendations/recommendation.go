package recommendations

type Recommendation struct {
	ID    string `json:"id"`
	Title string `json:"title,omitzero"`
}

type Recommendations struct {
	Recommendations []Recommendation `json:"recommendations"`
}
