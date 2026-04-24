# --- Build stage ---
FROM golang:alpine AS builder

WORKDIR /app

# Copy go.mod/go.sum từ root (single module)
COPY go.mod go.sum ./
RUN go mod download

# Copy toàn bộ source
COPY . .

# Build binary
RUN go build -o /bin/todos ./services/todos/cmd/main.go

# --- Runtime stage ---
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /bin/todos /bin/todos

EXPOSE 50051

ENTRYPOINT ["/bin/todos"]
