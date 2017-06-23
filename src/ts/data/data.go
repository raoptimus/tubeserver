package data

type CommentStatus int

const (
	CommentStatusApproved CommentStatus = iota
	CommentStatusSpam
)

type Language string

const (
	LanguageRussian Language = "ru"
	LanguageEnglish Language = "en"
)
