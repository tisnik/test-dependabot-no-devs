package utils

import (
	"slices"
)

func MergeUnique(firstField *[]string, secondField *[]string) *[]string {
	if firstField == nil {
		return secondField
	}

	if secondField == nil {
		return firstField
	}

	m := map[string]bool{}

	for _, value := range *firstField {
		m[value] = true
	}

	for _, value := range *secondField {
		m[value] = true
	}

	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}

	return &result
}

func ChooseMostFrequent(m map[string]int) *string {
	if len(m) == 0 {
		return nil
	}

	var maxCount int
	var item string

	for key, value := range m {
		if value > maxCount {
			maxCount = value
			item = key
		}
	}

	return &item
}

func Contains(sources *[]string, candidates *[]string) bool {
	if sources == nil || candidates == nil {
		return false
	}

	for _, source := range *sources {
		if slices.Contains(*candidates, source) {
			return true
		}
	}
	return false
}

func Paginate[T any](slices []T, offset *int, limit *int) []T {
	o := 0
	l := len(slices)

	if offset != nil {
		o = *offset
	}
	if limit != nil {
		l = *limit
	}

	if o >= len(slices) {
		return []T{}
	}

	end := min(o+l, len(slices))

	return slices[o:end]
}
