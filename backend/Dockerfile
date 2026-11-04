FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /api ./cmd/api

FROM alpine:3.20
RUN adduser -D -H app && mkdir /uploads && chown app /uploads
USER app
COPY --from=build /api /api
EXPOSE 8080
CMD ["/api"]
