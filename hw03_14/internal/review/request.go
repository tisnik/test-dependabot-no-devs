package review

type Request struct {
	Data map[string][]RawReview `json:"data"`
}
