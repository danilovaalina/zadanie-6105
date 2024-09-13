package repository

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"zadanie-6105/internal/model"
)

func (r *Repository) Employee(ctx context.Context, username string) (model.Employee, error) {
	query := `
	select e.id,
	       e.username,
	       array_agg(o.organization_id) organizations
	from employee e
	         left join organization_employee o on e.id = o.employee_id
	where username = $1
	group by e.id, e.username`

	rows, err := r.pool.Query(ctx, query, username)
	if err != nil {
		return model.Employee{}, errors.WithStack(err)
	}

	row, err := pgx.CollectExactlyOneRow[employeeRow](rows, pgx.RowToStructByNameLax[employeeRow])
	if err != nil {
		return model.Employee{}, errors.WithStack(err)
	}

	return r.employeeModel(row), nil
}

func (r *Repository) employeeModel(row employeeRow) model.Employee {
	return model.Employee{
		ID:              row.ID,
		Username:        row.Username,
		OrganizationIDs: row.Organizations,
	}
}

type employeeRow struct {
	ID            uuid.UUID   `db:"id"`
	Username      string      `db:"username"`
	Organizations []uuid.UUID `db:"organizations"`
}
