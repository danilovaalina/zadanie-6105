package api

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"zadanie-6105/internal/model"
)

type tenderBidsRequest struct {
	TenderID uuid.UUID `param:"tenderId"`
	Username string    `query:"username"`
	Limit    uint64    `query:"limit"`
	Offset   uint64    `query:"offset"`
}

func (a *API) tenderBids(c echo.Context) error {
	var req tenderBidsRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	opts := model.BidFilter{
		TenderID: req.TenderID,
		Offset:   req.Offset,
		Limit:    req.Limit,
	}

	bids, err := a.service.Bids(c.Request().Context(), req.Username, opts)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, a.bidsFromModel(bids))
}

func (a *API) myBids(c echo.Context) error {
	var req tenderBidsRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	opts := model.BidFilter{
		My:     true,
		Offset: req.Offset,
		Limit:  req.Limit,
	}

	bids, err := a.service.Bids(c.Request().Context(), req.Username, opts)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, a.bidsFromModel(bids))
}

type bidStatusRequest struct {
	BidID    uuid.UUID `param:"bidId"`
	Username string    `query:"username"`
}

func (a *API) bidStatus(c echo.Context) error {
	var req bidStatusRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	bid, err := a.service.Bid(c.Request().Context(), req.Username, req.BidID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.String(http.StatusOK, string(bid.Status))
}

type createBidRequest struct {
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	TenderID       uuid.UUID `json:"tenderId"`
	CreatorType    string    `json:"creatorType"`
	OrganizationID uuid.UUID `json:"organizationId"`
	Username       string    `json:"creatorUsername"`
}

func (a *API) createBid(c echo.Context) error {
	var req createBidRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	bid := model.Bid{
		Name:           req.Name,
		Description:    req.Description,
		TenderID:       req.TenderID,
		CreatorType:    model.CreatorType(req.CreatorType),
		OrganizationID: req.OrganizationID,
	}

	b, err := a.service.CreateBid(c.Request().Context(), req.Username, bid)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, a.bidFromModel(b))
}

type updateBidRequest struct {
	BidID       uuid.UUID `param:"bidId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

func (a *API) updateBid(c echo.Context) error {
	var req updateBidRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	bid := model.Bid{
		ID:          req.BidID,
		Name:        req.Name,
		Description: req.Description,
	}

	t, err := a.service.UpdateBid(c.Request().Context(), c.QueryParam("username"), bid)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, a.bidFromModel(t))
}

type updateBidStatusRequest struct {
	BidID uuid.UUID `param:"bidId"`
}

func (a *API) updateBidStatus(c echo.Context) error {
	var req updateBidStatusRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	bid := model.Bid{
		ID:     req.BidID,
		Status: model.BidStatus(c.QueryParam("status")),
	}

	b, err := a.service.UpdateBid(c.Request().Context(), c.QueryParam("username"), bid)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, a.bidFromModel(b))
}

type rollbackBidRequest struct {
	BidID     uuid.UUID `param:"bidId"`
	VersionID int64     `param:"version"`
}

func (a *API) rollbackBid(c echo.Context) error {
	var req rollbackBidRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	b, err := a.service.RollbackBid(c.Request().Context(), c.QueryParam("username"), req.BidID, req.VersionID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, a.bidFromModel(b))
}

type submitBidDecisionRequest struct {
	BidID uuid.UUID `param:"bidId"`
}

func (a *API) submitBidDecision(c echo.Context) error {
	var req submitBidDecisionRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	b, err := a.service.SubmitBidDecision(c.Request().Context(), c.QueryParam("username"), req.BidID,
		model.BidStatus(c.QueryParam("decision")))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, a.bidFromModel(b))
}

type bidsResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	CreatorType string    `json:"authorType"`
	CreatorID   uuid.UUID `json:"authorId"`
	VersionID   int64     `json:"version"`
	Created     time.Time `json:"createdAt"`
}

func (a *API) bidsFromModel(bids []model.Bid) []bidsResponse {
	var r = make([]bidsResponse, 0, len(bids))
	for _, b := range bids {
		r = append(r, a.bidFromModel(b))
	}

	return r
}

func (a *API) bidFromModel(bid model.Bid) bidsResponse {
	return bidsResponse{
		ID:          bid.ID,
		Name:        bid.Name,
		Status:      string(bid.Status),
		CreatorType: string(bid.CreatorType),
		CreatorID:   bid.CreatorID,
		VersionID:   bid.VersionID,
		Created:     bid.Created,
	}
}
