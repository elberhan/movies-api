package service

import "gitea.kood.tech/nikolailetshev/movies-api/repository"

type MovieService struct {
	repo *repository.MovieRepository
}

func NewMovieService(repo *repository.MovieRepository) *MovieService {
	return &MovieService{repo: repo}
}
