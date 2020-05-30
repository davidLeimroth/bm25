package model

type Document struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Body []string `json:"body"`
	Meta struct {
		DocumentType string `json:"ch:documentType"`
	} `json:"meta"`
	Ctr int
}
