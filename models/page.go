package models

type Page struct {
	Content       []any `json:"content"`
	Page          int   `json:"page"`
	Size          int   `json:"size"`
	TotalElements int   `json:"totalElements"`
	TotalPages    int   `json:"totalPages"`
}
