FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev pkgconfig librdkafka-dev

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -tags dynamic -o event-store .

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache librdkafka

COPY --from=builder /app/event-store .

EXPOSE 8090

CMD ["./event-store"]
