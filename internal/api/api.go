package api

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"zadanie-6105/internal/model"
)

type Service interface {
	Tenders(ctx context.Context, username string, opts model.TenderFilter) ([]model.Tender, error)
	Tender(ctx context.Context, username string, tenderID uuid.UUID) (model.Tender, error)
	CreateTender(ctx context.Context, username string, tender model.Tender) (model.Tender, error)
	UpdateTender(ctx context.Context, username string, tender model.Tender) (model.Tender, error)
	RollbackTender(ctx context.Context, username string, tenderID uuid.UUID, versionID int64) (model.Tender, error)
	Bids(ctx context.Context, username string, opts model.BidFilter) ([]model.Bid, error)
	Bid(ctx context.Context, username string, bidID uuid.UUID) (model.Bid, error)
	CreateBid(ctx context.Context, username string, bid model.Bid) (model.Bid, error)
	UpdateBid(ctx context.Context, username string, bid model.Bid) (model.Bid, error)
	RollbackBid(ctx context.Context, username string, tenderID uuid.UUID, versionID int64) (model.Bid, error)
	SubmitBidDecision(ctx context.Context, username string, bidID uuid.UUID, status model.BidStatus) (model.Bid, error)
}

type API struct {
	*echo.Echo
	service Service
}

func New(service Service) *API {
	a := &API{
		Echo:    echo.New(),
		service: service,
	}

	api := a.Group("/api")
	{
		api.GET("/ping", a.ping)

		tenders := api.Group("/tenders")
		{
			tenders.GET("", a.tenders)
			tenders.GET("/my", a.myTenders)
			tenders.POST("/new", a.createTender)

			tender := tenders.Group("/:tenderId")
			{
				tender.GET("/status", a.tenderStatus)
				tender.PATCH("/edit", a.updateTender)
				tender.PUT("/status", a.updateTenderStatus)
				tender.PUT("/rollback/:version", a.rollbackTender)
			}
		}

		bids := api.Group("/bids")
		{
			bids.GET("/my", a.myBids)
			bids.GET("/:tenderId/list", a.tenderBids)
			bids.POST("/new", a.createBid)

			bid := bids.Group("/:bidId")
			{
				bid.GET("/status", a.bidStatus)
				bid.PATCH("/edit", a.updateBid)
				bid.PUT("/status", a.updateBidStatus)
				bid.PUT("/rollback/:version", a.rollbackBid)
				bid.PUT("/submit_decision", a.submitBidDecision)
			}
		}
	}

	return a
}

func (a *API) ping(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
