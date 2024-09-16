package repository

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"zadanie-6105/internal/model"
)

const minQuorum = 3

func (r *Repository) Bids(ctx context.Context, opts model.BidFilter) ([]model.Bid, error) {
	b := r.builder.
		Select("b.id",
			"b.name",
			"b.description",
			"b.status",
			"b.tender_id",
			"b.creator_type",
			"b.creator_id",
			"b.organization_id",
			"b.version_id",
			"b.created",
		).From("bid b").Join("tender t on b.tender_id = t.id").Where(sq.Or{
		sq.Eq{"b.organization_id": opts.OrganizationIDs},
		sq.And{
			sq.NotEq{"b.status": model.BidStatusCreated},
			sq.Eq{"t.organization_id": opts.OrganizationIDs},
		},
	})

	if opts.BidID != uuid.Nil {
		b = b.Where(sq.Eq{"b.id": opts.BidID})
	}

	if opts.TenderID != uuid.Nil {
		b = b.Where(sq.Eq{"b.tender_id": opts.TenderID})
	}

	if opts.CreatorID != uuid.Nil {
		b = b.Where(sq.Eq{"b.creator_id": opts.CreatorID})
	}

	if len(opts.Status) > 0 {
		b = b.Where(sq.Eq{"b.status": opts.Status})
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

	bidRows, err := pgx.CollectRows[bidRow](rows, pgx.RowToStructByNameLax[bidRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(model.ErrTenderOrBidNotFound)
		}
		return nil, errors.WithStack(err)
	}

	bids := make([]model.Bid, 0, len(bidRows))
	for _, row := range bidRows {
		bids = append(bids, r.bidModel(row))
	}

	return bids, nil
}

func (r *Repository) CreateBid(ctx context.Context, bid model.Bid) (model.Bid, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
	insert into bid (id, name, description, status, tender_id, creator_type, creator_id, organization_id, version_id, created)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	returning id, name, description, status, tender_id, creator_type, creator_id, organization_id, version_id, created`

	rows, err := tx.Query(ctx, query,
		bid.ID, bid.Name, bid.Description, bid.Status, bid.TenderID, bid.CreatorType, bid.CreatorID, bid.OrganizationID, 1, time.Now(),
	)
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}

	row, err := pgx.CollectExactlyOneRow[bidRow](rows, pgx.RowToStructByNameLax[bidRow])
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}

	b := r.bidModel(row)

	err = r.saveBidVersion(ctx, tx, b)
	if err != nil {
		return model.Bid{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}

	return b, nil
}

func (r *Repository) UpdateBid(ctx context.Context, bid model.Bid) (model.Bid, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	b := r.builder.Update("bid").
		PrefixExpr(
			sq.StatementBuilder.
				Select("id").
				From("bid").
				Where(sq.And{
					sq.Eq{"id": bid.ID},
					sq.Eq{"creator_id": bid.CreatorID},
				}).
				Prefix("with t as (").Suffix("for update)"),
		).
		Set("version_id", sq.Expr("version_id + 1"))

	if bid.Name != "" {
		b = b.Set("name", bid.Name)
	}

	if bid.Description != "" {
		b = b.Set("description", bid.Description)
	}

	if bid.Status != "" {
		b = b.Set("status", bid.Status)
	}

	b = b.Where(sq.And{
		sq.Eq{"id": bid.ID},
		sq.Eq{"creator_id": bid.CreatorID},
	}).Suffix("returning id, name, description, status, tender_id, creator_type, creator_id, organization_id, version_id, created")

	query, args, err := b.ToSql()
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}

	row, err := pgx.CollectExactlyOneRow[bidRow](rows, pgx.RowToStructByNameLax[bidRow])
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}

	bb := r.bidModel(row)

	err = r.saveBidVersion(ctx, tx, bb)
	if err != nil {
		return model.Bid{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}

	return bb, nil
}

func (r *Repository) RollbackBid(ctx context.Context, bidID uuid.UUID, versionID int64, creatorID uuid.UUID) (model.Bid, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
	with u as (select id from tender where id = $1 and creator_id = $3 for update),
	     v as (select name, description, status
	           from bid_version
	           where bid_id = $1
	             and id = $2)
	update bid b
	set name        = v.name,
	    description = v.description,
	    status      = v.status,
	    version_id  = version_id + 1
	from v
	where id = $1 and creator_id = $3
	returning b.id, b.name, b.description, b.status, b.organization_id, b.creator_id, b.version_id, b.created`

	rows, err := tx.Query(ctx, query, bidID, versionID, creatorID)
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}

	row, err := pgx.CollectExactlyOneRow[bidRow](rows, pgx.RowToStructByNameLax[bidRow])
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}

	b := r.bidModel(row)

	err = r.saveBidVersion(ctx, tx, b)
	if err != nil {
		return model.Bid{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}

	return b, nil
}

func (r *Repository) SubmitBidDecision(ctx context.Context, bidID uuid.UUID, employee model.Employee, status model.BidStatus) (model.Bid, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
	insert
	into bid_agreement (bid_id, employee_id, status)
	select $1, $2, $3
	from bid b
	         join tender t on b.tender_id = t.id
	where b.id = $1
	  and b.status = 'Published'
	  and t.status = 'Published'
 	  and t.organization_id = any($4)
	    for update
	returning bid_id`

	var id uuid.UUID
	err = tx.QueryRow(ctx, query, bidID, employee.ID, status, employee.OrganizationIDs).Scan(&id)
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}

	query = `
	with q as (select least(count(o.employee_id), $2) count
	           from bid b
	                    join tender t on b.tender_id = t.id
	                    join organization_employee o on t.organization_id = o.organization_id
	           where b.id = $1)
	select status from bid_agreement
	where bid_id = $1 and status = 'Approved'
	group by status having count(bid_id) >= (select count from q)
	union
	select status from bid_agreement
	where bid_id = $1 and status = 'Rejected'
	group by status having count(bid_id) >= 1`

	var bidStatus model.BidStatus

	err = tx.QueryRow(ctx, query, bidID, minQuorum).Scan(&bidStatus)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.Bid{}, errors.WithStack(err)
	}

	var b model.Bid

	if bidStatus == "" {
		query = `
		select id, name, description, status, tender_id, creator_type, creator_id, organization_id, version_id, created
		from bid where id = $1`

		rows, err := tx.Query(ctx, query, bidID, bidStatus)
		if err != nil {
			return model.Bid{}, errors.WithStack(err)
		}

		row, err := pgx.CollectExactlyOneRow[bidRow](rows, pgx.RowToStructByNameLax[bidRow])
		if err != nil {
			return model.Bid{}, errors.WithStack(err)
		}

		b = r.bidModel(row)
	} else {
		query = `update bid set status = $2, version_id = version_id + 1 where id = $1
		returning id, name, description, status, tender_id, creator_type, creator_id, organization_id, version_id, created`

		rows, err := tx.Query(ctx, query, bidID, bidStatus)
		if err != nil {
			return model.Bid{}, errors.WithStack(err)
		}

		row, err := pgx.CollectExactlyOneRow[bidRow](rows, pgx.RowToStructByNameLax[bidRow])
		if err != nil {
			return model.Bid{}, errors.WithStack(err)
		}

		b := r.bidModel(row)

		err = r.saveBidVersion(ctx, tx, b)
		if err != nil {
			return model.Bid{}, err
		}

		if bidStatus == model.BidStatusApproved {
			query = `update tender set status = $2, version_id = version_id + 1 where id = $1
			returning id, name, description, status, service_type, organization_id, creator_id, version_id, created`

			rows, err := tx.Query(ctx, query, b.TenderID, model.TenderStatusClosed)
			if err != nil {
				return model.Bid{}, errors.WithStack(err)
			}

			row, err := pgx.CollectExactlyOneRow[tenderRow](rows, pgx.RowToStructByNameLax[tenderRow])
			if err != nil {
				return model.Bid{}, errors.WithStack(err)
			}

			tender := r.tenderModel(row)

			err = r.saveTenderVersion(ctx, tx, tender)
			if err != nil {
				return model.Bid{}, err
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return model.Bid{}, errors.WithStack(err)
	}

	return b, nil
}

func (r *Repository) saveBidVersion(ctx context.Context, tx pgx.Tx, bid model.Bid) error {
	query := `
	insert into bid_version (id, bid_id, name, description, status, created) 
	values ($1, $2, $3, $4, $5, $6)`

	_, err := tx.Exec(ctx, query, bid.VersionID, bid.ID, bid.Name, bid.Description, bid.Status, time.Now())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Repository) bidModel(row bidRow) model.Bid {
	return model.Bid{
		ID:             row.ID,
		Name:           row.Name,
		Description:    row.Description,
		Status:         model.BidStatus(row.Status),
		TenderID:       row.TenderID,
		CreatorType:    model.CreatorType(row.CreatorType),
		CreatorID:      row.CreatorID,
		OrganizationID: row.OrganizationID,
		VersionID:      row.VersionID,
		Created:        row.Created,
	}
}

type bidRow struct {
	ID             uuid.UUID `db:"id"`
	Name           string    `db:"name"`
	Description    string    `db:"description"`
	Status         string    `db:"status"`
	TenderID       uuid.UUID `db:"tender_id"`
	CreatorType    string    `db:"creator_type"`
	CreatorID      uuid.UUID `db:"creator_id"`
	OrganizationID uuid.UUID `db:"organization_id"`
	VersionID      int64     `db:"version_id"`
	Created        time.Time `db:"created"`
}
