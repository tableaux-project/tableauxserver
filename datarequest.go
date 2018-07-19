package tableauxserver

import "github.com/tableaux-project/tableaux"

// DataRequest represents the DTO of a single data request from the frontend.
type DataRequest struct {
	Start  int64  `json:"start"`
	Draw   int64  `json:"draw"`
	Length int64  `json:"length"`
	Locale string `json:"locale"`

	Columns []Column      `json:"columns"`
	Order   []ColumnOrder `json:"order"`
	Search  GlobalSearch  `json:"search"`
}

// Column represents the DTO of a single column that is requested from the frontend.
// Optionally, ColumnSearches might be added to the column, suggesting that this
// column should be filtered. If multiple ColumnSearches are given, they are
// expected to be OR-chained.
type Column struct {
	Name   string         `json:"name"`
	Search []ColumnSearch `json:"search"`
}

// ColumnOrder represents the DTO of a request to order a single column either
// ascending or descending.
type ColumnOrder struct {
	Column string         `json:"column"`
	Dir    tableaux.Order `json:"dir"`
}

// ColumnSearch represents the DTO of a request to search for a value in a specified
// mode. The DTO itself is not very useful, and should only be used in conjunction
// with a Column, which then describes the request to search for a value inside that
// Column.
type ColumnSearch struct {
	Value string              `json:"value"`
	Mode  tableaux.FilterMode `json:"mode"`
}

// GlobalSearch represents the DTO for a global search over all columns.
type GlobalSearch struct {
	Value string `json:"value"`
}
