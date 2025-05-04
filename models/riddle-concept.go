package models

type RiddleConcept struct {
	ThemeDescription string   `json:"themeDescription"`
	SuperSolution    string   `json:"superSolution"`
	WordPool         []string `json:"wordPool"`
}
