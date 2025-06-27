package accesslog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://127.0.0.1/healthcheck", nil)
	ctx := routing.NewContext(res, req)

	logger, entries := log.NewForTest()
	handler := Handler(logger)
	err := handler(ctx)

	assert.Nil(t, err)
	assert.Equal(t, 1, entries.Len())
	assert.Equal(t, "GET /healthcheck HTTP/1.1 200 0", entries.All()[0].Message)
}
