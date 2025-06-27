package healthcheck

import (
	"testing"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/internal/test"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
)

func TestHealthcheck(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	RegisterHandlers(router, "1.0.0")

	test.Endpoint(t, router, test.APITestCase{
		Name:         "healthcheck",
		Method:       "GET",
		URL:          "/healthcheck",
		WantStatus:   200,
		WantResponse: "*OK 1.0.0*",
	})
}
