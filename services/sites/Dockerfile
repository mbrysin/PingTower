FROM golang:1.26.1-alpine AS build

WORKDIR /src
COPY . .
RUN go build -o /sites ./services/sites

FROM alpine:3.21

WORKDIR /services/sites
COPY --from=build /sites /usr/local/bin/sites
COPY services/sites/migrations ./migrations

EXPOSE 8081
CMD ["sites"]
