package models

// Request estructura principal del modelo
type ZincRequest struct {
	From int         `json:"from"`
	Size int         `json:"size"`
	Sort []SortField `json:"sort"`
}
type ZincBodyRequest struct {
	From  int              `json:"from"`
	Size  int              `json:"size"`
	Sort  []SortField      `json:"sort"`
	Query MatchPhraseQuery `json:"query"`
}

// SortField estructura para la ordenación
type SortField struct {
	Id Order `json:"id"`
}

// Order define el orden de clasificación
type Order struct {
	Order string `json:"order"`
}

type MatchPhraseQuery struct {
	MatchPhrase struct {
		Body string `json:"body"`
	} `json:"match_phrase"`
}
