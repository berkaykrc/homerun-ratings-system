package serviceprovider

import (
	"net/http"

	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/errors"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/log"
	routing "github.com/go-ozzo/ozzo-routing/v2"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *routing.RouteGroup, service Service, logger log.Logger) {
	res := resource{service, logger}

	r.Get("/service-providers/<id>", res.get)
	r.Post("/service-providers", res.create)
}

type resource struct {
	service Service
	logger  log.Logger
}

func (r resource) get(c *routing.Context) error {
	serviceProvider, err := r.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		return err
	}

	return c.Write(serviceProvider)
}

func (r resource) create(c *routing.Context) error {
	var input CreateServiceProviderRequest
	if err := c.Read(&input); err != nil {
		r.logger.With(c.Request.Context()).Info(err)
		return errors.BadRequest("")
	}
	serviceProvider, err := r.service.Create(c.Request.Context(), input)
	if err != nil {
		return err
	}

	return c.WriteWithStatus(serviceProvider, http.StatusCreated)
}
