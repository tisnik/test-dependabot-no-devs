package http

// ReviewResponse is used as a response layout for reviews.
type ReviewResponse struct {
	Reviews any `json:"reviews,omitempty"`
}

// Format formats the response to requested format.
func Format(data any) ReviewResponse {
	return ReviewResponse{
		Reviews: data,
	}
}
