package rating

import (
	"net/http"

	"github.com/berkaykrc/homerun-ratings-system/internal/errors"
	"github.com/berkaykrc/homerun-ratings-system/pkg/log"
	routing "github.com/go-ozzo/ozzo-routing/v2"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *routing.RouteGroup, service Service, logger log.Logger) {
	res := resource{service, logger}

	r.Get("/ratings/<id>", res.get)
	r.Post("/ratings", res.create)
}

type resource struct {
	service Service
	logger  log.Logger
}

func (r resource) get(c *routing.Context) error {
	rating, err := r.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		return err
	}

	return c.Write(rating)
}

func (r resource) create(c *routing.Context) error {
	var input CreateRatingRequest
	if err := c.Read(&input); err != nil {
		r.logger.With(c.Request.Context()).Info(err)
		return errors.BadRequest("")
	}
	rating, err := r.service.Create(c.Request.Context(), input)
	if err != nil {
		return err
	}

	return c.WriteWithStatus(rating, http.StatusCreated)
}
