package service

import (
	"fmt"
	"strings"

	"gitea.kood.tech/nikolailetshev/movies-api/models"
	"gitea.kood.tech/nikolailetshev/movies-api/repository"
)

const (
	defaultPageSize = 10
	maxPageSize     = 100
	maxNameLength   = 100
)

type GenreService struct {
	repo *repository.GenreRepository
}

func NewGenreService(repo *repository.GenreRepository) *GenreService {
	return &GenreService{repo: repo}
}

func (s *GenreService) Create(name string) (*models.Genre, error) {
	name, err := validateName(name)
	if err != nil {
		return nil, err
	}
	return s.repo.Create(name)
}

func (s *GenreService) GetByID(id int) (*models.Genre, error) {
	if id < 1 {
		return nil, models.ErrNotFound
	}
	return s.repo.GetByID(id)
}

func (s *GenreService) GetAll(page, size int) (*models.Page, error) {
	page, size = normalizePagination(page, size)
	genres, total, err := s.repo.GetAll(size, page*size)
	if err != nil {
		return nil, err
	}

	return &models.Page{
		Content:       toAnySlice(genres),
		Page:          page,
		Size:          size,
		TotalElements: total,
		TotalPages:    totalPages(total, size),
	}, nil
}

// Update renames a genre after validating the new name
func (s *GenreService) Update(id int, name string) (*models.Genre, error) {
	if id < 1 {
		return nil, models.ErrNotFound
	}
	name, err := validateName(name)
	if err != nil {
		return nil, err
	}
	return s.repo.Update(id, name)
}

//helpers

func validateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("%w: field 'name' is required", models.ErrInvalidInput)
	}
	if len(name) > maxNameLength {
		return "", fmt.Errorf("%w: field 'name' must be at most %d characters long", models.ErrInvalidInput, maxNameLength)
	}
	return name, nil
}

func normalizePagination(page, size int) (int, int) {
	if page < 0 {
		page = 0
	}
	switch {
	case size < 1:
		size = defaultPageSize
	case size > maxPageSize:
		size = maxPageSize
	}
	return page, size
}

func totalPages(total, size int) int {
	if total == 0 {
		return 0
	}
	return (total + size - 1) / size
}

func toAnySlice[T any](items []*T) []any {
	out := make([]any, len(items))
	for i, item := range items {
		out[i] = item
	}
	return out
}
