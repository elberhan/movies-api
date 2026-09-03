package service

import (
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

	movie, err := s.repo.Create(title, releaseYear, duration, genreIDs, actorIDs)
	if err != nil {
		return nil, err
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

// List answers GET /api/movies, narrowed by each filter field that was set. An
// unmatched genre/actor/year filter is a not-found error; an empty catalogue is not.
func (s *MovieService) List(filter models.MovieFilter, page, size int) (*models.Page, error) {
	page, size, err := validatePagination(page, size)
	if err != nil {
		return nil, err
	}

	if err := s.checkFilterTargets(filter); err != nil {
		return nil, err
	}

	movies, total, err := s.repo.Find(filter, size, page*size)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		if err := emptyFilterError(filter); err != nil {
			return nil, err
		}
	}
	return newPage(movies, page, size, total), nil
}

// checkFilterTargets rejects a filter which genre/actor absent in the database.
func (s *MovieService) checkFilterTargets(filter models.MovieFilter) error {
	if filter.GenreID != nil {
		exists, err := s.repo.GenreExists(*filter.GenreID)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("%w: genre with id %d does not exist", models.ErrNotFound, *filter.GenreID)
		}
	}
	if filter.ActorID != nil {
		exists, err := s.repo.ActorExists(*filter.ActorID)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("%w: actor with id %d does not exist", models.ErrNotFound, *filter.ActorID)
		}
	}
	return nil
}

// emptyFilterError reports which filters matched no movie. Returns nil when the request has 
// no genre/actor/year filter, since an empty result is a valid answer, not a missing resource.
func emptyFilterError(filter models.MovieFilter) error {
	var parts []string
	if filter.GenreID != nil {
		parts = append(parts, fmt.Sprintf("genre with id %d", *filter.GenreID))
	}
	if filter.ActorID != nil {
		parts = append(parts, fmt.Sprintf("actor with id %d", *filter.ActorID))
	}
	if filter.Year != nil {
		parts = append(parts, fmt.Sprintf("year %d", *filter.Year))
	}
	if len(parts) == 0 {
		return nil
	}
	return fmt.Errorf("%w: no movies found for %s", models.ErrNotFound, strings.Join(parts, " and "))
}

func (s *MovieService) GetAll(page, size int) (*models.Page, error) {
	return s.List(models.MovieFilter{}, page, size)
}

func (s *MovieService) GetByGenre(genreID, page, size int) (*models.Page, error) {
	return s.List(models.MovieFilter{GenreID: &genreID}, page, size)
}

func (s *MovieService) GetByYear(year, page, size int) (*models.Page, error) {
	return s.List(models.MovieFilter{Year: &year}, page, size)
}

func (s *MovieService) GetByActor(actorID, page, size int) (*models.Page, error) {
	return s.List(models.MovieFilter{ActorID: &actorID}, page, size)
}

// SearchByTitle answers GET /api/movies/search?title=. The match is partial
// and case-insensitive.
func (s *MovieService) SearchByTitle(query string, page, size int) (*models.Page, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("%w: query parameter 'title' is required", models.ErrInvalidInput)
	}
	return s.List(models.MovieFilter{Title: query}, page, size)
}

// ActorsInMovie answers GET /api/movies/{id}/actors.
func (s *MovieService) ActorsInMovie(id, page, size int) (*models.Page, error) {
	if _, err := s.GetByID(id); err != nil {
		return nil, err
	}

	page, size, err := validatePagination(page, size)
	if err != nil {
		return nil, err
	}

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

	page, size, err := validatePagination(page, size)
	if err != nil {
		return nil, err
	}

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

	// genreIDs/actorIDs are nil when the request does not mention them (leave
	// the links untouched); an explicit empty list clears all of that link.
	// No minimum is enforced - a movie may have zero genres or actors.

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
