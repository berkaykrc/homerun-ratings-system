package test

import (
	"net/http"
	"net/http/httptest"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/internal/errors"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/accesslog"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/go-ozzo/ozzo-routing/v2/content"
)

// MockRoutingContext creates a routing.Context for testing handlers.
func MockRoutingContext(req *http.Request) (*routing.Context, *httptest.ResponseRecorder) {
	res := httptest.NewRecorder()
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	ctx := routing.NewContext(res, req)
	ctx.SetDataWriter(&content.JSONDataWriter{})
	return ctx, res
}

// MockRouter creates a routing.Router for testing APIs.
func MockRouter(logger log.Logger) *routing.Router {
	router := routing.New()
	router.Use(
		accesslog.Handler(logger),
		errors.Handler(logger),
		content.TypeNegotiator(content.JSON),
	)
	return router
}
