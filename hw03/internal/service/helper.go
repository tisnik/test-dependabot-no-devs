package service

const paginationDefaultLimit = 20

func paginate[T any](items []T, limit, offset int) []T {
	total := len(items)

	if limit == 0 {
		limit = paginationDefaultLimit
	}

	if offset >= total {
		return []T{}
	}

	end := offset + limit

	return items[offset:min(end, total)]
}
