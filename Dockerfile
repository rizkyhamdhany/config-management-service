# --- Build stage ---
FROM golang:1.21-alpine AS build
WORKDIR /

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./bin/api ./cmd

# --- Runtime stage (non-root) ---
FROM gcr.io/distroless/base-debian11:nonroot
WORKDIR /srv

# copy the exact file produced above; chown to nonroot for good measure
COPY --from=build /bin/api /srv/api
COPY --from=build db/migrations /srv/db/migrations

VOLUME ["/srv/data"]
ENV DATABASE_URL="file:/srv/data/configs.db?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL"
EXPOSE 8080
ENTRYPOINT ["/srv/api"]
