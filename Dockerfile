# ---------- BUILD ----------
FROM golang:1.25.7 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /out/gotracker \
    ./cmd/api


# ---------- RUN ----------
FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /out/gotracker ./gotracker

EXPOSE 8080

CMD ["./gotracker"]