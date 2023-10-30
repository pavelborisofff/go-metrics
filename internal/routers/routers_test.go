package routers

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInitRouter(t *testing.T) {
	r := InitRouter()

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
}
