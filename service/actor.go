package service

import "gitea.kood.tech/nikolailetshev/movies-api/models"

type ActorService interface {
	Create(name, birthDate string) (*models.Actor, error)
	GetByID(id int) (*models.Actor, error)
	GetAll(page, size int) (*models.Page, error)
	GetByName(name string, page, size int) (*models.Page, error)
	Update(id int, name, birthDate string) (*models.Actor, error)
	Delete(id int, force bool) error
}
