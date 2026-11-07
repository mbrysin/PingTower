FROM golang:1.26.1-alpine AS build

WORKDIR /src

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY services/sites/migrations ./services/sites/migrations
COPY services/checker/migrations ./services/checker/migrations

FROM alpine:3.21

RUN apk add --no-cache ca-certificates
COPY --from=build /go/bin/goose /usr/local/bin/goose

WORKDIR /app
COPY --from=build /src/services/sites/migrations ./services/sites/migrations
COPY --from=build /src/services/checker/migrations ./services/checker/migrations
