package enums

// PopularCountry represents an enum of popular countries (as content origins).
type PopularCountry string

const (
	CZ  PopularCountry = "CZ"
	SK  PopularCountry = "SK"
	DE  PopularCountry = "DE"
	USA PopularCountry = "USA"
	UK  PopularCountry = "UK"
	FR  PopularCountry = "FR"
	JP  PopularCountry = "JP"
)

var Countries = []PopularCountry{
	CZ, SK, DE, USA, UK, FR, JP,
}
