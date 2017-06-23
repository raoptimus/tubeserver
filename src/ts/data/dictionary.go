package data

type Dictionary struct {
	Id           string  `bson:"_id"` // word
	Translations []*Text `bson:"Translations"`
	Synonyms     []*Text `bson:"Synonyms"`
}
