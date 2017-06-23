package v1

type (
	SearchQuery struct {
		Text        string
		ResultCount int
	}
	SearchQueryList []*SearchQuery
)
