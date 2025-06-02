FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod ./
COPY go.su[m] ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api

FROM alpine:3.18

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/bin/api /app/api

COPY --from=builder /app/config/config.yaml /app/config/config.yaml

WORKDIR /app

EXPOSE 8000

CMD ["./api"]
