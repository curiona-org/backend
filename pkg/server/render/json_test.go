package render_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/pkg/server/render"
	"github.com/stretchr/testify/assert"
)

func TestRender_OK(t *testing.T) {
	t.Parallel()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	message := "Operation successful"
	data := map[string]string{"key": "value"}

	err := render.OK(c, message, data)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"success":true,"message":"Operation successful","data":{"key":"value"}}`, rec.Body.String())
}

func TestRender_Created(t *testing.T) {
	t.Parallel()
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	message := "Resource created"
	data := map[string]string{"id": "123"}

	err := render.Created(c, message, data)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.JSONEq(t, `{"success":true,"message":"Resource created","data":{"id":"123"}}`, rec.Body.String())
}

func TestRender_Error(t *testing.T) {
	t.Parallel()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	code := http.StatusBadRequest
	message := "Invalid request"
	errData := map[string]string{"error": "invalid_input"}

	err := render.Error(c, code, message, errData)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"success":false,"message":"Invalid request","error":{"error":"invalid_input"}}`, rec.Body.String())
}
