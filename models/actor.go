package models

// TODO structs + json tags
type Actor struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	BirthDate string `json:"birthDate"`
}
