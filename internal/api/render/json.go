package render

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/valyala/bytebufferpool"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func (r *Renderer) OK(w http.ResponseWriter, msg string, data any) {
	r.JSON(w, http.StatusOK, Response{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

func (r *Renderer) Created(w http.ResponseWriter, msg string, data any) {
	r.JSON(w, http.StatusCreated, Response{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

func (r *Renderer) Error(w http.ResponseWriter, code int, msg string, err any) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	r.JSON(w, code, Response{
		Success: false,
		Message: msg,
		Error:   err,
	})
}

func (r *Renderer) JSON(w http.ResponseWriter, code int, data Response) {
	b := bytebufferpool.Get()
	b.Reset()
	defer bytebufferpool.Put(b)

	if data == (Response{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)

		if code >= http.StatusOK && code < http.StatusMultipleChoices {
			fmt.Fprint(w, jsonOK)
			return
		}

		msg := escapeJSON(http.StatusText(code))
		fmt.Fprintf(w, jsonErrFormat, msg)
		return
	}

	if err := json.NewEncoder(b).Encode(data); err != nil {
		r.logger.Error().Err(err).Msg("failed to marshal json")
		msg := escapeJSON(cerrors.DefaultErrorMessage)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, jsonErrFormat, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err := b.WriteTo(w); err != nil {
		r.logger.Error().Err(err).Msg("failed to write json to response")
	}
}

const jsonOK = `{"success":true,"message":"OK"}`
const jsonErrFormat = `{"success":false,"message":"%s"}`

// escapeJSON does primitive JSON escaping.
func escapeJSON(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}
