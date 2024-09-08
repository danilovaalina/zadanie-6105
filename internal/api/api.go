package api

import (
	"net/http"
	"os"

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
	v := os.Getenv("POSTGRES_CONN")

	return c.String(http.StatusOK, "pong "+v)
}
