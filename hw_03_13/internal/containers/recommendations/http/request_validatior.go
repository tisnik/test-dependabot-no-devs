package http

import (
	"net/http"
	"strconv"

	"github.com/course-go/reelgoofy/internal/core/response"
)

const (
	attributeLimit  = "limit"
	attributeOffset = "offset"
)

// RequestAttributes represents a collection of acceptable query attributes.
type RequestAttributes struct {
	Limit  *int `json:"limit"`
	Offset *int `json:"offset"`
}

// NewRequestAttributes is a construct for RequestAttributes.
func NewRequestAttributes() RequestAttributes {
	return RequestAttributes{
		Limit:  nil,
		Offset: nil,
	}
}

// GetValidatedAttributes validates allowed attributes present in a request.
func GetValidatedAttributes(req *http.Request) (RequestAttributes, map[string]string) {
	queryParameters := req.URL.Query()
	requestAttributes := NewRequestAttributes()
	errMap := make(map[string]string)

	if queryParameters.Has(attributeLimit) {
		parsedLimit, err := strconv.Atoi(queryParameters.Get(attributeLimit))
		if err != nil || parsedLimit < 0 {
			errMap[attributeLimit] = string(response.InvalidLimitMessage)
		}
		requestAttributes.Limit = &parsedLimit
	}

	if queryParameters.Has(attributeOffset) {
		parsedOffset, err := strconv.Atoi(queryParameters.Get(attributeOffset))
		if err != nil || parsedOffset < 0 {
			errMap[attributeOffset] = string(response.InvalidOffsetMessage)
		}
		requestAttributes.Offset = &parsedOffset
	}

	if len(errMap) > 0 {
		return requestAttributes, errMap
	}

	return requestAttributes, nil
}
