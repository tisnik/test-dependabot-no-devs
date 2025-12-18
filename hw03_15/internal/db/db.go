package db

import (
	"github.com/google/uuid"
)

type Review struct {
	ReviewId  uuid.UUID
	ContentId uuid.UUID
	Title     *string
	Genres    *[]string
	Score     int
}

type Database map[uuid.UUID][]Review

func CreateDb() Database {
	return make(Database)
}
