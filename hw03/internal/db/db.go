package db

import (
	"sync"

	"github.com/course-go/reelgoofy/internal/models"
)

type MemoryDB struct {
	Mu      sync.RWMutex
	Reviews map[string]models.RawReview
}

func NewDB() *MemoryDB {
	return &MemoryDB{
		Reviews: make(map[string]models.RawReview),
	}
}
