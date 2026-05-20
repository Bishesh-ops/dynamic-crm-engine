FROM golang:1.25.10-alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o crm-engine ./cmd/api/

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/crm-engine .
COPY --from=builder /app/cmd/api/migrations ./cmd/api/migrations

COPY --from=builder /app/web ./web

EXPOSE 8080

CMD ["./crm-engine"]
