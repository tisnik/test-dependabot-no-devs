package api

import (
	"log/slog"

	"github.com/course-go/reelgoofy/internal/db"
)

const (
	qualityTreshold = 80.0
)

// API implements the StrictServerInterface.
type API struct {
	Database db.Database
	Logger   *slog.Logger
}
