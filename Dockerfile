# syntax=docker/dockerfile:1

FROM golang:1.26.1-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/cleartify ./cmd/cleartify

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/cleartify /usr/local/bin/cleartify

ENV PORT=8080
ENV DATABASE_PATH=/data/cleartify.db

EXPOSE 8080
ENTRYPOINT ["cleartify"]
