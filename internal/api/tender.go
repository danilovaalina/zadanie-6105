package api

import (
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"zadanie-6105/internal/model"
)

type tendersRequest struct {
	Username    string `query:"username"`
	ServiceType string `query:"serviceType"`
	Limit       uint64 `query:"limit"`
	Offset      uint64 `query:"offset"`
}

func (a *API) tenders(c echo.Context) error {
	var req tendersRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"reason": "invalid request format or params"})
	}

	opts := model.TenderFilter{
		ServiceType: model.ServiceType(req.ServiceType),
		Offset:      req.Offset,
		Limit:       req.Limit,
	}

	tenders, err := a.service.Tenders(c.Request().Context(), req.Username, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"reason": err.Error()})
	}

	return c.JSON(http.StatusOK, a.tendersFromModel(tenders))
}

func (a *API) myTenders(c echo.Context) error {
	var req tendersRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"reason": "invalid request format or params"})
	}

	opts := model.TenderFilter{
		My:          true,
		ServiceType: model.ServiceType(req.ServiceType),
		Offset:      req.Offset,
		Limit:       req.Limit,
	}

	tenders, err := a.service.Tenders(c.Request().Context(), req.Username, opts)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"reason": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"reason": err.Error()})
	}

	return c.JSON(http.StatusOK, a.tendersFromModel(tenders))
}

type tenderStatusRequest struct {
	Username string    `query:"username"`
	TenderID uuid.UUID `param:"tenderId"`
}

func (a *API) tenderStatus(c echo.Context) error {
	var req tenderStatusRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"reason": "invalid request format or params"})
	}

	tender, err := a.service.Tender(c.Request().Context(), req.Username, model.TenderFilter{TenderID: req.TenderID})
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"reason": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"reason": err.Error()})
	}

	return c.String(http.StatusOK, string(tender.Status))
}

type createTenderRequest struct {
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ServiceType    string    `json:"serviceType"`
	OrganizationID uuid.UUID `json:"organizationId"`
	Username       string    `json:"creatorUsername"`
}

func (a *API) createTender(c echo.Context) error {
	var req createTenderRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"reason": "invalid request format or params"})
	}

	tender := model.Tender{
		Name:           req.Name,
		Description:    req.Description,
		ServiceType:    model.ServiceType(req.ServiceType),
		OrganizationID: req.OrganizationID,
	}

	t, err := a.service.CreateTender(c.Request().Context(), req.Username, tender)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"reason": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"reason": err.Error()})
	}

	return c.JSON(http.StatusOK, a.tenderFromModel(t))
}

type updateTenderRequest struct {
	TenderID    uuid.UUID `param:"tenderId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ServiceType string    `json:"serviceType"`
}

func (a *API) updateTender(c echo.Context) error {
	var req updateTenderRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"reason": "invalid request format or params"})
	}

	tender := model.Tender{
		ID:          req.TenderID,
		Name:        c.QueryParam("username"),
		Description: req.Description,
		ServiceType: model.ServiceType(req.ServiceType),
	}

	t, err := a.service.UpdateTender(c.Request().Context(), c.QueryParam("username"), tender)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"reason": err.Error()})
		}
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, a.tenderFromModel(t))
}

type updateTenderStatusRequest struct {
	TenderID uuid.UUID `param:"tenderId"`
}

func (a *API) updateTenderStatus(c echo.Context) error {
	var req updateTenderStatusRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"reason": "invalid request format or params"})
	}

	tender := model.Tender{
		ID:     req.TenderID,
		Status: model.TenderStatus(c.QueryParam("status")),
	}

	t, err := a.service.UpdateTender(c.Request().Context(), c.QueryParam("username"), tender)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"reason": err.Error()})
		}
		if errors.Is(err, model.ErrTenderOrVersionNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"reason": err.Error()})
		}
		if errors.Is(err, model.ErrCreatorNotFound) {
			return c.JSON(http.StatusForbidden, echo.Map{"reason": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"reason": err.Error()})
	}

	return c.JSON(http.StatusOK, a.tenderFromModel(t))
}

type rollbackTenderRequest struct {
	TenderID  uuid.UUID `param:"tenderId"`
	VersionID int64     `param:"version"`
}

func (a *API) rollbackTender(c echo.Context) error {
	var req rollbackTenderRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"reason": "invalid request format or params"})
	}

	t, err := a.service.RollbackTender(c.Request().Context(), c.QueryParam("username"), req.TenderID, req.VersionID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"reason": err.Error()})
		}
		if errors.Is(err, model.ErrTenderOrVersionNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"reason": err.Error()})
		}
		if errors.Is(err, model.ErrCreatorNotFound) {
			return c.JSON(http.StatusForbidden, echo.Map{"reason": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"reason": err.Error()})
	}

	return c.JSON(http.StatusOK, a.tenderFromModel(t))
}

type tenderResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	ServiceType string    `json:"serviceType"`
	Version     int64     `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (a *API) tendersFromModel(tenders []model.Tender) []tenderResponse {
	var r = make([]tenderResponse, 0, len(tenders))
	for _, t := range tenders {
		r = append(r, a.tenderFromModel(t))
	}

	return r
}

func (a *API) tenderFromModel(tender model.Tender) tenderResponse {
	return tenderResponse{
		ID:          tender.ID,
		Name:        tender.Name,
		Description: tender.Description,
		Status:      string(tender.Status),
		ServiceType: string(tender.ServiceType),
		Version:     tender.VersionID,
		CreatedAt:   tender.Created,
	}
}
