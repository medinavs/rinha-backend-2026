FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v3 go build -ldflags="-s -w" -o app ./cmd/api

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/app /app/app
COPY --from=builder /app/resources/ /app/resources/
EXPOSE 9999
CMD ["/app/app"]
