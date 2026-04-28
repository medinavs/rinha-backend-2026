FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .

RUN mkdir -p /app/resources-bin \
    && go run ./cmd/preprocess \
        -in /app/resources/references.json.gz \
        -out /app/resources-bin/references.bin

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v3 \
    go build -trimpath -ldflags="-s -w" -o /bin/app ./cmd/api

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /bin/app /app/app
COPY --from=builder /app/resources-bin/references.bin /app/resources/references.bin
COPY --from=builder /app/resources/normalization.json /app/resources/normalization.json
COPY --from=builder /app/resources/mcc_risk.json /app/resources/mcc_risk.json
EXPOSE 9999
CMD ["/app/app"]
