package service

import (
	"context"

	"github.com/google/uuid"

	"zadanie-6105/internal/model"
)

type Repository interface {
	Employee(ctx context.Context, username string) (model.Employee, error)
	Tenders(ctx context.Context, opts model.TenderFilter) ([]model.Tender, error)
	CreateTender(ctx context.Context, tender model.Tender) (model.Tender, error)
	UpdateTender(ctx context.Context, tender model.Tender) (model.Tender, error)
	RollbackTender(ctx context.Context, tenderID uuid.UUID, versionID int64, creatorID uuid.UUID) (model.Tender, error)
	Bids(ctx context.Context, opts model.BidFilter) ([]model.Bid, error)
	CreateBid(ctx context.Context, bid model.Bid) (model.Bid, error)
	UpdateBid(ctx context.Context, bid model.Bid) (model.Bid, error)
	RollbackBid(ctx context.Context, bidID uuid.UUID, versionID int64, creatorID uuid.UUID) (model.Bid, error)
	SubmitBidDecision(ctx context.Context, bidID uuid.UUID, employee model.Employee, status model.BidStatus) (model.Bid, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}
