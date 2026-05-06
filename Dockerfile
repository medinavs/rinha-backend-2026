FROM --platform=linux/amd64 golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v3 \
    go build -trimpath -ldflags='-s -w' -o /out/rinha-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v3 \
    go build -trimpath -ldflags='-s -w' -o /out/rinha-preprocess ./cmd/preprocess
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v3 \
    go build -trimpath -ldflags='-s -w' -o /out/rinha-healthcheck ./cmd/healthcheck

RUN mkdir -p /out/data \
    && /out/rinha-preprocess \
        -in /src/resources/references.json.gz \
        -out /out/data/index.ivf8192.bin \
        -clusters 8192 \
        -nprobe 8 \
        -ambiguous-nprobe 32 \
        -repair=true

FROM --platform=linux/amd64 alpine:3.19
WORKDIR /app
COPY --from=build /out/rinha-api /usr/local/bin/rinha-api
COPY --from=build /out/rinha-healthcheck /usr/local/bin/rinha-healthcheck
COPY --from=build /out/data/index.ivf8192.bin /app/resources/index.ivf8192.bin
COPY --from=build /src/resources/normalization.json /app/resources/normalization.json
COPY --from=build /src/resources/mcc_risk.json /app/resources/mcc_risk.json
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/rinha-api"]
