package test

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/medvedovan/reelgoofy-hw3/internal/api"
	"github.com/medvedovan/reelgoofy-hw3/internal/handler"
	"github.com/medvedovan/reelgoofy-hw3/internal/middleware"
	"github.com/medvedovan/reelgoofy-hw3/internal/model"
	"github.com/medvedovan/reelgoofy-hw3/internal/repository"
	"github.com/medvedovan/reelgoofy-hw3/internal/server"
	"github.com/medvedovan/reelgoofy-hw3/internal/service"
)

func NewHandler(repo *repository.Repository) http.Handler {
	seedRepo(repo)
	validate := validator.New(validator.WithRequiredStructEnabled())
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	revService := service.NewReviewService(logger, validate, repo)
	recService := service.NewRecommendationService(logger, validate, repo)

	recHandler := handler.NewRecommendationHandler(logger, recService)
	revHandler := handler.NewReviewHandler(logger, revService)
	serverHandler := server.NewServerHandler(recHandler, revHandler)

	middlewareFuncs := []api.MiddlewareFunc{
		middleware.ContentType,
	}

	serverOptions := api.ChiServerOptions{
		BaseURL:     "/api/v1",
		Middlewares: middlewareFuncs,
	}

	chiHandler := api.HandlerWithOptions(serverHandler, serverOptions)

	return chiHandler
}

type ErrorReader struct{}

func (e *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

//nolint:maintidx
func seedRepo(repo *repository.Repository) {
	reviews := []model.Review{
		{
			Id:        "1a111111-1111-1111-1111-111111111111",
			ContentId: "c1a1b111-1111-1111-1111-111111111111",
			UserId:    "d1111111-1111-1111-1111-111111111111",
			Title: ptrStr(
				"Inception",
			),
			Genres: &[]string{"sci-fi", "thriller"},
			Tags:   &[]string{"dream", "mind-bending"},
			Description: ptrStr(
				"A thief who steals corporate secrets through dream-sharing technology.",
			),
			Director: ptrStr(
				"Christopher Nolan",
			),
			Actors:  &[]string{"Leonardo DiCaprio", "Joseph Gordon-Levitt"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				8880, //nolint:mnd
			),
			Released: ptrStr("2010-07-16"),
			Review:   "Amazing visuals and story.",
			Score:    95, //nolint:mnd
		},

		{
			Id:        "2a222222-2222-2222-2222-222222222222",
			ContentId: "c2b2b222-2222-2222-2222-222222222222",
			UserId:    "d1111111-1111-1111-1111-111111111111",
			Title: ptrStr(
				"The Dark Knight",
			),
			Genres:      &[]string{"action", "crime"},
			Tags:        &[]string{"joker", "batman"},
			Description: ptrStr("Batman faces the Joker in Gotham City."),
			Director: ptrStr(
				"Christopher Nolan",
			),
			Actors:  &[]string{"Christian Bale", "Heath Ledger"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				9120, //nolint:mnd
			),
			Released: ptrStr("2008-07-18"),
			Review:   "Best superhero movie ever.",
			Score:    98, //nolint:mnd
		},

		{
			Id:        "3a333333-3333-3333-3333-333333333333",
			ContentId: "c3c3c333-3333-3333-3333-333333333333",
			UserId:    "d2222222-2222-2222-2222-222222222222",
			Title: ptrStr(
				"Jurassic Park",
			),
			Genres:      &[]string{"adventure", "sci-fi"},
			Tags:        &[]string{"dinosaurs", "action"},
			Description: ptrStr("Dinosaurs are brought back to life in a theme park."),
			Director: ptrStr(
				"Steven Spielberg",
			),
			Actors:  &[]string{"Sam Neill", "Laura Dern"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				7620, //nolint:mnd
			),
			Released: ptrStr("1993-06-11"),
			Review:   "Good fun, but a bit dated.",
			Score:    75, //nolint:mnd
		},

		{
			Id:        "4a444444-4444-4444-4444-444444444444",
			ContentId: "c4d4d444-4444-4444-4444-444444444444",
			UserId:    "d2222222-2222-2222-2222-222222222222",
			Title: ptrStr(
				"Avatar",
			),
			Genres:      &[]string{"sci-fi", "fantasy"},
			Tags:        &[]string{"aliens", "visuals"},
			Description: ptrStr("Humans colonize Pandora and clash with Na'vi."),
			Director: ptrStr(
				"James Cameron",
			),
			Actors:  &[]string{"Sam Worthington", "Zoe Saldana"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				9720, //nolint:mnd
			),
			Released: ptrStr("2009-12-18"),
			Review:   "Visually stunning but story is average.",
			Score:    69, //nolint:mnd
		},

		{
			Id:        "5a555555-5555-5555-5555-555555555555",
			ContentId: "c5e5e555-5555-5555-5555-555555555555",
			UserId:    "d2222222-2222-2222-2222-222222222222",
			Title: ptrStr(
				"Cats",
			),
			Genres:      &[]string{"musical", "fantasy"},
			Tags:        &[]string{"weird", "adaptation"},
			Description: ptrStr("Digital adaptation of the musical Cats."),
			Director: ptrStr(
				"Tom Hooper",
			),
			Actors:  &[]string{"Jennifer Hudson", "Ian McKellen"},
			Origins: &[]string{"UK", "USA"},
			Duration: ptrInt(
				6360, //nolint:mnd
			),
			Released: ptrStr("2019-12-20"),
			Review:   "Bizarre and uncomfortable to watch.",
			Score:    35, //nolint:mnd
		},

		{
			Id:        "6a666666-6666-6666-6666-666666666666",
			ContentId: "c16e16e16-6666-6666-6666-666666666666",
			UserId:    "d3333333-3333-3333-3333-333333333333",
			Title: ptrStr(
				"Movie 43",
			),
			Genres:      &[]string{"comedy"},
			Tags:        &[]string{"gross", "sketch"},
			Description: ptrStr("Series of interconnected comedy sketches."),
			Director: ptrStr(
				"Peter Farrelly",
			),
			Actors:  &[]string{"Hugh Jackman", "Kate Winslet"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				5940, //nolint:mnd
			),
			Released: ptrStr("2013-01-25"),
			Review:   "Not funny at all.",
			Score:    40, //nolint:mnd
		},

		{
			Id:        "7a777777-7777-7777-7777-777777777777",
			ContentId: "c7e7e777-7777-7777-7777-777777777777",
			UserId:    "d4444444-4444-4444-4444-444444444444",
			Title: ptrStr(
				"Spider-Man: Homecoming",
			),
			Genres:      &[]string{"action", "adventure"},
			Tags:        &[]string{"spiderman", "marvel"},
			Description: ptrStr("Peter Parker balances high school life and superhero duties."),
			Director: ptrStr(
				"Jon Watts",
			),
			Actors:  &[]string{"Tom Holland", "Michael Keaton"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				7440, //nolint:mnd
			),
			Released: ptrStr("2017-07-07"),
			Review:   "Good but predictable.",
			Score:    70, //nolint:mnd
		},

		{
			Id:        "8a888888-8888-8888-8888-888888888888",
			ContentId: "c8f8f888-8888-8888-8888-888888888888",
			UserId:    "d2222222-2222-2222-2222-222222222222",
			Title: ptrStr(
				"Venom",
			),
			Genres:      &[]string{"action", "sci-fi"},
			Tags:        &[]string{"antihero", "marvel"},
			Description: ptrStr("A journalist becomes the host for a symbiote."),
			Director: ptrStr(
				"Ruben Fleischer",
			),
			Actors:  &[]string{"Tom Hardy", "Michelle Williams"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				6840, //nolint:mnd
			),
			Released: ptrStr("2018-10-05"),
			Review:   "Not bad, but story is weak.",
			Score:    60, //nolint:mnd
		},

		{
			Id:        "9a999999-9999-9999-9999-999999999999",
			ContentId: "c9e9e999-9999-9999-9999-999999999999",
			UserId:    "d5555555-5555-5555-5555-555555555555",
			Title: ptrStr(
				"Forrest Gump",
			),
			Genres:      &[]string{"drama", "romance"},
			Tags:        &[]string{"life", "historical"},
			Description: ptrStr("Life story of Forrest Gump."),
			Director: ptrStr(
				"Robert Zemeckis",
			),
			Actors:  &[]string{"Tom Hanks", "Robin Wright"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				8520, //nolint:mnd
			),
			Released: ptrStr("1994-07-06"),
			Review:   "Heartwarming.",
			Score:    96, //nolint:mnd
		},

		{
			Id:        "10a101010-1010-1010-1010-101010101010",
			ContentId: "c10e10e10-1010-1010-1010-101010101010",
			UserId:    "d2222222-2222-2222-2222-222222222222",
			Title: ptrStr(
				"Gladiator",
			),
			Genres:      &[]string{"action", "drama"},
			Tags:        &[]string{"roman empire", "revenge"},
			Description: ptrStr("A former general seeks revenge."),
			Director: ptrStr(
				"Ridley Scott",
			),
			Actors:  &[]string{"Russell Crowe", "Joaquin Phoenix"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				9300, //nolint:mnd
			),
			Released: ptrStr("2000-05-05"),
			Review:   "Epic and intense.",
			Score:    94, //nolint:mnd
		},

		{
			Id:        "11b111111-1111-1111-1111-111111111111",
			ContentId: "3f2504e0-4f89-41d3-9a0c-0305e82c3301",
			UserId:    "d6666666-6666-6666-6666-666666666666",
			Title: ptrStr(
				"Titanic",
			),
			Genres:      &[]string{"drama", "romance"},
			Tags:        &[]string{"ship", "tragedy"},
			Description: ptrStr("Love story on the doomed ship."),
			Director: ptrStr(
				"James Cameron",
			),
			Actors:  &[]string{"Leonardo DiCaprio", "Kate Winslet"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				11100, //nolint:mnd
			),
			Released: ptrStr("1997-12-19"),
			Review:   "Emotional and beautiful.",
			Score:    90, //nolint:mnd
		},

		{
			Id:        "12b222222-2222-2222-2222-222222222222",
			ContentId: "c12f12f12-2222-2222-2222-222222222222",
			UserId:    "d6666666-6666-6666-6666-666666666666",
			Title: ptrStr(
				"Avatar 2",
			),
			Genres:      &[]string{"sci-fi", "fantasy"},
			Tags:        &[]string{"underwater", "visuals"},
			Description: ptrStr("Return to Pandora."),
			Director: ptrStr(
				"James Cameron",
			),
			Actors:  &[]string{"Sam Worthington", "Zoe Saldana"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				10500, //nolint:mnd
			),
			Released: ptrStr("2022-12-16"),
			Review:   "Beautiful visuals.",
			Score:    67, //nolint:mnd
		},

		{
			Id:        "13b333333-3333-3333-3333-333333333333",
			ContentId: "c13f13f13-3333-3333-3333-333333333333",
			UserId:    "d7777777-7777-7777-7777-777777777777",
			Title: ptrStr(
				"The Matrix",
			),
			Genres:      &[]string{"sci-fi", "action"},
			Tags:        &[]string{"simulation", "revolution"},
			Description: ptrStr("A hacker learns the truth about reality."),
			Director: ptrStr(
				"The Wachowskis",
			),
			Actors:  &[]string{"Keanu Reeves", "Laurence Fishburne"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				8160, //nolint:mnd
			),
			Released: ptrStr("1999-03-31"),
			Review:   "Mind-blowing sci-fi.",
			Score:    97, //nolint:mnd
		},

		{
			Id:        "14b444444-4444-4444-4444-444444444444",
			ContentId: "c14f14f14-4444-4444-4444-444444444444",
			UserId:    "d6666666-6666-6666-6666-666666666666",
			Title: ptrStr(
				"The Matrix Reloaded",
			),
			Genres:      &[]string{"sci-fi", "action"},
			Tags:        &[]string{"simulation", "sequel"},
			Description: ptrStr("The fight continues."),
			Director: ptrStr(
				"The Wachowskis",
			),
			Actors:  &[]string{"Keanu Reeves", "Carrie-Anne Moss"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				8280, //nolint:mnd
			),
			Released: ptrStr("2003-05-15"),
			Review:   "Good sequel.",
			Score:    50, //nolint:mnd
		},

		{
			Id:        "15b555555-5555-5555-5555-555555555555",
			ContentId: "c15f15f15-5555-5555-5555-555555555555",
			UserId:    "d8888888-8888-8888-8888-888888888888",
			Title: ptrStr(
				"The Room",
			),
			Genres:      &[]string{"drama"},
			Tags:        &[]string{"cult", "so bad it's good"},
			Description: ptrStr("Life and troubles of Johnny."),
			Director: ptrStr(
				"Tommy Wiseau",
			),
			Actors:  &[]string{"Tommy Wiseau", "Juliette Danielle"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				5460, //nolint:mnd
			),
			Released: ptrStr("2003-06-27"),
			Review:   "Hilariously bad.",
			Score:    15, //nolint:mnd
		},

		{
			Id:        "16b666666-6666-6666-6666-666666666666",
			ContentId: "c16e16e16-6666-6666-6666-666666666666",
			UserId:    "d6666666-6666-6666-6666-666666666666",
			Title: ptrStr(
				"Movie 43",
			),
			Genres:      &[]string{"comedy"},
			Tags:        &[]string{"gross", "sketch"},
			Description: ptrStr("Series of interconnected comedy sketches."),
			Director: ptrStr(
				"Peter Farrelly",
			),
			Actors:  &[]string{"Hugh Jackman", "Kate Winslet"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				5940, //nolint:mnd
			),
			Released: ptrStr("2013-01-25"),
			Review:   "Terrible jokes, not funny.",
			Score:    20, //nolint:mnd
		},

		{
			Id:        "17b777777-7777-7777-7777-777777777777",
			ContentId: "c17f17f17-7777-7777-7777-777777777777",
			UserId:    "3f8c2c7e-4b1e-4e3d-9af1-0d2d4d8b2c11",
			Title: ptrStr(
				"Avengers: Endgame",
			),
			Genres:      &[]string{"action", "sci-fi"},
			Tags:        &[]string{"marvel", "avengers"},
			Description: ptrStr("Superheroes try to undo Thanos's actions."),
			Director: ptrStr(
				"Anthony Russo",
			),
			Actors:  &[]string{"Robert Downey Jr.", "Chris Evans"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				10200, //nolint:mnd
			),
			Released: ptrStr("2019-04-26"),
			Review:   "Epic conclusion.",
			Score:    92, //nolint:mnd
		},

		{
			Id:        "18b888888-8888-8888-8888-888888888888",
			ContentId: "c18f18f18-8888-8888-8888-888888888888",
			UserId:    "3f8c2c7e-4b1e-4e3d-9af1-0d2d4d8b2c11",
			Title: ptrStr(
				"Avengers: Infinity War",
			),
			Genres:      &[]string{"action", "sci-fi"},
			Tags:        &[]string{"marvel", "thanos"},
			Description: ptrStr("Heroes fight Thanos."),
			Director: ptrStr(
				"Anthony Russo",
			),
			Actors:  &[]string{"Robert Downey Jr.", "Chris Hemsworth"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				10140, //nolint:mnd
			),
			Released: ptrStr("2018-04-27"),
			Review:   "High stakes action.",
			Score:    88, //nolint:mnd
		},

		{
			Id:        "19b999999-9999-9999-9999-999999999999",
			ContentId: "c19f19f19-9999-9999-9999-999999999999",
			UserId:    "d10101010-1010-1010-1010-101010101010",
			Title: ptrStr(
				"The Lion King",
			),
			Genres:      &[]string{"animation"},
			Tags:        &[]string{"disney", "classic"},
			Description: ptrStr("Simba's journey to become king."),
			Director: ptrStr(
				"Roger Allers",
			),
			Actors:  &[]string{"Matthew Broderick", "Jeremy Irons"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				5340, //nolint:mnd
			),
			Released: ptrStr("1994-06-15"),
			Review:   "Beautiful animation.",
			Score:    90, //nolint:mnd
		},

		{
			Id:        "20b101010-1010-1010-1010-101010101010",
			ContentId: "c20f20f20-1010-1010-1010-101010101010",
			UserId:    "d10101010-1010-1010-1010-101010101010",
			Title: ptrStr(
				"Toy Story",
			),
			Genres:      &[]string{"animation", "family"},
			Tags:        &[]string{"toys", "adventure"},
			Description: ptrStr("Toys come to life when humans are away."),
			Director: ptrStr(
				"John Lasseter",
			),
			Actors:  &[]string{"Tom Hanks", "Tim Allen"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				4860, //nolint:mnd
			),
			Released: ptrStr("1995-11-22"),
			Review:   "Classic and fun for all ages.",
			Score:    69, //nolint:mnd
		},
		{
			Id:        "30c202020-2020-2020-2020-202020202020",
			ContentId: "e4eaaaf2-d142-11e1-b3e4-080027620cdd",
			UserId:    "d20202020-2020-2020-2020-202020202020",
			Title: ptrStr(
				"The Conjuring",
			),
			Genres: &[]string{"horror"},
			Tags:   &[]string{"haunted house", "ghosts"},
			Description: ptrStr(
				"A family is terrorized by a dark presence in their farmhouse.",
			),
			Director: ptrStr(
				"James Wan",
			),
			Actors:  &[]string{"Vera Farmiga", "Patrick Wilson"},
			Origins: &[]string{"USA"},
			Duration: ptrInt(
				6720, //nolint:mnd
			),
			Released: ptrStr("2013-07-19"),
			Review:   "Terrifying and expertly crafted, delivering relentless tension and unforgettable scares.",
			Score:    67, //nolint:mnd
		},
	}

	repo.AddReviews(&reviews)
}

func ptrInt(i int) *int       { return &i }
func ptrStr(s string) *string { return &s }
