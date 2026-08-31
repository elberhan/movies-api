package service

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"gitea.kood.tech/nikolailetshev/movies-api/models"
	"gitea.kood.tech/nikolailetshev/movies-api/repository"
)

const (
	maxTitleLength = 200
	minReleaseYear = 1888
	maxReleaseYear = 2100

	// SQLite reads a negative LIMIT as "no limit": a detail view shows every link
	allRows = -1
)

type MovieService struct {
	repo *repository.MovieRepository
}

func NewMovieService(repo *repository.MovieRepository) *MovieService {
	return &MovieService{repo: repo}
}

// Create creates a new movie in the DB, linked together with its genre and actor.
func (s *MovieService) Create(title string, releaseYear, duration int, genreIDs, actorIDs []int) (*models.MovieDetail, error) {
	title, err := validateTitle(title)
	if err != nil {
		return nil, err
	}
	if err := validateReleaseYear(releaseYear); err != nil {
		return nil, err
	}
	if err := validateDuration(duration); err != nil {
		return nil, err
	}

	if len(genreIDs) == 0 {
		return nil, fmt.Errorf("%w: field 'genres' needs a minimum of one genre id", models.ErrInvalidInput)
	}
	if len(actorIDs) == 0 {
		return nil, fmt.Errorf("%w: field 'actors' needs a minimum of one actor id", models.ErrInvalidInput)
	}

	movie, err := s.repo.Create(title, releaseYear, duration)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SetGenres(movie.ID, genreIDs); err != nil {
		return nil, s.discardMovie(movie.ID, err)
	}
	if err := s.repo.SetActors(movie.ID, actorIDs); err != nil {
		return nil, s.discardMovie(movie.ID, err)
	}

	return s.detail(movie.ID)
}

// detail returns movie with its complete genre and actor lists.
func (s *MovieService) detail(id int) (*models.MovieDetail, error) {
	movie, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	genres, _, err := s.repo.GetGenresByMovie(id, allRows, 0)
	if err != nil {
		return nil, err
	}
	actors, _, err := s.repo.GetActorsByMovie(id, allRows, 0)
	if err != nil {
		return nil, err
	}

	// force delete of a genre or an actor can leave a movie with no links,
	// and the answer must then carry [] and not nil
	if genres == nil {
		genres = []*models.Genre{}
	}
	if actors == nil {
		actors = []*models.Actor{}
	}

	return &models.MovieDetail{Movie: *movie, Genres: genres, Actors: actors}, nil
}

// discardMovie removes a movie whose links failed, so that a failed Create
// leaves no row behind. It returns cause, or cause together with the delete error.
func (s *MovieService) discardMovie(movieID int, cause error) error {
	if err := s.repo.ForceDelete(movieID); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

//helpers

func validateTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", fmt.Errorf("%w: field 'title' is required", models.ErrInvalidInput)
	}
	// NUL byte check
	if strings.IndexByte(title, 0) >= 0 {
		return "", fmt.Errorf("%w: field 'title' must not contain a NUL character", models.ErrInvalidInput)
	}
	// length check
	if utf8.RuneCountInString(title) > maxTitleLength {
		return "", fmt.Errorf("%w: field 'title' must be maximum %d characters long", models.ErrInvalidInput, maxTitleLength)
	}
	return title, nil
}

func validateReleaseYear(year int) error {
	if year < minReleaseYear || year > maxReleaseYear {
		return fmt.Errorf("%w: field 'releaseYear' must be between %d and %d",
			models.ErrInvalidInput, minReleaseYear, maxReleaseYear)
	}
	return nil
}

func validateDuration(duration int) error {
	if duration < 1 {
		return fmt.Errorf("%w: field 'duration' must be a positive number of seconds", models.ErrInvalidInput)
	}
	return nil
}

