package model

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
)

var (
	ErrCreatorNotFound         = errors.New("creator of tender not found")
	ErrTenderOrVersionNotFound = errors.New("tender or version not found")
)

type TenderStatus string

const (
	TenderStatusCreated   TenderStatus = "Created"
	TenderStatusPublished TenderStatus = "Published"
	TenderStatusClosed    TenderStatus = "Closed"
)

type ServiceType string

const (
	TenderServiceTypeConstruction ServiceType = "Construction"
	TenderServiceTypeDelivery     ServiceType = "Delivery"
	TenderServiceTypeManufacture  ServiceType = "Manufacture"
)

type TenderFilter struct {
	My              bool
	TenderID        uuid.UUID
	CreatorID       uuid.UUID
	VersionID       int64
	ServiceType     ServiceType
	Status          []TenderStatus
	OrganizationIDs []uuid.UUID
	Offset          uint64
	Limit           uint64
}

type Tender struct {
	ID             uuid.UUID
	Name           string
	Description    string
	ServiceType    ServiceType
	Status         TenderStatus
	OrganizationID uuid.UUID
	CreatorID      uuid.UUID
	VersionID      int64
	Created        time.Time
}
