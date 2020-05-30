package search

import "strings"

type Document struct {
	Terms []string
	ID string
	Row int
}

func NewDocument(terms []string, id string, row int) *Document {
	return &Document{
		Terms: terms,
		ID:    id,
		Row:   row,
	}
}

func NewDocumentFromString(body string, id string, row int, f SplitFunc) *Document {
	return NewDocument(f(body), id, row)
}

type SplitFunc func(body string) (terms []string)

func SplitAtSpace(body string) (terms []string) {
	return strings.Split(body, " ")
}