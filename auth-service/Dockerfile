FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev librdkafka-dev pkgconf

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -tags dynamic -o auth-service .

FROM alpine:latest
RUN apk --no-cache add ca-certificates librdkafka

WORKDIR /root/

COPY --from=builder /app/auth-service .
COPY --from=builder /app/.env .

EXPOSE 8088

CMD ["./auth-service"]
