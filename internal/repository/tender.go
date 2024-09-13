package repository

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"zadanie-6105/internal/model"
)

func (r *Repository) Tenders(ctx context.Context, opts model.TenderFilter) ([]model.Tender, error) {
	b := r.builder.
		Select("id",
			"name",
			"description",
			"status",
			"service_type",
			"organization_id",
			"creator_id",
			"version_id",
			"created",
		).From("tender").Where(sq.Or{
		sq.Eq{"organization_id": opts.OrganizationIDs},
		sq.Eq{"status": model.TenderStatusPublished},
	})

	if opts.TenderID != uuid.Nil {
		b = b.Where(sq.Eq{"id": opts.TenderID})
	}

	if opts.CreatorID != uuid.Nil {
		b = b.Where(sq.Eq{"creator_id": opts.CreatorID})
	}

	if opts.ServiceType != "" {
		b = b.Where(sq.Eq{"service_type": opts.ServiceType})
	}

	if len(opts.Status) > 0 {
		b = b.Where(sq.Eq{"status": opts.Status})
	}

	if opts.Offset > 0 {
		b = b.Offset(opts.Offset)
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = defaultLimit
	}

	b = b.Limit(limit)

	query, args, err := b.ToSql()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	tenderRows, err := pgx.CollectRows[tenderRow](rows, pgx.RowToStructByNameLax[tenderRow])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	tenders := make([]model.Tender, 0, len(tenderRows))
	for _, row := range tenderRows {
		tenders = append(tenders, r.tenderModel(row))
	}

	return tenders, nil
}

func (r *Repository) CreateTender(ctx context.Context, tender model.Tender) (model.Tender, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Tender{}, errors.WithStack(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
	insert into tender (id, name, description, status, service_type, organization_id, creator_id, version_id, created)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	returning id, name, description, status, service_type, organization_id, creator_id, version_id, created`

	rows, err := tx.Query(ctx, query, tender.ID, tender.Name, tender.Description, tender.Status, tender.ServiceType,
		tender.OrganizationID, tender.CreatorID, 1, time.Now())
	if err != nil {
		return model.Tender{}, errors.WithStack(err)
	}

	row, err := pgx.CollectExactlyOneRow[tenderRow](rows, pgx.RowToStructByNameLax[tenderRow])
	if err != nil {
		return model.Tender{}, errors.WithStack(err)
	}

	t := r.tenderModel(row)

	err = r.saveTenderVersion(ctx, tx, t)
	if err != nil {
		return model.Tender{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return model.Tender{}, errors.WithStack(err)
	}

	return t, nil
}

func (r *Repository) UpdateTender(ctx context.Context, tender model.Tender) (model.Tender, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Tender{}, errors.WithStack(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	b := r.builder.Update("tender").
		PrefixExpr(
			sq.StatementBuilder.
				Select("id").
				From("tender").
				Where(sq.And{
					sq.Eq{"id": tender.ID},
					sq.Eq{"creator_id": tender.CreatorID},
				}).Prefix("with t as (").Suffix("for update)"),
		).
		Set("version_id", sq.Expr("version_id + 1"))

	if tender.Name != "" {
		b = b.Set("name", tender.Name)
	}

	if tender.Description != "" {
		b = b.Set("description", tender.Description)
	}

	if tender.Status != "" {
		b = b.Set("status", tender.Status)
	}

	if tender.ServiceType != "" {
		b = b.Set("service_type", tender.ServiceType)
	}

	b = b.Where(sq.And{
		sq.Eq{"id": tender.ID},
		sq.Eq{"creator_id": tender.CreatorID},
	}).Suffix("returning id, name, description, status, service_type, organization_id, creator_id, version_id, created")

	query, args, err := b.ToSql()
	if err != nil {
		return model.Tender{}, errors.WithStack(err)
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return model.Tender{}, errors.WithStack(err)
	}

	row, err := pgx.CollectExactlyOneRow[tenderRow](rows, pgx.RowToStructByNameLax[tenderRow])
	if err != nil {
		return model.Tender{}, errors.WithStack(err)
	}

	t := r.tenderModel(row)

	err = r.saveTenderVersion(ctx, tx, t)
	if err != nil {
		return model.Tender{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return model.Tender{}, errors.WithStack(err)
	}

	return t, nil
}

func (r *Repository) RollbackTender(ctx context.Context, tenderID uuid.UUID, versionID int64, creatorID uuid.UUID) (model.Tender, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Tender{}, errors.WithStack(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
	with u as (select id from tender where id = $1 and creator_id = $3 for update),
	     v as (select name, description, status, service_type
	           from tender_version
	           where tender_id = $1
	             and id = $2)
	update tender t
	set name         = v.name,
	    description  = v.description,
	    status       = v.status,
	    service_type = v.service_type,
	    version_id   = version_id + 1
	from v
	where id = $1 and creator_id = $3
	returning t.id, t.name, t.description, t.status, t.service_type, t.organization_id, t.creator_id, t.version_id, t.created`

	rows, err := tx.Query(ctx, query, tenderID, versionID, creatorID)
	if err != nil {
		return model.Tender{}, errors.WithStack(err)
	}

	row, err := pgx.CollectExactlyOneRow[tenderRow](rows, pgx.RowToStructByNameLax[tenderRow])
	if err != nil {
		return model.Tender{}, errors.WithStack(err)
	}

	t := r.tenderModel(row)

	err = r.saveTenderVersion(ctx, tx, t)
	if err != nil {
		return model.Tender{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return model.Tender{}, errors.WithStack(err)
	}

	return t, nil
}

func (r *Repository) saveTenderVersion(ctx context.Context, tx pgx.Tx, tender model.Tender) error {
	query := `
	insert into tender_version (id, tender_id, name, description, status, service_type, created) 
	values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := tx.Exec(ctx, query,
		tender.VersionID, tender.ID, tender.Name, tender.Description, tender.Status, tender.ServiceType, time.Now(),
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Repository) tenderModel(row tenderRow) model.Tender {
	return model.Tender{
		ID:             row.ID,
		Name:           row.Name,
		Description:    row.Description,
		ServiceType:    model.ServiceType(row.ServiceType),
		Status:         model.TenderStatus(row.Status),
		OrganizationID: row.OrganizationID,
		CreatorID:      row.CreatorID,
		VersionID:      row.VersionID,
		Created:        row.Created,
	}
}

type tenderRow struct {
	ID             uuid.UUID `db:"id"`
	Name           string    `db:"name"`
	Description    string    `db:"description"`
	Status         string    `db:"status"`
	ServiceType    string    `db:"service_type"`
	OrganizationID uuid.UUID `db:"organization_id"`
	CreatorID      uuid.UUID `db:"creator_id"`
	VersionID      int64     `db:"version_id"`
	Created        time.Time `db:"created"`
}
