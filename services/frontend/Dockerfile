FROM golang:1.26.1-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /out/frontend \
    ./services/frontend

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /out/frontend ./frontend
COPY services/frontend/templates ./services/frontend/templates
COPY services/frontend/static ./services/frontend/static

ENV ADDR=:8083
EXPOSE 8083

ENTRYPOINT ["./frontend"]
