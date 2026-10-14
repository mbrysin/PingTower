FROM golang:1.26.1-alpine AS build

WORKDIR /src
COPY . .
RUN go build -o /auth ./services/auth

FROM alpine:3.21

WORKDIR /services/auth
COPY --from=build /auth /usr/local/bin/auth
COPY services/auth/migrations ./migrations

EXPOSE 8080
CMD ["auth"]
