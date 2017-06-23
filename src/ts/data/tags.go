package data

type (
	Tags struct {
		Language Language `bson:"Language"`
		Tags     []string `bson:"Tags"`
	}
	TagsList []*Tags
)

func (s *TagsList) Get(lang Language) []string {
	for _, t := range *s {
		if t.Language == lang {
			return t.Tags
		}
	}

	return []string{}
}
