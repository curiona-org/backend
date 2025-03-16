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
	Code    string `json:"code,omitempty"`
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

func (r *Renderer) Error(w http.ResponseWriter, statusCode int, msg string, err any, code string) {
	r.JSON(w, statusCode, Response{
		Success: false,
		Message: msg,
		Code:    code,
		Error:   err,
	})
}

func (r *Renderer) JSON(w http.ResponseWriter, statusCode int, data Response) {
	b := bytebufferpool.Get()
	b.Reset()
	defer bytebufferpool.Put(b)

	if data == (Response{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
			fmt.Fprint(w, jsonOK)
			return
		}

		msg := escapeJSON(http.StatusText(statusCode))
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
	w.WriteHeader(statusCode)
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
