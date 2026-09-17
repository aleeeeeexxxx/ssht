FROM node:20-alpine AS web-builder

WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=web-builder /app/cmd/ssht/dist ./cmd/ssht/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o ssht ./cmd/ssht

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/ssht .

EXPOSE 6001

ENTRYPOINT ["./ssht"]
CMD ["-addr", ":6001"]
