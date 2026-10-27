FROM golang:1.26.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway ./services/gateway

FROM alpine:3.22

WORKDIR /app
COPY --from=builder /gateway .

EXPOSE 8082

CMD ["./gateway"]
