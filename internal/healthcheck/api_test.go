package healthcheck

import (
	"net/http"
	"testing"

	"github.com/berkaykrc/homerun-ratings-system/internal/test"
	"github.com/berkaykrc/homerun-ratings-system/pkg/log"
)

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	RegisterHandlers(router, "0.9.0")
	test.Endpoint(t, router, test.APITestCase{
		Name:         "ok",
		Method:       "GET",
		URL:          "/healthcheck",
		Body:         "",
		Header:       nil,
		WantStatus:   http.StatusOK,
		WantResponse: `"OK 0.9.0"`,
	})
}
