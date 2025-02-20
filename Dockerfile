ARG GO_VERSION=1.23.4
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download -x

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -v -a -installsuffix cgo -o /bin/server ./cmd/application/main.go

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -v -a -installsuffix cgo -o /bin/migrate ./cmd/migrate/main.go

FROM alpine:3.17.2 AS final

RUN apk --update add \
    ca-certificates \
    tzdata \
    && \
    update-ca-certificates

# Create a non-privileged user that the app will run under.
# See https://docs.docker.com/go/dockerfile-user-best-practices/
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser

COPY --from=builder /bin/server /bin/
COPY --from=builder ./.env /bin/
COPY --from=builder /bin/migrate /bin/

EXPOSE 8080

# Run migrations
CMD ["/bin/migrate"]

# Run the service
ENTRYPOINT ["/bin/server"]