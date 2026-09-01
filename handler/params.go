package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gitea.kood.tech/nikolailetshev/movies-api/models"
)

func pathID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		return 0, models.ErrNotFound
	}
	return id, nil
}

func queryInt(r *http.Request, key string) (*int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: query parameter '%s' must be a whole number", models.ErrInvalidInput, key)
	}
	return &value, nil
}

func pageParams(r *http.Request) (page, size int, err error) {
	p, err := queryInt(r, "page")
	if err != nil {
		return 0, 0, err
	}
	s, err := queryInt(r, "size")
	if err != nil {
		return 0, 0, err
	}

	if p != nil {
		page = *p
	}
	if s != nil {
		size = *s
	}
	return page, size, nil
}

func movieFilter(r *http.Request) (models.MovieFilter, error) {
	var filter models.MovieFilter
	var err error

	if filter.GenreID, err = queryInt(r, "genre"); err != nil {
		return filter, err
	}
	if filter.ActorID, err = queryInt(r, "actor"); err != nil {
		return filter, err
	}
	if filter.Year, err = queryInt(r, "year"); err != nil {
		return filter, err
	}
	filter.Title = strings.TrimSpace(r.URL.Query().Get("title"))

	return filter, nil
}
