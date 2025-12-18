package dto

import "github.com/google/uuid"

type MovieDataDTO struct {
	ContentId uuid.UUID
	Title     string
	Director  string
	Genres    []string
	Ratings   []int
}

func (m MovieDataDTO) AverageRating() int {
	total := 0
	for _, v := range m.Ratings {
		total += v
	}
	return total / len(m.Ratings)
}
