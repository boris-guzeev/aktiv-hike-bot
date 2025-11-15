FROM golang:1.25-alpine as build
WORKDIR /app

RUN apk add --no-cache git tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARC=amd64 go build -o /bot ./cmd/bot

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /bot /app/bot

ENTRYPOINT [ "/app/bot" ]