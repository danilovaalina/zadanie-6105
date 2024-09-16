package service

import (
	"context"
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"

	"zadanie-6105/internal/model"
)

func (s *Service) Bids(ctx context.Context, username string, opts model.BidFilter) ([]model.Bid, error) {
	employee, err := s.repository.Employee(ctx, username)
	if err != nil {
		return nil, err
	}

	if opts.My {
		opts.CreatorID = employee.ID
	}

	opts.OrganizationIDs = employee.OrganizationIDs

	bids, err := s.repository.Bids(ctx, opts)
	if err != nil {
		return nil, err
	}

	return bids, nil
}

func (s *Service) Bid(ctx context.Context, username string, bidID uuid.UUID) (model.Bid, error) {
	opts := model.BidFilter{
		BidID: bidID,
	}

	bids, err := s.Bids(ctx, username, opts)
	if err != nil {
		return model.Bid{}, err
	}

	if len(bids) == 0 {
		return model.Bid{}, errors.New("TODO")
	}

	return bids[0], nil
}

func (s *Service) CreateBid(ctx context.Context, username string, bid model.Bid) (model.Bid, error) {
	employee, err := s.repository.Employee(ctx, username)
	if err != nil {
		return model.Bid{}, err
	}

	if !slices.Contains(employee.OrganizationIDs, bid.OrganizationID) {
		return model.Bid{}, model.ErrNoRights
	}

	bid.ID, err = uuid.NewV7()
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}

	bid.Status = model.BidStatusCreated
	bid.CreatorID = employee.ID
	if bid.CreatorType == "" {
		bid.CreatorType = model.CreatorTypeUser
	}

	b, err := s.repository.CreateBid(ctx, bid)
	if err != nil {
		return model.Bid{}, err
	}

	return b, nil
}

func (s *Service) UpdateBid(ctx context.Context, username string, bid model.Bid) (model.Bid, error) {
	employee, err := s.repository.Employee(ctx, username)
	if err != nil {
		return model.Bid{}, err
	}

	bid.CreatorID = employee.ID

	b, err := s.repository.UpdateBid(ctx, bid)
	if err != nil {
		return model.Bid{}, err
	}

	return b, nil
}

func (s *Service) RollbackBid(ctx context.Context, username string, tenderID uuid.UUID, versionID int64) (model.Bid, error) {
	employee, err := s.repository.Employee(ctx, username)
	if err != nil {
		return model.Bid{}, err
	}

	b, err := s.repository.RollbackBid(ctx, tenderID, versionID, employee.ID)
	if err != nil {
		return model.Bid{}, err
	}

	return b, nil
}

func (s *Service) SubmitBidDecision(ctx context.Context, username string, bidID uuid.UUID, status model.BidStatus) (model.Bid, error) {
	employee, err := s.repository.Employee(ctx, username)
	if err != nil {
		return model.Bid{}, err
	}

	b, err := s.repository.SubmitBidDecision(ctx, bidID, employee, status)
	if err != nil {
		return model.Bid{}, err
	}

	return b, nil
}
