package model

import (
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("user not found")
var ErrNoRights = errors.New("insufficient rights to perform the action")

type Employee struct {
	ID              uuid.UUID
	Username        string
	OrganizationIDs []uuid.UUID
}
