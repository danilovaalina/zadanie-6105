package model

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
)

type BidStatus string

var ErrTenderOrBidNotFound = errors.New("tender or bid not found")

const (
	BidStatusCreated   BidStatus = "Created"
	BidStatusPublished BidStatus = "Published"
	BidStatusClosed    BidStatus = "Closed"
	BidStatusApproved  BidStatus = "Approved"
	BidStatusRejected  BidStatus = "Rejected"
)

type CreatorType string

const (
	CreatorTypeOrganization CreatorType = "Organization"
	CreatorTypeUser         CreatorType = "User"
)

type BidFilter struct {
	My              bool
	BidID           uuid.UUID
	CreatorID       uuid.UUID
	Status          []BidStatus
	TenderID        uuid.UUID
	OrganizationIDs []uuid.UUID
	Offset          uint64
	Limit           uint64
}

type Bid struct {
	ID             uuid.UUID
	Name           string
	Description    string
	Status         BidStatus
	TenderID       uuid.UUID
	CreatorType    CreatorType
	CreatorID      uuid.UUID
	OrganizationID uuid.UUID
	VersionID      int64
	Created        time.Time
}
