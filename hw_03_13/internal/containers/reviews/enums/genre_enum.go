package enums

// Genre represents an enum of supported genres (for content).
type Genre string

const (
	Horror   Genre = "Horror"
	SciFi    Genre = "Sci-Fi"
	Action   Genre = "Action"
	Drama    Genre = "Drama"
	Thriller Genre = "Thriller"
	Romance  Genre = "Romance"
)

var Genres = []Genre{
	Horror, SciFi, Action, Drama, Thriller, Romance,
}
