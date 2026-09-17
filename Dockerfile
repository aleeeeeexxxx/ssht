FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ssht ./cmd/ssht

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/ssht .

EXPOSE 6001

ENTRYPOINT ["./ssht"]
CMD ["-addr", ":6001"]
