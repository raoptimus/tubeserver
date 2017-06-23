package data

import "fmt"

type (
	Text struct {
		Quote    string   `bson:"Quote"`
		Language Language `bson:"Language"`
	}
	TextList []*Text
)

func (s TextList) Get(lang Language, allowEmpty bool) string {
	// see https://workflowboard.com/jira/browse/MOB-449
	switch lang {
	case LanguageRussian, LanguageEnglish:
		// allowed
	default:
		lang = LanguageEnglish
	}

	def := ""
	for _, txt := range s {
		if txt.Language == lang {
			return txt.Quote
		}
		if txt.Language == LanguageEnglish || def == "" {
			def = txt.Quote
		}
	}

	if !allowEmpty {
		return def
	}
	return ""
}

func (t TextList) String() string {
	s := ""
	for _, text := range t {
		s += fmt.Sprintf("[%s]%q", text.Language, text.Quote)
	}
	return s
}
