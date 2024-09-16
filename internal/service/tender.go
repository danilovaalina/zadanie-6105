package service

import (
	"context"
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"

	"zadanie-6105/internal/model"
)

func (s *Service) Tenders(ctx context.Context, username string, opts model.TenderFilter) ([]model.Tender, error) {
	employee, err := s.repository.Employee(ctx, username)
	if err != nil {
		return nil, err
	}

	if opts.My {
		opts.CreatorID = employee.ID
	}

	opts.OrganizationIDs = employee.OrganizationIDs

	tenders, err := s.repository.Tenders(ctx, opts)
	if err != nil {
		return nil, err
	}

	return tenders, nil
}

func (s *Service) Tender(ctx context.Context, username string, opts model.TenderFilter) (model.Tender, error) {
	tenders, err := s.Tenders(ctx, username, opts)
	if err != nil {
		return model.Tender{}, err
	}

	if len(tenders) == 0 {
		return model.Tender{}, model.ErrTenderOrVersionNotFound
	}

	return tenders[0], nil
}

func (s *Service) CreateTender(ctx context.Context, username string, tender model.Tender) (model.Tender, error) {
	employee, err := s.repository.Employee(ctx, username)
	if err != nil {
		return model.Tender{}, err
	}

	if !slices.Contains(employee.OrganizationIDs, tender.OrganizationID) {
		return model.Tender{}, errors.New("insufficient rights to perform the action")
	}

	tender.ID, err = uuid.NewV7()
	if err != nil {
		return model.Tender{}, errors.WithStack(err)
	}

	tender.Status = model.TenderStatusCreated
	tender.CreatorID = employee.ID

	t, err := s.repository.CreateTender(ctx, tender)
	if err != nil {
		return model.Tender{}, err
	}

	return t, nil
}

func (s *Service) UpdateTender(ctx context.Context, username string, tender model.Tender) (model.Tender, error) {
	employee, err := s.repository.Employee(ctx, username)
	if err != nil {
		return model.Tender{}, err
	}

	opts := model.TenderFilter{
		TenderID: tender.ID,
	}

	_, err = s.Tender(ctx, username, opts)
	if err != nil {
		return model.Tender{}, err
	}

	tender.CreatorID = employee.ID

	t, err := s.repository.UpdateTender(ctx, tender)
	if err != nil {
		return model.Tender{}, err
	}

	return t, nil
}

func (s *Service) RollbackTender(ctx context.Context, username string, tenderID uuid.UUID, versionID int64) (model.Tender, error) {
	employee, err := s.repository.Employee(ctx, username)
	if err != nil {
		return model.Tender{}, err
	}

	opts := model.TenderFilter{
		TenderID:  tenderID,
		VersionID: versionID,
	}

	_, err = s.Tender(ctx, username, opts)
	if err != nil {
		return model.Tender{}, err
	}

	t, err := s.repository.RollbackTender(ctx, tenderID, versionID, employee.ID)
	if err != nil {
		return model.Tender{}, err
	}

	return t, nil
}
