package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type API struct {
	*echo.Echo
}

func New() *API {
	a := &API{
		Echo: echo.New(),
	}

	a.Echo.GET("/api/ping", a.ping)

	return a
}

func (a *API) ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}
