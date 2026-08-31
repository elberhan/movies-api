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

func (s *MovieService) GetByID(id int) (*models.Movie, error) {
	if id < 1 {
		return nil, models.ErrNotFound
	}
	return s.repo.GetByID(id)
}

// GetDetail returns the movie details.
func (s *MovieService) GetDetail(id int) (*models.MovieDetail, error) {
	if id < 1 {
		return nil, models.ErrNotFound
	}
	return s.detail(id)
}

func (s *MovieService) GetAll(page, size int) (*models.Page, error) {
	page, size = normalizePagination(page, size)
	movies, total, err := s.repo.GetAll(size, page*size)
	if err != nil {
		return nil, err
	}
	return newPage(movies, page, size, total), nil
}

// GetByGenre answers GET /api/movies?genre={id}. An unknown genre gives an
// empty page, because a filter that matches nothing is not a missing resource.
func (s *MovieService) GetByGenre(genreID, page, size int) (*models.Page, error) {
	page, size = normalizePagination(page, size)
	movies, total, err := s.repo.GetByGenre(genreID, size, page*size)
	if err != nil {
		return nil, err
	}
	return newPage(movies, page, size, total), nil
}

func (s *MovieService) GetByYear(year, page, size int) (*models.Page, error) {
	page, size = normalizePagination(page, size)
	movies, total, err := s.repo.GetByYear(year, size, page*size)
	if err != nil {
		return nil, err
	}
	return newPage(movies, page, size, total), nil
}

func (s *MovieService) GetByActor(actorID, page, size int) (*models.Page, error) {
	page, size = normalizePagination(page, size)
	movies, total, err := s.repo.GetByActor(actorID, size, page*size)
	if err != nil {
		return nil, err
	}
	return newPage(movies, page, size, total), nil
}

// SearchByTitle answers GET /api/movies/search?title=. The match is partial
// and case-insensitive.
func (s *MovieService) SearchByTitle(query string, page, size int) (*models.Page, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("%w: query parameter 'title' is required", models.ErrInvalidInput)
	}

	page, size = normalizePagination(page, size)
	movies, total, err := s.repo.SearchByTitle(query, size, page*size)
	if err != nil {
		return nil, err
	}
	return newPage(movies, page, size, total), nil
}

// ActorsInMovie answers GET /api/movies/{id}/actors.
func (s *MovieService) ActorsInMovie(id, page, size int) (*models.Page, error) {
	if _, err := s.GetByID(id); err != nil {
		return nil, err
	}

	page, size = normalizePagination(page, size)
	actors, total, err := s.repo.GetActorsByMovie(id, size, page*size)
	if err != nil {
		return nil, err
	}
	return newPage(actors, page, size, total), nil
}

// GenresInMovie answers GET /api/movies/{id}/genres.
func (s *MovieService) GenresInMovie(id, page, size int) (*models.Page, error) {
	if _, err := s.GetByID(id); err != nil {
		return nil, err
	}

	page, size = normalizePagination(page, size)
	genres, total, err := s.repo.GetGenresByMovie(id, size, page*size)
	if err != nil {
		return nil, err
	}
	return newPage(genres, page, size, total), nil
}

// UpdatePartial applies a PATCH: only provided fields are changed.
func (s *MovieService) UpdatePartial(id int, title *string, releaseYear, duration *int, genreIDs, actorIDs []int) (*models.MovieDetail, error) {
	movie, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	updated := *movie // the copy keeps the fields the request does not carry
	if title != nil {
		if updated.Title, err = validateTitle(*title); err != nil {
			return nil, err
		}
	}
	if releaseYear != nil {
		if err := validateReleaseYear(*releaseYear); err != nil {
			return nil, err
		}
		updated.ReleaseYear = *releaseYear
	}
	if duration != nil {
		if err := validateDuration(*duration); err != nil {
			return nil, err
		}
		updated.Duration = *duration
	}

	// reject an empty list here, so that a bad request never opens a transaction
	if genreIDs != nil && len(genreIDs) == 0 {
		return nil, fmt.Errorf("%w: field 'genres' needs a minimum of one genre id", models.ErrInvalidInput)
	}
	if actorIDs != nil && len(actorIDs) == 0 {
		return nil, fmt.Errorf("%w: field 'actors' needs a minimum of one actor id", models.ErrInvalidInput)
	}

	if _, err := s.repo.Update(id, &updated, genreIDs, actorIDs); err != nil {
		return nil, err
	}

	return s.detail(id)
}

// Delete removes a movie. A movie with genres or actors is removed only with
// force, which drops the links first.
func (s *MovieService) Delete(id int, force bool) error {
	movie, err := s.GetByID(id)
	if err != nil {
		return err
	}

	genres, actors, err := s.relationCounts(id)
	if err != nil {
		return err
	}

	if genres+actors > 0 && !force {
		return fmt.Errorf("%w: cannot delete movie '%s' because it has %d genres and %d actors associated",
			models.ErrHasRelations, movie.Title, genres, actors)
	}

	if force {
		return s.repo.ForceDelete(id)
	}
	return s.repo.Delete(id)
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

// relationCounts reports how many genres and actors the movie links to. The
// repository gives a total only next to a page of rows, so it asks for one row.
func (s *MovieService) relationCounts(id int) (genres, actors int, err error) {
	if _, genres, err = s.repo.GetGenresByMovie(id, 1, 0); err != nil {
		return 0, 0, err
	}
	if _, actors, err = s.repo.GetActorsByMovie(id, 1, 0); err != nil {
		return 0, 0, err
	}
	return genres, actors, nil
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

// newPage wraps raw repository data into the uniform models.Page JSON struct
func newPage[T any](items []*T, page, size, total int) *models.Page {
	return &models.Page{
		Content:       toAnySlice(items),
		Page:          page,
		Size:          size,
		TotalElements: total,
		TotalPages:    totalPages(total, size),
	}
}
