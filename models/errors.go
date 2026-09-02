package models

import "errors"

var ErrNotFound = errors.New("resource not found")
var ErrInvalidInput = errors.New("invalid input")
var ErrHasRelations = errors.New("resource has associated records")
var ErrConflict = errors.New("resource already exists")
