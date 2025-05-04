# ---------- Build Stage ----------
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w -extldflags '-static'" -o main .

# ---------- Runtime Stage ----------
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main .

ENTRYPOINT ["./main"]
