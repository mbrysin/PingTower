FROM golang:1.26.1-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/stats \
    ./services/stats

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/stats /usr/local/bin/stats

EXPOSE 8082

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/stats"]
