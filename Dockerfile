FROM golang:1.24 AS builder
LABEL authors="Pavel"

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /app/bin/server ./cmd/pr-service

FROM gcr.io/distroless/base-debian12

COPY --from=builder /app/bin/server ./server

EXPOSE 8080

ENTRYPOINT ["./server"]
