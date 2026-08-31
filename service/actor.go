package service

import (
	"fmt"
	"strings"
	"time"

	"gitea.kood.tech/nikolailetshev/movies-api/models"
	"gitea.kood.tech/nikolailetshev/movies-api/repository"
)

type ActorService struct {
	actorRepo *repository.ActorRepository
}

func NewActorService(repo *repository.ActorRepository) *ActorService {
	return &ActorService{actorRepo: repo}
}

func (s *ActorService) Create(name, birthDate string) (*models.Actor, error) {
	name, err := validateName(name)
	if err != nil {
		return nil, err
	}
	if err := validateBirthDate(birthDate); err != nil {
		return nil, err
	}
	return s.actorRepo.Create(name, birthDate)
}

func (s *ActorService) GetByID(id int) (*models.Actor, error) {
	if id < 1 {
		return nil, models.ErrNotFound
	}
	return s.actorRepo.GetByID(id)
}

func (s *ActorService) GetAll(page, size int) (*models.Page, error) {
	page, size, err := validatePagination(page, size)
	if err != nil {
		return nil, err
	}

	actors, total, err := s.actorRepo.GetAll(size, page*size)
	if err != nil {
		return nil, err
	}

	return &models.Page{
		Content:       toAnySlice(actors),
		Page:          page,
		Size:          size,
		TotalElements: total,
		TotalPages:    totalPages(total, size),
	}, nil
}

func (s *ActorService) GetByName(name string, page, size int) (*models.Page, error) {
	page, size, err := validatePagination(page, size)
	if err != nil {
		return nil, err
	}

	actors, err := s.actorRepo.GetByName(name, size, page*size)
	if err != nil {
		return nil, err
	}

	return &models.Page{
		Content:       toAnySlice(actors),
		Page:          page,
		Size:          size,
		TotalElements: len(actors),
		TotalPages:    totalPages(len(actors), size),
	}, nil
}

// UpdatePartial applies a PATCH: only provided fields are changed.
// name and birthDate are pointers so an absent field is distinguishable
// from an empty one.
func (s *ActorService) UpdatePartial(id int, name, birthDate *string) (*models.Actor, error) {
	actor, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// keep existing values for fields not present in the request
	newName := actor.Name
	if name != nil {
		newName, err = validateName(*name)
		if err != nil {
			return nil, err
		}
	}

	newBirthDate := actor.BirthDate
	if birthDate != nil {
		if err := validateBirthDate(*birthDate); err != nil {
			return nil, err
		}
		newBirthDate = *birthDate
	}

	return s.actorRepo.Update(id, newName, newBirthDate)
}

func (s *ActorService) Delete(id int, force bool) error {
	actor, err := s.GetByID(id)
	if err != nil {
		return err
	}

	count, err := s.actorRepo.MovieCount(id)
	if err != nil {
		return err
	}

	if count > 0 && !force {
		return fmt.Errorf("%w: cannot delete actor '%s' because they are associated with %d movies",
			models.ErrHasRelations, actor.Name, count)
	}

	if force && count > 0 {
		if err := s.actorRepo.DeleteActorMovieLinks(id); err != nil {
			return err
		}
	}

	return s.actorRepo.Delete(id)
}

func validateBirthDate(birthDate string) error {
	birthDate = strings.TrimSpace(birthDate)
	if _, err := time.Parse("2006-01-02", birthDate); err != nil {
		return fmt.Errorf("%w: field 'birthDate' must be a valid date in YYYY-MM-DD format", models.ErrInvalidInput)
	}
	return nil
}
