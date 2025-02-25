package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/curiona-org/backend/internal/admin"
	"github.com/curiona-org/backend/internal/api/render"
	"github.com/curiona-org/backend/internal/app"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/chat"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/pkg/validation"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-playground/validator/v10"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

type API struct {
	server    *http.Server
	router    chi.Router
	render    *render.Renderer
	validator *validation.CustomValidator

	application app.CurionaApplication
	adminApp    admin.Application
	chatApp     chat.Application
}

func New(ctx context.Context, port string, curionaApp app.CurionaApplication, adminApp admin.Application, chatApp chat.Application) *API {
	router := chi.NewRouter()
	api := &API{
		router:      router,
		render:      render.New(ctx),
		validator:   validation.New(),
		application: curionaApp,
		adminApp:    adminApp,
		chatApp:     chatApp,
	}

	api.SetupMiddlewares()
	api.SetupRoutes()
	api.server = &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Addr:              ":" + port,
		Handler:           router,
	}

	return api
}

func (a *API) Start(ctx context.Context) {
	log := logger.FromContext(ctx)

	shutdownChan := make(chan error, 1)
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		log.Warn().Msg("shutting down api server...")
		shutdownChan <- a.server.Shutdown(shutdownCtx)
	}()

	log.Info().Msgf("Listening on %s", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil {
		log.Warn().Msg(err.Error())
	}
}

func (a *API) SetupRoutes() {
	a.router.Get("/health", a.HealthCheck)

	a.router.Post("/auth", a.Auth)
	a.router.Post("/auth/refresh", a.AuthRefresh)

	authMiddleware := a.authMiddleware(a.application)

	// authenticated routes
	a.router.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/profile", a.GetProfile)
		r.Patch("/profile", a.UpdateProfile)

		r.Get("/roadmaps", a.ListUserRoadmaps)
		r.Get("/roadmaps/{slug}", a.GetRoadmapBySlug)
		r.Post("/roadmaps", a.GenerateRoadmap)
		r.Delete("/roadmaps/{slug}", a.DeleteUserRoadmap)

		r.Get("/roadmaps/topic/{slug}", a.GetTopicBySlug)
		r.Patch("/roadmaps/topic/{slug}/finish", a.MarkTopicAsFinished)
		r.Patch("/roadmaps/topic/{slug}/incomplete", a.MarkTopicAsIncomplete)
	})
}

func (a *API) SetupMiddlewares() {
	a.router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete,
			http.MethodHead, http.MethodOptions},
	}))
	a.router.Use(middleware.RealIP)
	a.router.Use(a.requestIDMiddleware)
	a.router.Use(a.populateLog)
	a.router.Use(a.loggerMiddleware)
	a.router.Use(middleware.Recoverer)
}

type Middleware func(next http.Handler) http.Handler

func (a *API) authMiddleware(app app.CurionaApplication) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context()

			authorization := r.Header.Get("Authorization")
			if authorization == "" {
				a.handleError(w, r, cerrors.ErrUnauthorized)
				return
			}

			bearer := strings.Split(authorization, " ")
			if len(bearer) < 2 {
				a.handleError(w, r, cerrors.ErrUnauthorized)
				return
			}

			t := bearer[1]
			token, err := app.AuthVerify(reqCtx, t)
			if err != nil {
				a.handleError(w, r, cerrors.ErrUnauthorized)
				return
			}

			ctx := context.WithValue(reqCtx, auth.ContextKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

const requestIDHeader = "X-Request-Id"

type requestIDContextKey struct{}

func (a *API) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = xid.New().String()
		}

		ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *API) populateLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.Get()
		ctx := log.WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *API) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)

		requestID, ok := ctx.Value(requestIDContextKey{}).(string)
		if ok {
			log.UpdateContext(func(logC zerolog.Context) zerolog.Context {
				return logC.Str("request_id", requestID)
			})
		}

		start := time.Now()
		lrw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(lrw, r)

		status := lrw.Status()
		requestUri, err := url.QueryUnescape(r.RequestURI)
		if err != nil {
			requestUri = fmt.Sprintf("%s (URL decode failed: %v)", r.RequestURI, err)
		}
		if status >= http.StatusInternalServerError {
			log.Error().
				Str("uri", requestUri).
				Int("status", status).
				Str("remote_addr", r.RemoteAddr).
				Int("bytes", lrw.BytesWritten()).
				Str("method", r.Method).
				Str("user_agent", r.UserAgent()).
				Dur("elapsed_ms", time.Since(start)).
				Msg("Incoming request")
		} else {
			log.Info().
				Str("uri", requestUri).
				Int("status", status).
				Str("remote_addr", r.RemoteAddr).
				Int("bytes", lrw.BytesWritten()).
				Str("method", r.Method).
				Str("user_agent", r.UserAgent()).
				Dur("elapsed_ms", time.Since(start)).
				Msg("Incoming request")
		}
	})
}

func (a *API) Bind(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

func (a *API) Param(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func (a *API) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *cerrors.AppError
	var msg string
	code := http.StatusInternalServerError
	if errors.As(err, &appErr) {
		code = appErr.Code()
		msg = appErr.Message()
	}

	var validationErrMsgs []validationErrMsg
	if validationErrs, isValidationErr := err.(validator.ValidationErrors); isValidationErr {
		code = http.StatusUnprocessableEntity

		validationErrMsgs = make([]validationErrMsg, 0)
		for _, err := range validationErrs {
			validationErrMsgs = append(validationErrMsgs, getValidationErrMsg(err))
		}
	}

	if len(validationErrMsgs) > 0 {
		a.render.Error(w, code, "Validation failed.", validationErrMsgs)
		return
	}

	if msg == "" {
		msg = cerrors.DefaultErrorMessage
	}

	a.render.Error(w, code, msg, nil)
}

type validationErrMsg struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func getValidationErrMsg(err validator.FieldError) validationErrMsg {
	errMsg := validationErrMsg{
		Field: err.Field(),
	}

	errMsg.Message = map[string]string{
		"required":         err.Field() + " is required.",
		"required_without": err.Field() + " is required.",
		"email":            "Must be a valid email address.",
		"min":              err.Field() + " must be at least " + err.Param() + " characters long.",
		"max":              err.Field() + " must not exceed " + err.Param() + " characters.",
		"url":              "Must be a valid URL.",
		"oneof":            err.Field() + " must be one of the following: " + err.Param() + ".",
	}[err.Tag()]

	return errMsg
}
