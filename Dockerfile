# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS base
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
FROM base AS test
COPY . .
CMD ["go","test","-count=1","./..."]
FROM base AS build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/auction-service ./cmd/api
FROM alpine:3.21 AS runtime
RUN apk add --no-cache ca-certificates tzdata && addgroup -S app && adduser -S -G app app
COPY --from=build /out/auction-service /usr/local/bin/auction-service
USER app
EXPOSE 8082
ENTRYPOINT ["/usr/local/bin/auction-service"]
