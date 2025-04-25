package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/curiona-org/backend/internal/admin"
	"github.com/curiona-org/backend/internal/api/render"
	"github.com/curiona-org/backend/internal/api/websocket"
	"github.com/curiona-org/backend/internal/app"
	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/config"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/pkg/validator"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type API struct {
	server *http.Server
	router chi.Router
	ws     *websocket.Manager

	render         *render.Renderer
	validator      validator.Validator
	tracerProvider *tracesdk.TracerProvider

	application app.CurionaApplication
	adminApp    admin.Application
}

func New(ctx context.Context, port string, curionaApp app.CurionaApplication, adminApp admin.Application, tracer *tracesdk.TracerProvider) *API {
	router := chi.NewRouter()
	api := &API{
		router:         router,
		ws:             websocket.NewManager(),
		render:         render.New(ctx),
		validator:      validator.NewPlayground(),
		application:    curionaApp,
		adminApp:       adminApp,
		tracerProvider: tracer,
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

	a.router.Get("/roadmaps", a.ListCommunityRoadmaps)

	// authenticated routes
	a.router.Group(func(r chi.Router) {
		r.Use(a.authMiddleware)

		r.Get("/profile", a.GetProfile)
		r.Patch("/profile", a.UpdateProfile)
		r.Get("/profile/roadmaps", a.ListUserRoadmaps)

		r.Get("/roadmaps/{slug}", a.GetRoadmapBySlug)
		r.Post("/roadmaps", a.GenerateRoadmap)
		r.Delete("/roadmaps/{slug}", a.DeleteUserRoadmap)

		r.Get("/roadmaps/topic/{slug}", a.GetTopicBySlug)
		r.Patch("/roadmaps/topic/{slug}/finish", a.MarkTopicAsFinished)
		r.Patch("/roadmaps/topic/{slug}/incomplete", a.MarkTopicAsIncomplete)
		r.Handle("/roadmaps/{slug}/assist", http.HandlerFunc(a.RoadmapChatAssist))
	})

	// admin routes
	a.router.Group(func(r chi.Router) {
		r.Use(a.authMiddleware)
		r.Use(a.adminMiddleware)

		r.Get("/admin/users", a.AdminListUsers)
		r.Get("/admin/users/{id}", a.AdminGetUser)
		r.Delete("/admin/users/{id}", a.AdminDeleteUser)
		r.Patch("/admin/users/{id}/suspend", a.AdminSuspendUser)
		r.Patch("/admin/users/{id}/unsuspend", a.AdminUnsuspendUser)

		r.Get("/admin/roadmaps", a.AdminListRoadmaps)
		r.Get("/admin/roadmaps/{id}", a.AdminGetRoadmap)
		r.Delete("/admin/roadmaps/{id}", a.AdminDeleteRoadmap)
	})

}

func (a *API) SetupMiddlewares() {
	a.router.Use(a.populateLog)
	a.router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{
			http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost,
			http.MethodDelete, http.MethodHead, http.MethodOptions,
		},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))
	a.router.Use(a.secureHeadersMiddleware)
	a.router.Use(middleware.RealIP)
	a.router.Use(otelhttp.NewMiddleware("api", otelhttp.WithTracerProvider(a.tracerProvider)))
	a.router.Use(a.requestIDMiddleware)
	a.router.Use(a.loggerMiddleware)
	a.router.Use(middleware.Recoverer)
	a.router.Use(middleware.Compress(5))
	if config.IsProduction() {
		a.router.Use(httprate.LimitByIP(100, time.Minute))
	}
}

func (a *API) secureHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// https://developer.mozilla.org/en-US/observatory/docs/faq#can_i_scan_non-websites_such_as_api_endpoints
		// https://infosec.mozilla.org/guidelines/web_security#content-security-policy
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func (a *API) authMiddleware(next http.Handler) http.Handler {
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
		token, err := a.application.AuthVerify(reqCtx, t)
		if err != nil {
			a.handleError(w, r, cerrors.ErrUnauthorized)
			return
		}

		ctx := token.WithContext(reqCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *API) adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)
		auth := auth.FromContext(ctx)

		isAdmin, err := a.adminApp.IsAdmin(ctx, auth.AccountID)
		if err != nil {
			log.Error().Err(err).Msg("failed checking if account is admin or not")
			a.handleError(w, r, cerrors.ErrUnauthorized)
			return
		}

		if !isAdmin {
			a.handleError(w, r, cerrors.ErrForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

const requestIDHeader = "X-Request-Id"

type requestIDContextKey struct{}

func (a *API) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = xid.New().String()
		}
		w.Header().Set(requestIDHeader, requestID)

		span := trace.SpanFromContext(r.Context())
		span.SetAttributes(attribute.String("http.request_id", requestID))

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
			log.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("request_id", requestID)
			})
		}

		start := time.Now()
		lrw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(lrw, r)

		status := lrw.Status()
		requestURI, err := url.QueryUnescape(r.RequestURI)
		if err != nil {
			requestURI = fmt.Sprintf("%s (URL decode failed: %v)", r.RequestURI, err)
		}

		traceCtx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		span := trace.SpanContextFromContext(traceCtx)
		if span.IsValid() {
			log.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("trace_id", span.TraceID().String())
			})
		}

		// TODO: log internal error messages.
		if status >= http.StatusInternalServerError {
			log.Error().
				Str("uri", requestURI).
				Int("status", status).
				Str("remote_addr", r.RemoteAddr).
				Int("bytes", lrw.BytesWritten()).
				Str("method", r.Method).
				Str("origin", r.Header.Get("Origin")).
				Str("user_agent", r.UserAgent()).
				Dur("elapsed_ms", time.Since(start)).
				Msg("Incoming request")
		} else {
			log.Info().
				Str("uri", requestURI).
				Int("status", status).
				Str("remote_addr", r.RemoteAddr).
				Int("bytes", lrw.BytesWritten()).
				Str("method", r.Method).
				Str("origin", r.Header.Get("Origin")).
				Str("user_agent", r.UserAgent()).
				Dur("elapsed_ms", time.Since(start)).
				Msg("Incoming request")
		}
	})
}

func (a *API) handleError(w http.ResponseWriter, r *http.Request, err error) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	var cerr cerrors.CurionaError
	if !errors.As(err, &cerr) {
		cerr = cerrors.New(err)
	}

	// check if the error is a validation error
	validationErrors := a.validator.ParseErrors(err)
	if len(validationErrors) > 0 {
		err := cerrors.ErrValidation
		a.render.Error(w, err.HTTPStatusCode(), err.Message(), validationErrors, err.ErrorCode())
		return
	}

	if cerr.HTTPStatusCode() >= http.StatusInternalServerError && cerr.HTTPStatusCode() < http.StatusInternalServerError+100 {
		log.Error().Ctx(ctx).Err(err).Send()
	}

	a.render.Error(w, cerr.HTTPStatusCode(), cerr.Message(), nil, cerr.ErrorCode())
}

func (a *API) Bind(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

func (a *API) Param(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func (a *API) ParamInt(r *http.Request, key string) (int, error) {
	param := chi.URLParam(r, key)
	return strconv.Atoi(param)
}
