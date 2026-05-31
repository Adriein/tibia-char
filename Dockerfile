FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY app/go.mod app/go.sum ./

RUN go mod download

COPY app/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o tibia-char ./cmd/tibia-char/main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/tibia-char .

COPY --from=builder /app/ui ./ui

EXPOSE 4000

CMD ["./tibia-char"]
