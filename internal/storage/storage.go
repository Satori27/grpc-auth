package storage

import "errors"


var (
	ErrUserExists = errors.New("user already exists")
	ErrUserNotFound = errors.New("user does not exists")
	ErrRoleIsEmpty = errors.New("empty role")
	ErrAppNotFound = errors.New("app not found")
	ErrAppNameExists = errors.New("app name already exists")
)