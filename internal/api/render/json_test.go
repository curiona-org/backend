package render_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/curiona-org/backend/internal/api/render"
	"github.com/stretchr/testify/assert"
)

func TestRenderer_OK(t *testing.T) {
	t.Parallel()
	renderer := render.New(context.Background())
	w := httptest.NewRecorder()
	msg := "Success"
	data := map[string]string{"key": "value"}

	renderer.OK(w, msg, data)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	expected := `{"success":true,"message":"Success","data":{"key":"value"}}`
	assert.JSONEq(t, expected, w.Body.String())
}

func TestRenderer_Created(t *testing.T) {
	t.Parallel()
	renderer := render.New(context.Background())
	w := httptest.NewRecorder()
	msg := "Resource created"
	data := map[string]string{"key": "value"}

	renderer.Created(w, msg, data)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	expected := `{"success":true,"message":"Resource created","data":{"key":"value"}}`
	assert.JSONEq(t, expected, w.Body.String())
}

func TestRenderer_Error(t *testing.T) {
	t.Parallel()
	renderer := render.New(context.Background())
	w := httptest.NewRecorder()
	msg := "An error occurred"
	err := "error details"

	renderer.Error(w, http.StatusBadRequest, msg, err, "")

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))

	expected := `{"success":false,"message":"An error occurred","error":"error details"}`
	assert.JSONEq(t, expected, w.Body.String())
}

func TestRenderer_JSON(t *testing.T) {
	t.Parallel()
	renderer := render.New(context.Background())
	w := httptest.NewRecorder()
	data := render.Response{
		Success: true,
		Message: "Test message",
		Data:    map[string]string{"key": "value"},
	}

	renderer.JSON(w, http.StatusOK, data)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	expected := `{"success":true,"message":"Test message","data":{"key":"value"}}`
	assert.JSONEq(t, expected, w.Body.String())
}

func TestRenderer_JSON_EmptyResponse(t *testing.T) {
	t.Parallel()
	renderer := render.New(context.Background())
	w := httptest.NewRecorder()

	renderer.JSON(w, http.StatusOK, render.Response{})

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	expected := `{"success":true,"message":"OK"}`
	assert.JSONEq(t, expected, w.Body.String())
}

func TestRenderer_JSON_EmptyResponseError(t *testing.T) {
	t.Parallel()
	renderer := render.New(context.Background())
	w := httptest.NewRecorder()

	renderer.JSON(w, http.StatusInternalServerError, render.Response{})

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	expected := `{"success":false,"message":"Internal Server Error"}`
	assert.JSONEq(t, expected, w.Body.String())
}

func TestRenderer_JSON_ErrorEncoding(t *testing.T) {
	t.Parallel()
	renderer := render.New(context.Background())
	w := httptest.NewRecorder()

	// Use a channel to cause json.Marshal to fail
	data := render.Response{
		Success: true,
		Message: "Test message",
		Data:    make(chan int),
	}

	renderer.JSON(w, http.StatusOK, data)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	expected := `{"success":false,"message":"Oops! We encountered an unexpected error. Please try again."}`
	assert.JSONEq(t, expected, w.Body.String())
}
