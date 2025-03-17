ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -v -a -installsuffix cgo -o /bin/server ./cmd/application/main.go

FROM gcr.io/distroless/static-debian12

COPY --from=builder /bin/server /bin/

EXPOSE 8080

ENTRYPOINT ["/bin/server"]