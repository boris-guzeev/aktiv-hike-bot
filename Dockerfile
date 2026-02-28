FROM --platform=linux/amd64 golang:1.25-alpine as build
WORKDIR /app

RUN apk add --no-cache git tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# --- Build admin bot ---
FROM build AS build-admin
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/bot ./cmd/admin-bot

# --- Build client bot ---
FROM build AS build-client
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/bot ./cmd/client-bot && chmod 775 /out/bot

# --- Runtime admin bot ---
FROM gcr.io/distroless/base-debian12 AS admin
WORKDIR /app
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build-admin /out/bot /app/bot
ENTRYPOINT [ "/app/bot" ]

# --- Runtime client bot ---
FROM gcr.io/distroless/base-debian12 AS client
WORKDIR /app
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build-client /out/bot /app/bot
ENTRYPOINT [ "/app/bot" ]