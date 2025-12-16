package tests_test

import "github.com/course-go/reelgoofy/internal/domain"

var (
	// User 1.

	ReviewA = domain.Review{
		ID: "137b33bf-066a-44f7-9a9b-d65071d27270",
		RawReview: domain.RawReview{
			ContentID:   "937b33bf-066a-44f7-9a9b-d65071d27270",
			UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa98",
			Title:       "One Flew over the Cuckoo's Nest",
			Genres:      []string{"drama"},
			Tags:        []string{"suicide"},
			Description: "A movie about gangsters.",
			Director:    "Christopher Nolan",
			Actors:      []string{"Tim Robbins"},
			Origins:     []string{"USA"},
			Duration:    8520,
			Released:    "2022-09-13",
			Review:      "I really enjoyed this one.",
			Score:       75,
		},
	}

	ReviewB = domain.Review{
		ID: "237b33bf-066a-44f7-9a9b-d65071d27270",
		RawReview: domain.RawReview{
			ContentID:   "937b33bf-066a-44f7-9a9b-d65071d27273",
			UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa98",
			Title:       "Star Wars: Rise of Skywalker",
			Genres:      []string{"drama"},
			Tags:        []string{"suicide"},
			Description: "A movie about gangsters.",
			Director:    "JJ Abrams",
			Actors:      []string{"Tim Robbins"},
			Origins:     []string{"USA"},
			Duration:    890,
			Released:    "2022-09-13",
			Review:      "I really enjoyed this one.",
			Score:       10,
		},
	}

	ReviewC = domain.Review{
		ID: "237b33bf-066a-44f7-9a9b-d65071d27278",
		RawReview: domain.RawReview{
			ContentID:   "937b33de-066a-44f7-9a9b-d65071d27280",
			UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa98",
			Title:       "Trainspotting",
			Genres:      []string{"drama"},
			Tags:        []string{"suicide"},
			Description: "Family comedy",
			Director:    "Snorty Marty",
			Actors:      []string{"Tim Robbins"},
			Origins:     []string{"USA"},
			Duration:    8520,
			Released:    "2022-09-13",
			Review:      "I really enjoyed this one.",
			Score:       60,
		},
	}

	// User 2.

	ReviewD = domain.Review{
		ID: "237b33bf-066a-44f7-9a9b-d65071d27200",
		RawReview: domain.RawReview{
			ContentID:   "937b33bf-066a-44f7-9a9b-d65071d27200",
			UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa00",
			Title:       "The banker",
			Genres:      []string{"drama"},
			Tags:        []string{"suicide"},
			Description: "Horror",
			Director:    "Drzgros McDuck",
			Actors:      []string{"Tim Robbins"},
			Origins:     []string{"USA"},
			Duration:    8520,
			Released:    "2022-09-13",
			Review:      "I really enjoyed this one.",
			Score:       80,
		},
	}

	ReviewE = domain.Review{
		ID: "137b33bf-066a-44f7-9a9b-d65071d27270",
		RawReview: domain.RawReview{
			ContentID:   "937b33bf-066a-44f7-9a9b-d65071d27270",
			UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa00",
			Title:       "One Flew over the Cuckoo's Nest",
			Genres:      []string{"drama"},
			Tags:        []string{"suicide"},
			Description: "A movie about gangsters.",
			Director:    "Christopher Nolan",
			Actors:      []string{"Tim Robbins"},
			Origins:     []string{"USA"},
			Duration:    8520,
			Released:    "2022-09-13",
			Review:      "I really enjoyed this one.",
			Score:       75,
		},
	}

	ReviewF = domain.Review{
		ID: "237b33bf-066a-44f7-9a9b-d65071d27201",
		RawReview: domain.RawReview{
			ContentID:   "937b33bf-066a-44f7-9a9b-d65071d27273",
			UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa00",
			Title:       "Star Wars: Rise of Skywalker",
			Genres:      []string{"drama"},
			Tags:        []string{"suicide"},
			Description: "A movie about gangsters.",
			Director:    "JJ Abrams",
			Actors:      []string{"Tim Robbins"},
			Origins:     []string{"USA"},
			Duration:    890,
			Released:    "2022-09-13",
			Review:      "I really enjoyed this one.",
			Score:       20,
		},
	}

	// User 3.

	ReviewG = domain.Review{
		ID: "237b33bf-066a-44f7-9a9b-d65071d27201",
		RawReview: domain.RawReview{
			ContentID:   "937b33bf-066a-44f7-9a9b-d65071d27273",
			UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa05",
			Title:       "Star Wars: Rise of Skywalker",
			Genres:      []string{"drama"},
			Tags:        []string{"suicide"},
			Description: "A movie about gangsters.",
			Director:    "JJ Abrams",
			Actors:      []string{"Tim Robbins"},
			Origins:     []string{"USA"},
			Duration:    890,
			Released:    "2022-09-13",
			Review:      "I really enjoyed this one.",
			Score:       30,
		},
	}

	ReviewH = domain.Review{
		ID: "237b33bf-066a-44f7-9a9b-d65071d27201",
		RawReview: domain.RawReview{
			ContentID:   "937b33bf-066a-44f7-9a9b-d65071d27299",
			UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa05",
			Title:       "Treasure planet",
			Genres:      []string{"drama"},
			Tags:        []string{"suicide"},
			Description: "A movie about gangsters.",
			Director:    "JJ Abrams",
			Actors:      []string{"Tim Robbins"},
			Origins:     []string{"USA"},
			Duration:    890,
			Released:    "2022-09-13",
			Review:      "I really enjoyed this one.",
			Score:       90,
		},
	}

	ReviewI = domain.Review{
		ID: "237b33bf-066a-44f7-9a9b-d65071d27201",
		RawReview: domain.RawReview{
			ContentID:   "937b33bf-066a-44f7-9a9b-d65071d27233",
			UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa05",
			Title:       "Schindlers list",
			Genres:      []string{"drama"},
			Tags:        []string{"suicide"},
			Description: "Sci fi",
			Director:    "JJ Abrams",
			Actors:      []string{"Tim Robbins"},
			Origins:     []string{"USA"},
			Duration:    890,
			Released:    "2022-09-13",
			Review:      "I really enjoyed this one.",
			Score:       100,
		},
	}

	ReviewJ = domain.Review{
		ID: "237b33bf-066a-44f7-9a9b-d65071d27201",
		RawReview: domain.RawReview{
			ContentID:   "937b33bf-066a-44f7-9a9b-d65071d27244",
			UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa05",
			Title:       "Sharkando",
			Genres:      []string{"drama"},
			Tags:        []string{"suicide"},
			Description: "Tragedy",
			Director:    "JJ Abrams",
			Actors:      []string{"Tim Robbins"},
			Origins:     []string{"USA"},
			Duration:    890,
			Released:    "2022-09-13",
			Review:      "I really enjoyed this one.",
			Score:       5,
		},
	}

	RawReviewA = domain.RawReview{
		ContentID:   "837b33bf-066a-44f7-9a9b-d65071d27270",
		UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa98",
		Title:       "One Flew over the Cuckoo's Nest",
		Genres:      []string{"drama"},
		Tags:        []string{"suicide"},
		Description: "A movie about gangsters.",
		Director:    "Christopher Nolan",
		Actors:      []string{"Tim Robbins"},
		Origins:     []string{"USA"},
		Duration:    8520,
		Released:    "2022-09-13",
		Review:      "I really enjoyed this one.",
		Score:       75,
	}

	RawReviewB = domain.RawReview{
		ContentID:   "937b33bf-066a-44f7-9a9b-d65071d27273",
		UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa98",
		Title:       "Star Wars: Rise of Skywalker",
		Genres:      []string{"drama"},
		Tags:        []string{"suicide"},
		Description: "A movie about gangsters.",
		Director:    "JJ Abrams",
		Actors:      []string{"Tim Robbins"},
		Origins:     []string{"USA"},
		Duration:    890,
		Released:    "2022-09-13",
		Review:      "I really enjoyed this one.",
		Score:       10,
	}

	RawReviewD = domain.RawReview{
		ContentID:   "737b33bf-066a-44f7-9a9b-d65071d27273",
		UserID:      "2f99df7d-751c-40c9-aeea-8be8cd7bfa98",
		Title:       "Star Wars: Rise of Skywalker",
		Genres:      []string{"drama"},
		Tags:        []string{"suicide"},
		Description: "A movie about gangsters.",
		Director:    "JJ Abrams",
		Actors:      []string{"Tim Robbins"},
		Origins:     []string{"USA"},
		Duration:    890,
		Released:    "2022-09-13",
		Review:      "I really enjoyed this one.",
		Score:       10,
	}

	RawReviewC = domain.RawReview{
		ContentID:   "637b33bf-066a-44f7-9a9b-d65071d27273",
		UserID:      "3f99df7d-751c-40c9-aeea-8be8cd7bfa98",
		Title:       "Trainspotting",
		Genres:      []string{"drama"},
		Tags:        []string{"suicide"},
		Description: "Family comedy",
		Director:    "Snorty Marty",
		Actors:      []string{"Tim Robbins"},
		Origins:     []string{"USA"},
		Duration:    8520,
		Released:    "2022-09-13",
		Review:      "I really enjoyed this one.",
		Score:       60,
	}
)
