package model

import (
	"github.com/google/uuid"
)

type Employee struct {
	ID              uuid.UUID
	Username        string
	OrganizationIDs []uuid.UUID
}
