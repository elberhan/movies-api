package service

import (
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
